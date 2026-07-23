# Seats, Rollers, and Teleports

First page of the Decoration section. This section covers furniture whose whole purpose is what happens when a unit moves onto, off of, or through it: not the click-driven interactions covered in [[FURNITURE-INTERACTIONS]], but the pieces that act through the room's own movement and surface system. This page covers the three that change where or how a unit stands: seating, rollers, and teleports. [[DECORATION-WALL]] covers wall-mounted items; [[DECORATION-AMBIENCE]] covers room-wide mood and surface decoration.

## Sitting and laying aren't an interaction type at all

A chair doesn't have `InteractionType: "sit"`. Whether a unit can sit or lay on an item, and exactly where, comes entirely from its declared **slots**: geometry stored in the definition's `Metadata` (see [[FURNITURE-MODEL]]) and resolved fresh for every placed instance:

```go
// Slot describes one resolved sit/lay position for a placed furniture item.
type Slot struct {
	Point        grid.Point
	Z            grid.Height
	BodyRotation worldunit.Rotation // forced facing for a unit using the slot
	Status       SlotStatus         // "sit" or "lay"
}

// Slots derives the resolved sit/lay slots for a placed item.
func Slots(item Item) []Slot {
	top := item.Top()
	for _, declared := range item.Definition.Slots {
		dx, dy := rotateOffset(declared.DX, declared.DY, item.Rotation)
		point, _ := grid.NewPoint(int(item.Point.X)+dx, int(item.Point.Y)+dy)
		slots = append(slots, Slot{Point: point, Z: top,
			BodyRotation: rotateBody(declared.BodyRotation, item.Rotation), Status: declared.Status})
	}
	return slots
}
```

A slot is declared once in local coordinates (an offset from the item's own tile, a forced facing, sit or lay) and `Slots()` rotates both the offset and the facing by however the placed instance is actually turned. A chair's catalog definition only ever describes "the seat is one tile in front of the chair's back, facing forward"; rotating the chair rotates both automatically.

Those resolved slots feed straight into the surface resolver from [[ROOMS-HEIGHTMAP]] as ordinary sections, just with a terminal state instead of a plain open one:

```go
// slotState maps a slot status to its resolver section state.
func slotState(status SlotStatus) surface.State {
	if status == SlotStatusLay {
		return surface.StateLay
	}
	return surface.StateSit
}
```

And settling onto that section is where the forced facing actually takes effect. A unit that paths onto a sit or lay section gets its rotation overridden by the slot, not by whatever direction it was walking:

```go
case surface.StateSit:
	world.settleOnSection(playerID, roomUnit, position.Point, worldfurniture.SlotStatusSit, worldunit.StatusSit, section)
case surface.StateLay:
	world.settleOnSection(playerID, roomUnit, position.Point, worldfurniture.SlotStatusLay, worldunit.StatusLay, section)
...
roomUnit.Settle(unitStatus, heightValue(slot.Z-section.Z()), slot.BodyRotation, slot.BodyRotation)
```

The practical result: sitting down is pathfinding to a tile, not a special client request. A player clicking a chair sends the same room-move packet as walking anywhere else; the chair only changes what happens once they arrive. This is also why a chair with two seats can be occupied by two different players facing two different directions, and why laying furniture (beds, some rugs) works through the identical mechanism with a different `Status`. Nothing about pathfinding, broadcasting, or occupancy needs to know sitting and laying are different from walking; the surface section already carries the difference.

## Rollers: carrying units instead of seating them

A roller doesn't seat anyone either; see [[FURNITURE-ADVANCED]] for the full mechanism, cannon-scheduled batched moves included. What's worth restating here, next to seating, is the framing: a roller is "non-blocking" in the same sense a chair is. A unit standing on one isn't obstructed, it's *acted on*, just on the room's autonomous cycle (see [[ROOMS-ENTITIES]]) instead of on arrival. A unit can be sitting on a chair that itself sits on a roller, and both behaviors apply independently: the roller moves the whole stack, the chair still forces the sitting unit's facing once it settles again.

## Teleports: the other way to change where a unit stands

Teleports, covered fully in [[FURNITURE-ADVANCED]], are the third member of this family for the same reason: instead of a unit choosing its own destination through normal pathing, the paired-pad phase machine relocates it, across the room or across rooms entirely, while still going through the ordinary settle-onto-a-section step described above once it arrives. A teleport's destination tile still has to resolve to a real walkable section through the same surface system a normal step would use; nothing about arriving via teleport skips that check.

## What this page's three behaviors have in common

Seats, rollers, and teleports are the three ways a furniture item changes a unit's position or facing without the unit's own pathfinding decision being the only input. None of them are click interactions from [[FURNITURE-INTERACTIONS]]'s registry. A chair has no `Use` handler at all, a roller acts from the room cycle, and a teleport's click only starts an approach phase; it doesn't move anyone by itself. If a new piece of furniture needs to change where a unit stands or how it's oriented, this is the family it belongs to, not the click-dispatch system.
