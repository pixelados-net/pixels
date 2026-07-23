# Wall Items and Post-its

Second page of the Decoration section. Wall-mounted furniture (posters, frames, post-its) is a different placement model from floor furniture: no footprint, no rotation, no stacking, just a fixed spot on a wall plane. This page covers that generic path and the one wall item with real interactive state, the post-it.

## Wall placement is its own command path

Wall items don't go through the floor placement flow described in [[FURNITURE-MODEL]] and [[ROOMS-HEIGHTMAP]] at all. There's no footprint to validate against the surface resolver, because a wall item never occupies a grid tile. Placement instead validates a wall-coordinate string and, for most wall items, that's the entire behavior:

```go
func (handler Handler) placeWall(ctx context.Context, command Command, ...) error {
	if !furnituremodel.ValidWallPosition(command.WallPosition) || definition.InteractionType == "postit" {
		return nil // post-its place through their own dedicated flow instead
	}
	placed, err := handler.Furniture.Place(ctx, furnitureservice.PlaceParams{
		ItemID: item.ID, ActorPlayerID: player.ID(), RoomID: roomID,
		WallPosition: command.WallPosition, UniqueInteractionType: uniqueInteraction,
	})
	packet, err := outwalladd.Encode(outwalladd.Item{
		ID: placed.ID, SpriteID: definition.SpriteID, WallPosition: command.WallPosition,
		ExtraData: placed.ExtraData, UsagePolicy: 0, OwnerID: placed.OwnerPlayerID, OwnerName: player.Username(),
	})
	return broadcast.RoomPacket(...)
}
```

A poster or a picture frame is, mechanically, exactly this and nothing more: a `Kind: KindWall` definition with no special `InteractionType`, placed at a wall position, broadcast once as an add packet. There's no separate "picture" behavior to implement; the generic wall-add path *is* the picture behavior. What distinguishes one wall item from another is entirely catalog data (`SpriteID`, `Name`, `Description`) and, when relevant, the `ExtraData` a definition happens to carry.

## The one-per-room constraint pattern

A small number of wall interaction types are meant to exist once per room. A mood light is the concrete example (see [[DECORATION-AMBIENCE]]), and placement enforces that by scanning the room's current items before allowing a second one:

```go
if definition.InteractionType == "dimmer" {
	uniqueInteraction = definition.InteractionType
	exists, err := handler.roomHasInteraction(ctx, roomID, uniqueInteraction)
	if exists {
		return handler.sendBubbleAlert(ctx, command.Handler, "session.bubble.furniture.max_dimmers")
	}
}
```

`UniqueInteractionType` is passed down into the same `Place` call every other wall item uses. Uniqueness is enforced once, at the placement boundary, rather than duplicated into every interaction type that happens to need it. A future one-per-room wall item only needs to opt into this check by interaction type string, not implement its own scan.

## Post-its: the wall item with real state

Post-its are the one wall item family with actual player-editable content, gated by their own `interaction_type == "postit"` check and routed through a dedicated command path rather than the generic wall-add flow:

```go
func (handler Handler) handlePostIt(ctx context.Context, ...) error {
	if definition.InteractionType != "postit" {
		return err
	}
	switch command.Kind {
	case KindPostItPlace:
		return handler.placePostIt(ctx, player, active, roomID, item, definition, command)
	case KindPostItGet:
		packet, err := outitemdata.Encode(item.ID, postItData(item.ExtraData))
		if err != nil {
			return err
		}
		return command.Handler.Send(ctx, packet)
	case KindPostItSave, KindPostItSet:
		return handler.savePostIt(ctx, player, active, roomID, item, definition, command)
	}
}
```

A post-it's color and text live together in its `ExtraData` (see [[INVENTORY-FURNITURE]] for the wire-visible-state convention that field follows), defaulting to Nitro's classic yellow (`DefaultPostItColor = "FFFF33"`) the moment it's placed, so the client always has something valid to render even before the owner writes anything. Saved text passes through the same filtering chain as any other player-authored string in the game (global censor, then room-scoped word filters), consistent with how chat and room names are filtered elsewhere in the codebase.

A related, smaller wall check exists for post-it *boards*: some wall items are the pole a post-it board actually hangs from, checked directly by the post-it command rather than through the generic interaction registry (`definition.InteractionType == "sticky_pole"`), since it's a placement-time constraint owned entirely by the post-it feature, not a click behavior in its own right.
