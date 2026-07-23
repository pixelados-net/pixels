package runtime

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// processLocked executes every matching candidate with the room state locked.
func (engine *Engine) processLocked(ctx context.Context, loaded *state, event trigger.Event, now time.Time) (Trace, error) {
	candidates := loaded.byKind[event.Kind]
	if len(candidates) == 0 {
		return Trace{}, nil
	}
	started := time.Now()
	execution := execution{
		context: ctx, state: loaded, event: event, now: now,
		queue: make([]eventQueue, 0, len(candidates)), visited: make(map[configuration.Point]struct{}, len(candidates)),
		trace: Trace{ID: event.ID, Kind: event.Kind, StartedAt: started},
	}
	for _, node := range candidates {
		if engine.matcher.Match(node, event) {
			execution.queue = append(execution.queue, eventQueue{stack: loaded.generation.Stacks[node.Point], trigger: node})
		}
	}
	err := engine.run(&execution)
	execution.trace.Duration = time.Since(started)
	engine.recordTrace(execution.trace)
	loaded.appendTrace(execution.trace)
	return execution.trace, err
}

// processTriggerLocked executes one already-selected timer trigger.
func (engine *Engine) processTriggerLocked(ctx context.Context, loaded *state, node *configuration.Node, event trigger.Event, now time.Time) (Trace, error) {
	started := time.Now()
	execution := execution{
		context: ctx, state: loaded, event: event, now: now,
		queue: []eventQueue{{stack: loaded.generation.Stacks[node.Point], trigger: node}}, visited: make(map[configuration.Point]struct{}, 1),
		trace: Trace{ID: event.ID, Kind: event.Kind, StartedAt: started},
	}
	err := engine.run(&execution)
	execution.trace.Duration = time.Since(started)
	engine.recordTrace(execution.trace)
	loaded.appendTrace(execution.trace)
	return execution.trace, err
}

// run drains one breadth-first trace within configured budgets.
func (engine *Engine) run(execution *execution) error {
	var result error
	for len(execution.queue) > 0 {
		request := execution.queue[0]
		execution.queue = execution.queue[1:]
		if request.stack == nil || request.depth > engine.config.MaxCallDepth {
			continue
		}
		if _, visited := execution.visited[request.stack.Point]; visited {
			continue
		}
		if execution.trace.Stacks >= engine.config.MaxStacksPerTrace {
			execution.trace.BudgetExhausted = true
			break
		}
		execution.visited[request.stack.Point] = struct{}{}
		execution.trace.Stacks++
		engine.activateNode(execution.context, execution.event.RoomID, request.trigger)
		passed, err := engine.conditionsPass(execution, request.stack)
		result = joined(result, err)
		if err != nil {
			engine.metrics.stackResults[2].Add(1)
		} else if passed {
			engine.metrics.stackResults[0].Add(1)
		} else {
			engine.metrics.stackResults[1].Add(1)
		}
		if err != nil || !passed {
			continue
		}
		engine.activateNodes(execution.context, execution.event.RoomID, request.stack.Extras)
		effects := engine.selectEffects(execution.state, request.stack, execution.event.ID)
		for _, node := range effects {
			if execution.trace.Effects >= engine.config.MaxEffectsPerTrace {
				execution.trace.BudgetExhausted = true
				break
			}
			execution.trace.Effects++
			result = joined(result, engine.execute(execution, node, request.depth))
		}
	}
	return result
}

// conditionsPass applies whole-stack AND or OR semantics.
func (engine *Engine) conditionsPass(execution *execution, stack *configuration.Stack) (bool, error) {
	if len(stack.Conditions) == 0 {
		return true, nil
	}
	if engine.views == nil {
		return false, nil
	}
	view, found := engine.views.View(execution.event.RoomID)
	if !found {
		return false, nil
	}
	matched := false
	for _, node := range stack.Conditions {
		result, err := engine.conditions.Evaluate(node, condition.Context{Event: execution.event, Now: execution.now, ResetAt: execution.state.resetAt, Effects: stack.Effects}, view)
		if err != nil {
			return false, err
		}
		matched = matched || result.Pass
		if stack.Or && result.Pass {
			return true, nil
		}
		if !stack.Or && !result.Pass {
			return false, nil
		}
	}
	return matched, nil
}

// selectEffects applies random or unseen stack selectors.
func (engine *Engine) selectEffects(loaded *state, stack *configuration.Stack, eventID uint64) []*configuration.Node {
	if len(stack.Effects) <= 1 {
		return stack.Effects
	}
	if stack.Unseen {
		index := loaded.unseen[stack.Point] % len(stack.Effects)
		loaded.unseen[stack.Point] = index + 1
		return stack.Effects[index : index+1]
	}
	if stack.Random {
		index := randomIndex(eventID, stack.Point, len(stack.Effects))
		return stack.Effects[index : index+1]
	}
	return stack.Effects
}

// execute runs or schedules one effect and enqueues its directives.
func (engine *Engine) execute(execution *execution, node *configuration.Node, depth int) error {
	if node.Delay > 0 {
		if engine.scheduler == nil || execution.state.delayed >= engine.config.MaxDelayedPerRoom {
			return nil
		}
		generationID := execution.state.generation.ID
		event := execution.event
		execution.state.delayed++
		engine.metrics.delayedTasks.Add(1)
		if !engine.scheduler.Schedule(event.RoomID, generationID, node.Delay, func(now time.Time) {
			_ = engine.executeDelayed(context.Background(), event, node, generationID, now, depth)
		}) {
			execution.state.delayed--
			engine.metrics.delayedTasks.Add(-1)
		}
		return nil
	}
	result, err := engine.effects.Execute(execution.context, node, execution.event)
	if err != nil {
		return err
	}
	engine.recordEffect(result.Status)
	if result.Status == effect.Applied {
		engine.activateNode(execution.context, execution.event.RoomID, node)
	}
	engine.applyResult(execution, result, depth)
	return nil
}

// executeDelayed executes one effect only against the generation that scheduled it.
func (engine *Engine) executeDelayed(ctx context.Context, event trigger.Event, node *configuration.Node, generationID uint64, now time.Time, depth int) error {
	value, found := engine.rooms.Load(event.RoomID)
	if !found {
		return nil
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	if loaded.generation.ID != generationID {
		return nil
	}
	if loaded.delayed > 0 {
		loaded.delayed--
		engine.metrics.delayedTasks.Add(-1)
	}
	started := time.Now()
	execution := execution{context: ctx, state: loaded, event: event, now: now, visited: make(map[configuration.Point]struct{}), trace: Trace{ID: event.ID, Kind: event.Kind, StartedAt: started}}
	result, err := engine.effects.Execute(ctx, node, event)
	if err == nil {
		engine.recordEffect(result.Status)
		if result.Status == effect.Applied {
			engine.activateNode(ctx, event.RoomID, node)
		}
		execution.trace.Effects++
		engine.applyResult(&execution, result, depth)
		err = engine.run(&execution)
	}
	execution.trace.Duration = time.Since(started)
	engine.recordTrace(execution.trace)
	loaded.appendTrace(execution.trace)
	return err
}

// applyResult applies engine-owned effect directives.
func (engine *Engine) applyResult(execution *execution, result effect.Result, depth int) {
	if result.ResetTimers {
		execution.state.resetAt = execution.now
		execution.state.timers = buildTimers(execution.state.generation, execution.now)
	}
	for _, targetID := range result.CallTargets {
		target := execution.state.generation.Nodes[targetID]
		if target != nil {
			execution.queue = append(execution.queue, eventQueue{stack: execution.state.generation.Stacks[target.Point], depth: depth + 1})
		}
	}
	for _, derived := range result.Derived {
		if derived.RoomID == execution.event.RoomID && len(execution.queue) < engine.config.MaxEventsPerTrace {
			for _, candidate := range execution.state.byKind[derived.Kind] {
				if engine.matcher.Match(candidate, derived) {
					execution.queue = append(execution.queue, eventQueue{stack: execution.state.generation.Stacks[candidate.Point], trigger: candidate, depth: depth + 1})
				}
			}
		}
	}
}

// appendTrace appends one trace to the fixed ring.
func (loaded *state) appendTrace(trace Trace) {
	loaded.traces[loaded.traceNext] = trace
	loaded.traceNext = (loaded.traceNext + 1) % len(loaded.traces)
	if loaded.traceCount < len(loaded.traces) {
		loaded.traceCount++
	}
}
