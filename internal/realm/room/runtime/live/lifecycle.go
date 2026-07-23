package live

import (
	"context"
	"time"
)

// startLoop starts the room owner goroutine.
func (room *Room) startLoop(ctx context.Context, interval time.Duration, movementPublisher MovementPublisher, doorbellPublisher DoorbellPublisher, cyclePublisher CyclePublisher, doorbellTimeout time.Duration) {
	if movementPublisher == nil && doorbellPublisher == nil && cyclePublisher == nil {
		return
	}
	room.loop.Start(ctx, interval, func(loopCtx context.Context) {
		movements := room.Tick()
		if len(movements) > 0 && movementPublisher != nil {
			_ = movementPublisher(loopCtx, room, movements)
		}
		expired := room.SweepDoorbell(time.Now(), doorbellTimeout)
		if len(expired) > 0 && doorbellPublisher != nil {
			_ = doorbellPublisher(loopCtx, room, expired)
		}
		if cyclePublisher != nil {
			_ = cyclePublisher(loopCtx, room, time.Now())
		}
	})
}

// stopLoop stops the room owner goroutine.
func (room *Room) stopLoop() { room.loop.Stop() }
