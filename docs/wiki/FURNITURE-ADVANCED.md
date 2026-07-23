# Standalone Interaction Subsystems

Third page of the Furniture section. Some furniture behavior is too involved to live as a case in `essential`'s switch. It needs its own persistence, its own configuration, or work that spans multiple ticks. Each of these registers itself as an `External` handler (see [[FURNITURE-INTERACTIONS]]) and lives in its own package under `internal/realm/furniture/interactions`.

## Rollers

Rollers move whatever is standing on them, units and stacked items alike, one tile per cycle, in the direction they face. Unlike every other interaction on this page, a roller isn't click-triggered at all; it advances on the room's own cadence:

```go
func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, _ time.Time) error {
	rollers := active.FurnitureByInteraction("roller")
	if !active.AdvanceRollerCycle(len(rollers) > 0) {
		return nil
	}
	...
}
```

`AdvanceRollerCycle` is a per-room counter that only fires the actual step every few room ticks. Rollers move at a slower, deliberate cadence, not once per 500ms tick (see [[ROOMS-RUNTIME]] for the tick itself). Each cycle finds the tile in front of the roller, resolves its surface column, checks nothing already occupies the destination, and moves every eligible unit or item riding the roller onto it in one batch, so a roller carrying a whole cluster of stacked items moves them together rather than tearing the stack apart.

Position writes are batched rather than persisted per step: `enqueuePersistence` queues the moved item, and a background `runPersistence` loop flushes the batch, so a roller running continuously doesn't generate one database write per tile per item.

## Teleports

Paired teleport pads (`teleport` on the entry pad, `teleport_tile` role during transit) move a player between two placed teleporters, including across different rooms, through an explicit phase machine rather than an instant jump:

```go
const (
	PhaseResolving Phase = iota + 1 // reserve the player while the durable pair resolves
	PhaseApproach                    // wait for the unit to reach the source's front tile
	PhaseEnter                       // wait for the controlled step onto the source pad
	PhaseCross                       // wait for the source opening animation
	PhaseForward                     // keep the departure visible before cross-room navigation
	PhaseArrival                     // wait until a cross-room renderer can receive destination visuals
	PhaseExit                        // wait for the destination to open before walking out
	PhaseSettle                      // wait for the controlled exit path to finish
)
```

`Cycle` advances any transit whose phase deadline has elapsed, each phase driving one small piece of the animation (open the source, hide the unit, place them at the destination, open the destination, walk them off). The phase machine exists because a teleport that spans two different rooms can't be a single atomic move: the destination room might not even be active yet, so each step is a resumable, timed transition rather than one function call.

## Rentables

Rentable furniture (time-limited placements bought with a recurring or one-off charge) tracks its own active window:

```go
func (state State) ActiveAt(now time.Time) bool { ... }
func (state State) SecondsRemaining(now time.Time) int32 { ... }
```

Expiry is computed against the stored end time rather than ticked down by a background job, the same pattern room ads use (see [[NAVIGATOR-ROOM-ADS]]). A rental "expires" the instant its window passes, for every reader, without anything needing to notice in real time.

## Love locks

Love locks require confirmation from two players before they commit to their locked state. One player starts it, the other must independently confirm, and either side can cancel before both have:

```go
type Store interface {
	Start(ctx context.Context, itemID int64, playerID int64) (Pending, bool, error)
	Invite(ctx context.Context, itemID int64, playerID int64) (Pending, bool, error)
	Finish(ctx context.Context, itemID int64, playerID int64) (bool, error)
	Cancel(ctx context.Context, itemID int64, playerID int64) (bool, error)
}
```

The pending confirmation is itself durable. It survives a disconnect between the two clicks, since there's no guarantee both players are online in the same short window.

## Mystery boxes

Mystery boxes resolve to one prize from a weighted pool when opened, plus a separate "inscribe a trophy" flow for the decorative name plate some mystery-box prizes come with. Both are ordinary handlers registered against `essential`, not part of its click switch, because a mystery box's open is really a furniture *replacement* (box becomes prize) rather than a state change on the same item.

## Firework

Firework furniture models a charge-then-explode cycle with a room-scheduled recharge:

```go
func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error)
func (service *Service) rechargeItem(ctx context.Context, request essential.Request) error
```

Clicking a charged firework explodes it (broadcast the explosion state to the room) and starts a recharge timer scheduled through the room's task queue; clicking an uncharged one does nothing. The configuration for how long recharge takes lives beside the package's own `Config`, following the pattern in [[CONFIGURATION]].

## The custom stack helper

`stackheight` is not a furniture *interaction*. Nothing clicks it. It's a dedicated packet handler (`instack`) that lets a room's owner or a rights-holder pin an exact stack height override onto a placed item, for furniture whose natural stack height doesn't match what the room designer wants stacked on it:

```go
func normalizedHeight(height int32) (*int32, error)
```

The normalized value feeds directly into the surface resolver's column model described in [[ROOMS-HEIGHTMAP]]. This is the mechanism that lets a designer manually correct a stacking quirk without waiting for a definition-level fix.

## The furniture games bridge

`game` is a thin adapter, not a game engine itself:

```go
// Package game adapts room games to the furniture interaction boundary.
type Service struct{ games *roomgames.Service }

func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	return service.games.UseFurniture(...)
}
```

The real game logic, Battle Banzai's flood fill, Freeze's explosions, Football's ball physics, lives in the room realm's own games service, described in [[ARCHITECTURE]] and detailed in [[GAMES-OVERVIEW]]. This package exists purely so a game-tile click, which arrives through the same generic furniture-click path as everything else on this page, can hand off to that service without the furniture realm needing to know anything about how the games work.

## Decor items driven from elsewhere

A few interaction types aren't dispatched from the furniture realm at all. They're read directly by the commands that own the surrounding feature. Post-its check for `sticky_pole` (the wall pole a post-it board hangs from) directly in the post-it command, and mannequins check for `background_toner` directly in the mannequin command. These aren't part of `essential`'s registry because the behavior they gate isn't a click response, it's a placement-time constraint owned entirely by that feature's own command.

## Composing behavior

Nothing stops a definition from combining a generic behavior with a specialized one at different points in its lifecycle. A roller can carry a die; a switch can sit next to a teleporter. Because each interaction type is resolved independently per item rather than per room, there's no ordering dependency to manage: `essential.Use` and every package on this page each look only at the one item they were called with.
