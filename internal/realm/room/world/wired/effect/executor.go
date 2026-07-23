package effect

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Execute executes one of the thirty canonical effects.
func (executor *Executor) Execute(ctx context.Context, node *configuration.Node, event trigger.Event) (Result, error) {
	if node == nil {
		return Result{Status: Skipped}, nil
	}
	key := node.Descriptor.Key
	if operation, found := furnitureOperation(key); found {
		if executor.services.Furniture == nil {
			return Result{Status: Blocked}, nil
		}
		return executor.services.Furniture.ExecuteFurniture(ctx, operation, node, event)
	}
	if operation, found := avatarOperation(key); found {
		if executor.services.Avatar == nil {
			return Result{Status: Blocked}, nil
		}
		return executor.services.Avatar.ExecuteAvatar(ctx, operation, node, event)
	}
	if operation, found := botOperation(key); found {
		if executor.services.Bot == nil {
			return Result{Status: Blocked}, nil
		}
		return executor.services.Bot.ExecuteBot(ctx, operation, node, event)
	}
	if operation, found := gameOperation(key); found {
		if executor.services.Game == nil {
			return Result{Status: Blocked}, nil
		}
		return executor.services.Game.ExecuteGame(ctx, operation, node, event)
	}
	if operation, found := progressionOperation(key); found {
		if executor.services.Progression == nil {
			return Result{Status: Blocked}, nil
		}
		return executor.services.Progression.ExecuteProgression(ctx, operation, node, event)
	}
	switch key {
	case "wf_act_reset_timers":
		return Result{Status: Applied, ResetTimers: true}, nil
	case "wf_act_give_reward":
		if executor.services.Reward == nil {
			return Result{Status: Blocked}, nil
		}
		return executor.services.Reward.Claim(ctx, node, event)
	case "wf_act_call_stacks":
		result := Result{Status: Applied, CallTargets: make([]int64, len(node.Targets))}
		for index, target := range node.Targets {
			result.CallTargets[index] = target.ItemID
		}
		return result, nil
	default:
		return Result{Status: Blocked}, fmt.Errorf("unsupported WIRED effect %s", key)
	}
}
