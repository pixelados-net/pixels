# Entities and Synchronization

Fourth page of the Rooms section. [[ROOMS-RUNTIME]] covers the room's own goroutine and tick; this page covers who and what actually lives inside that tick: players, bots, and pets sharing one movement model, furniture as a separate but connected world, and the broadcast mechanism that keeps every connected client showing the same room.

## One unit model, three kinds

A player walking, a bot patrolling, and a pet wandering are all the same underlying thing to the room world: a `Unit`, with one shared position, pathfinding, and status model.

```go
type Kind uint8

const (
	KindPlayer Kind = iota + 1
	KindBot
	KindPet
)
```

Movement, rotation, and status assignment (walking, sitting, laying; see [[DECORATION-SEATING]]) are implemented exactly once, against `Unit`, regardless of kind. A bot doesn't get its own pathfinder and a pet doesn't get its own collision model; both ride the same `path` and `surface` packages a player does.

## Keeping three id spaces from colliding

Units are looked up by a single `int64` key in one room-scoped map, so the three kinds need disjoint keyspaces without any type-switch bookkeeping. Each domain picks its own partition:

- **Players** use their own durable `PlayerID` directly, always a small positive number.
- **Bots** use the negation of their id:

```go
// EntityKey returns the collision-free negative world key for one bot.
func EntityKey(botID int64) int64 {
	if botID < 0 {
		return botID
	}
	return -botID
}
```

- **Pets** use a large negative offset plus their id, far below any bot key:

```go
const entityBase int64 = -(1 << 62)

func EntityKey(petID int64) int64 { return entityBase + petID }
```

Three numeric ranges, one map, zero collisions by construction: a bot with id `12` and a pet with id `12` produce `-12` and `-4611686018427387892` respectively, nowhere near each other or near any real player id. Reading a unit back and telling what it is doesn't need a side table either; `Unit.Kind` carries that directly, and the wire projection separately encodes bots and pets with negative protocol user ids, matching the same "no accidental collision with a real player id" logic the client itself expects.

## Riding: one unit projected onto another

A player riding a pet doesn't get a second unit. The *rider's* position is computed as an offset over the *mount's* position, using a fixed avatar effect and a height offset:

```go
const (
	RidingEffectID     int32       = 77            // Nitro's mounted-rider avatar effect
	RidingHeightOffset grid.Height = grid.HeightScale // one full room unit above the pet
)
```

Mounting itself is an explicit, gated action (`equipment.Service.Mount`), not something the room infers from proximity: it requires a saddle (`HasSaddle`) and either ownership or the pet's `PublicRide` flag being set, and dismounting is the same call with the flag reversed. While mounted, the pet's own `Cycle` step (below) special-cases the rider and stops making autonomous movement decisions for that pet, since a ridden pet is being driven by the player, not by its own AI.

## The single extensible tick pipeline

[[ROOMS-RUNTIME]] showed the room loop calling `room.Tick()` and then draining due tasks. What it deferred is what actually happens after that: a room-owned `cyclePublishers` list, appended to independently by every domain that needs per-tick work, all invoked in sequence from the one loop the room already has:

```go
// AddCyclePublisher appends a domain cycle to the single owner loop used by every active room.
func (registry *Registry) AddCyclePublisher(publisher CyclePublisher) {
	registry.mutex.Lock()
	registry.cyclePublishers = append(registry.cyclePublishers, publisher)
	registry.mutex.Unlock()
}
```

Every domain with autonomous room behavior registers itself this way at startup, independently of the others:

```go
// furniture/module.go: rollers advance on every room tick
rooms.AddCyclePublisher(service.Cycle)

// bot/module.go: bot AI decisions
rooms.AddCyclePublisher(service.Cycle)

// pet/runtime_wiring.go: pet AI decisions
rooms.AddCyclePublisher(runtime.Cycle)

// furniture/interactions/teleport/transfer.go: teleport phase machine
runtime.AddCyclePublisher(service.Cycle)

// room/world/wired/wiring/runtime.go: WIRED trigger evaluation
rooms.AddCyclePublisher(func(ctx context.Context, active *roomlive.Room, now time.Time) error {
	return engine.Cycle(ctx, active.ID(), now)
})
```

None of these packages know about each other, and none of them own a goroutine or a timer of their own. They're all just callbacks invoked, in registration order, from the same 500ms tick described in [[ROOMS-RUNTIME]]. A room with no pets, no bots, and no active teleports still calls every one of these callbacks every tick; each one is responsible for cheaply doing nothing when it has no due work (the pet cycle's `now.Before(pet.nextDue)` check is a representative example). This is what lets the room stay a single goroutine no matter how many independent systems have a stake in it: the extensibility point is a slice of callbacks, not more concurrency.

Teardown mirrors this with `AddClosePublisher`. When a room deactivates, every registered domain gets one chance to release whatever room-scoped state it was holding, all from one call:

```go
rooms.AddClosePublisher(func(roomID int64) {
	engine.Close(roomID)
	games.Close(roomID)
	groups.CloseRoom(roomID)
})
```

WIRED, the games engine (see [[GAMES-OVERVIEW]]), and social-group room state all release cleanly this way, the moment nobody's left to observe them, without the room realm needing to know what any of them are.

## Furniture is not a unit

Placed furniture lives in a separate model, the grid and surface system described in [[ROOMS-HEIGHTMAP]], not the unit map. The two systems interact at defined seams, not by furniture becoming a kind of unit:

- A unit stepping onto or off a furniture item's footprint fires `walkedon`/`walkedoff` events, which is what drives pressure plates, color plates, and hand-item tiles (see [[FURNITURE-INTERACTIONS]]).
- A unit settling onto a sit or lay section (declared by the item's slots) gets a forced body rotation and status assigned by the surface resolver itself, covered in [[DECORATION-SEATING]].
- Placing, moving, or picking up furniture changes the surface columns at its footprint, which is why furniture mutations broadcast a heightmap update in addition to the add/move/remove packet: every connected client's local pathing cache has to be told the walkable surface actually changed, not just that an object moved.

## How synchronization actually works

There's no separate "sync" step and no snapshot diffing. The active `*Room` in memory is the single authoritative copy of everything covered on this page and on [[ROOMS-HEIGHTMAP]]. Every read a handler or a cycle callback does goes against that same struct, guarded by the room's own lock, and every write is visible to the very next read. What travels to clients is a one-way stream of packets describing what just changed, sent best-effort to whoever is currently present:

```go
// RoomPacket sends a packet to active room occupants. Delivery is best-effort: a failed send to
// one occupant (typically a connection mid-disconnect) never fails the caller, because the failing
// connection's own lifecycle handles its cleanup and a command must not disconnect the acting
// player just because a bystander's socket died.
func RoomPacket(ctx context.Context, connections *netconn.Registry, active *live.Room, packet codec.Packet, excludedPlayerID int64) error {
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID == excludedPlayerID {
			continue
		}
		connection, found := connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			continue
		}
		_ = connection.Send(ctx, packet)
	}
	return nil
}
```

`excludedPlayerID` is how an action's own actor avoids getting an echo of a broadcast they already have a direct reply for. It isn't a correctness mechanism, just a redundant-packet skip.

`Occupants()` and `Units()` answer different questions and it matters which one a piece of code asks. `Occupants()` is *network presence*: everyone who should receive room broadcasts, including an overflow visitor parked as a spectator with no position in the world (`Spectator bool // reports an invisible overflow occupant without a world unit`). `Units()` is *world presence*: only entities with real position, movement, and status, a strict subset of occupants. Broadcasting always goes through occupants; movement, pathing, and stacking always go through units.

A newly entered player is brought up to date once, at entry (`RoomSpawn`: current units, their statuses, their persistent actions like an ongoing dance), and after that, they receive exactly the same incremental broadcasts everyone already in the room receives. There's no separate "late joiner" code path once the catch-up packet is sent. That's the same pattern [[INVENTORY-FURNITURE]] uses for inventory lists: a full state dump on demand, incremental deltas for everything after. A room is "in sync" precisely because there's one writer (the room's own goroutine, per [[ROOMS-RUNTIME]]) and every reader is just a fan-out target of that writer's packet stream. There's nothing to reconcile because nothing but the room itself ever mutates room state.
