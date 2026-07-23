# Room Surfaces and the Mood Light

Third page of the Decoration section. Two features change how the whole room looks rather than any one tile: paint (floor, wallpaper, and window view) and the mood light. Both live in `internal/realm/room/decoration`, and both work the same way underneath: a consumable furniture item, owned by whoever manages the room, mutates one piece of durable room appearance and broadcasts the result to everyone present.

## Paint: consuming an item to repaint a plane

Floor, wallpaper, and window-view ("landscape") are the three paintable room planes:

```go
const (
	SurfaceFloor     Surface = "floor"
	SurfaceWallpaper Surface = "wallpaper"
	SurfaceLandscape Surface = "landscape"
)

type Appearance struct {
	Floor     string
	Wallpaper string
	Landscape string
}
```

A paint item is ordinary inventory furniture whose catalog **name** doubles as the surface it paints. `definition.Name` is read directly as the `Surface` value at the moment it's used, so there's no separate "which surface does this item affect" column to keep in sync with the catalog:

```go
func (handler Handler) handleSurface(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	...
	surface := roomdecor.Surface(definition.Name)
	if err = handler.Decoration.ApplySurface(ctx, item.ID, player.ID(), roomID, surface, item.ExtraData); err != nil {
		...
	}
	...
}
```

`ApplySurface` validates the surface is one of the three known planes and that the value matches Nitro's dot-delimited numeric pattern (a bare pattern like `101`, or a compound one like `101.2`) before consuming the item, so malformed input never reaches the database. Applying paint is destructive to the item, which is why the same handler that broadcasts the new look also removes it from the actor's inventory and asks the client to refresh:

```go
packet, err := outpaint.Encode(string(surface), item.ExtraData)
...
if err = handler.broadcast(ctx, active, packet); err != nil { return err }

packet, err = outremove.Encode(item.ID)      // the paint can is consumed
...
packet, err = outrefresh.Encode()             // tell the client its inventory changed
```

Unlike ordinary furniture consumption elsewhere in the codebase, this doesn't go through the atomic charge-and-deliver transaction pattern from [[INVENTORY-WALLET]]. There's nothing being delivered in exchange; the item's entire purpose is being spent on this one mutation. `Appearance` is stored per room and re-sent to anyone who enters afterward, so a repainted room stays repainted regardless of who's currently inside when it happens.

## The mood light: presets, not one live color

The dimmer holds three saved presets, not a single live color value:

```go
type Preset struct {
	ID             int32  // one-based preset slot (1-3)
	BackgroundOnly bool   // tint only the background, or the whole scene
	Color          string // validated uppercase hex
	Brightness     int32
	Selected       bool   // the currently active preset
}

const (
	MinimumBrightness int32 = 76
	MaximumBrightness int32 = 255
)
```

A player with room rights can edit any of the three slots and optionally apply it immediately (`Apply` on the save command), or just toggle the mood light on and off without changing which preset is selected:

```go
func (handler Handler) handleDimmer(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	if command.Kind == KindDimmerSettings {
		return sendDimmerSettings(ctx, command, state) // read-only: anyone can open the panel
	}
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil || !allowed {
		return err // editing requires room rights
	}
	if command.Kind == KindDimmerSave {
		state, err = handler.Decoration.SaveDimmer(ctx, roomID, ownerID, roomdecor.Preset{...}, command.Apply)
	} else {
		state, err = handler.Decoration.ToggleDimmer(ctx, roomID, ownerID)
	}
	...
}
```

Reading the current three presets is open to anyone in the room (it's how the client shows the preset picker at all), while saving or toggling requires the same room-management check used throughout the furniture and decoration commands (see [[USERS-PERMISSIONS]] for hotel permission resolution). Because a room can only ever hold one dimmer item, enforced at placement time (see [[DECORATION-WALL]]), "the room's mood light state" and "this one wall item's state" are the same durable record; there's no ambiguity about which dimmer a save applies to.

Applying a change re-encodes the dimmer's own wall item as an update and broadcasts it exactly like any other wall item mutation:

```go
packet, err := outwallupdate.Encode(item.ID, definition.SpriteID, *item.WallPosition, state.ExtraData, 0, item.OwnerPlayerID)
...
return handler.broadcast(ctx, active, packet)
```

The mood light doesn't need its own room-wide "ambient state" packet distinct from ordinary furniture updates. From the wire's perspective, changing the room's lighting is just that one wall item's `ExtraData` changing, the same mechanism [[DECORATION-WALL]] uses for everything else on that page.
