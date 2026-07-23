package runtime

import (
	"container/heap"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// timerEntry stores one compiled timer deadline.
type timerEntry struct {
	// node stores the timer trigger.
	node *configuration.Node
	// deadline stores the next deadline.
	deadline time.Time
	// period stores repeat cadence or zero for one-shot.
	period time.Duration
	// kind stores the emitted event kind.
	kind trigger.Kind
}

// timerQueue orders timer entries by deadline and item id.
type timerQueue []timerEntry

// Len returns timer count.
func (queue timerQueue) Len() int { return len(queue) }

// Less orders deadlines and stabilizes equal deadlines by item id.
func (queue timerQueue) Less(left int, right int) bool {
	if queue[left].deadline.Equal(queue[right].deadline) {
		return queue[left].node.ItemID < queue[right].node.ItemID
	}
	return queue[left].deadline.Before(queue[right].deadline)
}

// Swap exchanges timer entries.
func (queue timerQueue) Swap(left int, right int) {
	queue[left], queue[right] = queue[right], queue[left]
}

// Push appends one timer entry.
func (queue *timerQueue) Push(value any) { *queue = append(*queue, value.(timerEntry)) }

// Pop removes the last heap entry.
func (queue *timerQueue) Pop() any {
	previous := *queue
	last := len(previous) - 1
	value := previous[last]
	*queue = previous[:last]
	return value
}

// buildTimers creates a deadline heap from compiled trigger nodes.
func buildTimers(generation *configuration.Generation, now time.Time) timerQueue {
	queue := make(timerQueue, 0)
	for _, node := range generation.Triggers {
		kind, periodic := timerKind(node.Descriptor.Key)
		if kind == 0 || node.Parameters.Duration <= 0 {
			continue
		}
		period := time.Duration(0)
		if periodic {
			period = node.Parameters.Duration
		}
		queue = append(queue, timerEntry{node: node, deadline: now.Add(node.Parameters.Duration), period: period, kind: kind})
	}
	heap.Init(&queue)
	return queue
}

// timerKind maps timer descriptors to event kind and repeat policy.
func timerKind(key string) (trigger.Kind, bool) {
	switch key {
	case "wf_trg_periodically":
		return trigger.Periodic, true
	case "wf_trg_period_long":
		return trigger.PeriodicLong, true
	case "wf_trg_at_given_time":
		return trigger.AtTime, false
	case "wf_trg_at_time_long":
		return trigger.AtTimeLong, false
	default:
		return 0, false
	}
}
