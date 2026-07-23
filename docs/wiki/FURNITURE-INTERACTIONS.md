# Interaction Types

Second page of the Furniture section. `InteractionType` (see [[FURNITURE-MODEL]]) is a string, not an enum with generated dispatch. Every value is a case in a small number of switches, and this page walks through what each one does. There are two dispatch layers: a generic one for behavior every furniture item might need, and a specialized one for behavior only certain items need.

## Layer one: generic state behavior

`internal/realm/furniture/interactions` resolves the two interaction types that don't need any specialized logic. They just cycle or flip a state string:

```go
// Resolve returns behavior for one definition interaction type.
func (registry *Registry) Resolve(interactionType string) (Behavior, bool) {
	switch interactionType {
	case "default", "toggle":
		return registry.toggle, true
	case "gate":
		return registry.gate, true
	default:
		return nil, false
	}
}
```

`toggle` cycles ordinary multi-state furniture (a lamp with several color states) through its `InteractionModesCount` states in order, wrapping back to zero. `gate` is a binary behavior: it flips between two states and changes whether the tile is walkable, rejecting the flip if a unit currently occupies the footprint. That occupancy check is what stops a gate from trapping someone under it.

Anything not `default`, `toggle`, or `gate` falls through to layer two.

## Layer two: specialized interactions

The `essential` package owns everything else, dispatched from one switch in `Service.Use`:

```go
switch request.Item.Definition.InteractionType {
case "dice", "colorwheel", "random_state":
	return true, service.useRandom(ctx, request)
case "pressureplate", "colorplate", "handitem_tile":
	return true, nil // handled by occupancy events instead of clicks, see below
case "onewaygate", "switch", "switch_remote_control", "multiheight":
	return true, service.useTraversal(ctx, request)
case "vendingmachine", "vendingmachine_no_sides", "handitem":
	return true, service.useHandItem(ctx, request)
case "cannon":
	return true, service.useCannon(ctx, request)
case "effect_giver":
	return true, service.useEffectGiver(ctx, request)
default:
	for _, handler := range service.external {
		if handled, err := handler.UseFurniture(ctx, request); handled || err != nil {
			return handled, err
		}
	}
	return false, nil
}
```

That `default` branch is how [[FURNITURE-ADVANCED]]'s standalone packages (roller, teleport, rentable, lovelock, mystery box, firework, the room-games bridge) plug in without `essential` knowing they exist. Each registers itself as an `External` handler at startup via `AddExternal`, and unclaimed clicks fall through the chain until one claims them.

### Random-state family: dice, colorwheel, random_state

All three resolve to the same delayed-roll mechanic in `useRandom`, differing only in how many states they roll across and how long the roll takes:

- **`dice`** requires the clicking player be standing on an adjacent tile (dice are meant to be rolled by someone actually at the table), rolls across `InteractionModesCount` faces, and settles after 1.5 seconds.
- **`colorwheel`** requires room-furniture management rights rather than adjacency, only someone who can edit the room can spin it, and takes 3 seconds.
- **`random_state`** is the generic case: state count and delay come from `CustomParams` (`states=6,delay=1000`), so a catalog item can define its own random behavior without a bespoke interaction type.

While rolling, the item's `ExtraData` is set to Nitro's rolling sentinel value; a second click while already rolling is a no-op. The actual roll is scheduled on the room's own task queue rather than blocking the click:

```go
key := scheduledKey(request.Item.ID, 1)
request.Room.ScheduleReplacing(key, delay, func(time.Time) {
	result := service.random.IntN(modes) + 1
	...
})
```

`ScheduleReplacing` is the room runtime's delayed-work primitive, covered in [[ROOMS-RUNTIME]]. Nothing in the interaction packages spawns its own goroutine or timer; every delay is a task scheduled against the owning room. A dice value can also be reset early through Nitro's dedicated close-dice request (`CloseDice`), which the client sends when a player picks the dice back up.

### Occupancy family: pressureplate, colorplate, handitem_tile, effect_tile

These four don't respond to clicks at all. `useRandom`'s cousins here are `nil`-handled in the click switch because the real trigger is a unit stepping on or off the tile, delivered as `furniture.walkedon` / `furniture.walkedoff` events:

```go
switch item.Definition.InteractionType {
case "pressureplate":
	service.schedulePressure(ctx, active, item)
case "colorplate":
	return service.changeColorPlate(ctx, active, item, 1) // -1 on walk-off
case "handitem_tile":
	return service.giveHandItem(ctx, active, payload.PlayerID, item, false)
case "effect_tile":
	return service.giveTileEffect(ctx, payload.PlayerID, item)
}
```

- **`pressureplate`** debounces occupancy through a short replacing task (100ms) so a player briefly crossing the tile doesn't cause a visible flicker, then sets state to `"1"` or `"0"` based on whether anyone is still standing in the footprint.
- **`colorplate`** increments or decrements a bounded counter (clamped to `[0, InteractionModesCount-1]`) each time someone steps on or off. A crowd of people on a colorplate visibly deepens its color.
- **`handitem_tile`** hands the walking player a hand item on entry (see the vending family below for what "give a hand item" means) without requiring a click.
- **`effect_tile`** grants and immediately enables a gender-specific avatar effect from `EffectMale`/`EffectFemale`, routed through the player realm's effect grant described in [[USERS-PROFILE]], and expires after a fixed duration (24 hours) if not consumed sooner.

### Traversal family: onewaygate, switch, switch_remote_control, multiheight

Grouped because all four change how units move through the room, and three of the four involve the room's controlled-walk mechanism: pathing a player to a specific tile before the interaction actually fires.

- **`switch`** toggles immediately if the clicking player is already adjacent; otherwise it walks them to the nearest activator tile first and toggles on arrival, polled once per tick until the walk settles:

```go
func (service *Service) useSwitch(ctx context.Context, request Request, remote bool) error {
	...
	if adjacentToItem(unit.Position.Point, request.Item) {
		return service.toggleFinal(ctx, request)
	}
	// otherwise: sort activator tiles by distance, controlled-walk to the nearest, toggle on arrival
```

- **`switch_remote_control`** is the same toggle without the walk. It requires room-management rights instead of proximity, for a switch meant to be operated from anywhere in the room.
- **`onewaygate`** only opens from its front tile and only when the back tile is clear, then briefly sets state to `"1"` to let exactly one crossing through: the directional gate you can walk through one way but not back.
- **`multiheight`** is `useTraversal`'s default case. It doesn't gate or switch anything by itself; it's the interaction type used together with a definition's `Multiheight` configuration (see [[FURNITURE-MODEL]]) to let a footprint expose more than one walkable height, which the surface resolver described in [[ROOMS-HEIGHTMAP]] turns into real sections.

### Hand-item family: vendingmachine, vendingmachine_no_sides, handitem

- **`handitem`** gives the clicking player a hand item directly if they're already standing on or adjacent to it, no walk needed.
- **`vendingmachine`** and **`vendingmachine_no_sides`** require the player to be standing on one of the item's designated activator tiles; if they're not, the same controlled-walk-then-activate pattern used by `switch` kicks in. The `_no_sides` variant simply narrows which tiles count as activators for furniture shaped so its side tiles shouldn't trigger it.

### cannon

An environmental "kick" interaction: clicking near a cannon walks the player to an adjacent activator tile if needed, then after a short lit-fuse delay (750ms, with a 2-second cooldown between shots) fires and forcibly moves whoever is in front of it. This is the mechanic behind furniture that launches other units.

### effect_giver

Picks one effect at random from `EffectPool` and grants it to the clicking player through the same effect-grant path as `effect_tile`, the click-triggered counterpart to the walk-triggered version.

## Writing a new interaction type

If a new definition needs behavior beyond cycling a state, the decision tree following from this page is: does an existing family already cover the shape of the behavior (occupancy-driven vs click-driven vs walk-then-activate)? If yes, extend that family's switch case. If the behavior is genuinely standalone, needing its own persistence, its own config, its own background work, it belongs as an `External` handler following the pattern in [[FURNITURE-ADVANCED]], not as a new case bolted onto `essential`.
