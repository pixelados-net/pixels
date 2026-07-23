# The Room Runtime

First page of the Rooms section. When a room has nobody in it, it's just a database row. The moment a first player enters, the room realm activates a live instance of it (`internal/realm/room/runtime/live`), and that instance owns a goroutine of its own for as long as anyone stays inside. This page covers that model: how many goroutines a room costs, what runs on its tick, and how delayed work gets scheduled without spawning anything extra. [[ROOMS-ENTRY]] covers how a player gets into that instance in the first place; [[ROOMS-HEIGHTMAP]] covers the 2.5D world the tick moves units through.

## One goroutine per active room, not one global loop and not one per player

`Registry.Activate` creates a `Room` and starts its loop the moment the room becomes active:

```go
func (registry *Registry) Activate(snapshot Snapshot) (*Room, error) {
	room, err := NewRoom(snapshot)
	...
	room.startLoop(context.Background(), registry.tickInterval, registry.movementPublish, registry.doorbellPublish, cycles, registry.doorbellTimeout)
	registry.rooms[snapshot.ID] = room
	return room, nil
}
```

`startLoop` hands off to a small owned-loop primitive that spawns exactly one goroutine per room and ticks it on an interval:

```go
func (loop *Loop) Start(ctx context.Context, interval time.Duration, tick Tick) {
	...
	go run(loopCtx, interval, tick, done)
}

func run(ctx context.Context, interval time.Duration, tick Tick, done chan<- struct{}) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick(ctx)
		}
	}
}
```

That's the entire concurrency model for room simulation: an active room is a `time.Ticker` and a goroutine, nothing more exotic. There's no global tick shared by every room (a spike in one room's occupancy can't slow another room's movement) and no per-entity goroutine or timer for individual units or furniture items (a room with five hundred pieces of furniture still costs exactly one goroutine). `stopLoop` cancels the context and blocks on the `done` channel until the goroutine actually exits, so deactivating a room is synchronous: by the time `Activate`'s counterpart returns, the room's goroutine is provably gone, not just asked to stop.

## What one tick does

`DefaultTickInterval` is 500 milliseconds. Each tick, `startLoop`'s callback does three things in order:

```go
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
```

1. **`room.Tick()`** advances the world simulation, one step of pathfinding movement for every walking unit, and returns whatever changed, so only rooms with actual movement pay for a broadcast that tick.
2. **`SweepDoorbell`** expires any doorbell request that's waited too long without an answer (see [[ROOMS-ENTRY]] for the doorbell queue itself).
3. **`cyclePublisher`** fans out to every domain that registered a per-tick hook. This is how the roller cycle described in [[FURNITURE-ADVANCED]] and the teleport phase machine advance, without either of those packages needing their own ticker. They're called from the one loop the room already has.

## Scheduling delayed work without extra goroutines

Almost everything on the [[FURNITURE-INTERACTIONS]] and [[FURNITURE-ADVANCED]] pages needs to do something *later*: a dice settling after 1.5 seconds, a switch toggling once a controlled walk finishes, a firework recharging. None of that spawns a `time.AfterFunc` or its own goroutine; it's appended to the room's own task queue and drained on the next tick:

```go
// Tick advances room world movement once.
func (room *Room) Tick() []Movement {
	...
	movements := room.world.Tick()
	room.mutex.Unlock()
	room.runDueTasks(time.Now())
	return movements
}

// ScheduleReplacing queues work after replacing the same non-zero key.
func (room *Room) ScheduleReplacing(key roomtask.Key, after time.Duration, run func(time.Time)) {
	room.tasks.Replace(key, time.Now().Add(after), run)
}
```

The queue (`internal/realm/room/runtime/live/task`) is deliberately simple: a mutex-protected slice, no heap, because a room rarely has more than a handful of pending tasks at once. Two entry points cover every case interaction code needs:

- **`Schedule`** appends independent work, fire-and-forget, never cancelled.
- **`ScheduleReplacing`** takes a `Key` and replaces any existing task under that key instead of stacking a duplicate. This is what makes a rapid double-click on a dice or a switch safe: the second click's scheduled resolution simply overwrites the first's rather than both eventually firing. Every specialized interaction on [[FURNITURE-INTERACTIONS]] builds its key from the item's id plus a small per-interaction discriminator (`scheduledKey(itemID, kind)`), so different interaction types on the same item never collide.

Because `runDueTasks` runs synchronously as part of `Tick()`, on the same goroutine, scheduled callbacks never race the room's own state. A task that reads or mutates room world state during its callback is running exactly where a normal tick step would, no extra locking required beyond what the room already does internally.

## Deactivation

A room's live instance is released once its last occupant leaves. `stopLoop` cancels the goroutine, and the registry drops the room from its active map. Nothing about a live room is itself durable; it's a cache of active simulation state rebuilt from the database record and current furniture placements the next time someone enters. If the process restarts, no active room survives it, which is consistent with the durable-first rule described in [[USERS-MODEL]]: live state is disposable, the database is not.
