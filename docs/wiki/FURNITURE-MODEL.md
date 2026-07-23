# The Furniture Definition

First page of the Furniture section. Every placeable object in the game, from a chair to a teleporter pad, is one row in the furniture catalog: a `Definition`. This page covers what that row contains and what each field controls. [[FURNITURE-INTERACTIONS]] covers how definitions drive behavior at runtime; [[FURNITURE-ADVANCED]] covers the standalone subsystems built on top of it.

## Floor and wall, nothing else

```go
// Kind describes a furniture definition category.
type Kind string

const (
	// KindFloor marks a floor-placed furniture definition.
	KindFloor Kind = "floor"

	// KindWall marks a wall-placed furniture definition.
	KindWall Kind = "wall"
)
```

There is no third placement kind. Wall items (posters, wall lamps, art) carry wall position data instead of grid coordinates when placed; everything else is a floor item with a footprint. The rest of the `Definition` struct is shared between both:

```go
type Definition struct {
	sharedmodel.Base

	SpriteID    int    // Nitro rendering class id
	Name        string // stable technical identifier, e.g. chair_plasto
	PublicName  string
	Description string
	Kind        Kind

	Width       int
	Length      int
	StackHeight float64

	AllowStack           bool
	AllowWalk            bool
	AllowSit             bool
	AllowLay             bool
	AllowInventoryStack  bool
	AllowTrade           bool
	AllowMarketplaceSale bool
	AllowRecycle         bool

	RedeemableCredits int32

	EffectPool   []int32
	EffectMale   *int32
	EffectFemale *int32

	InteractionType       string
	InteractionModesCount int
	Multiheight           string
	CustomParams          string

	Metadata json.RawMessage
}
```

## Footprint and stacking

`Width` and `Length` describe the tile footprint at rotation zero; the room world rotates it on demand rather than storing four pre-rotated shapes (`worldfurniture.Footprint` and `worldfurniture.Dimensions`, used throughout the interaction packages, do this rotation). `StackHeight` is how much height in room units the item adds for anything placed on top of it. A chair contributes its seat height, a rug contributes almost nothing. Whether anything is allowed to use that added height at all is a separate flag, `AllowStack`; a `StackHeight` on a non-stackable item is inert. The runtime consequence of both, how the surface resolver turns a footprint and a stack height into an actual walkable section, is [[ROOMS-HEIGHTMAP]]'s subject, not this page's.

`AllowWalk`, `AllowSit`, and `AllowLay` are independent: a chair typically allows sit but not walk or lay, a rug typically allows walk but not sit, and a bed allows lay. All three can coexist on one definition when the client's own animation set supports it.

## The capability flags

`AllowInventoryStack`, `AllowTrade`, `AllowMarketplaceSale`, and `AllowRecycle` gate the four systems covered on [[INVENTORY-FURNITURE]] and its neighboring pages: whether identical copies collapse into one inventory row, whether direct player trading will accept the item, whether the marketplace will list it, and whether the recycler will consume it for prize odds. These are catalog decisions, not code paths. A new limited-edition release that shouldn't be recyclable is a row edit, not a deploy.

`RedeemableCredits` is the credit value paid out when the item is consumed through a redemption exchange (see [[INVENTORY-WALLET]] for the atomic charge-and-deliver pattern that pays it out).

## Effects

`EffectPool`, `EffectMale`, and `EffectFemale` feed two different interaction types covered in detail on [[FURNITURE-INTERACTIONS]]: an `effect_giver` definition picks randomly from `EffectPool`, while an `effect_tile` definition grants exactly `EffectMale` or `EffectFemale` depending on the walking player's gender. Both ultimately call the same player-realm grant described in [[USERS-PROFILE]]. The furniture layer decides *which* effect; the player layer owns *storing* it.

## The interaction configuration trio

Three fields exist purely to configure the interaction system:

- **`InteractionType`** is the string that selects behavior: `"dice"`, `"roller"`, `"vendingmachine"`, and so on. This is the field [[FURNITURE-INTERACTIONS]] is entirely about.
- **`InteractionModesCount`** bounds how many discrete states a multi-state item cycles through: how many faces a die has, how many colors a color wheel offers.
- **`CustomParams`** and **`Multiheight`** carry free-form, interaction-specific configuration that doesn't deserve its own column. `CustomParams` uses an `Arcturus`-compatible `key=value,key=value` grammar. The states/delay pair parsed for non-dice random-state items is a working example:

```go
// parseRandomParams parses Arcturus-compatible states and delay parameters.
func parseRandomParams(value string) (int, time.Duration) {
	states := 0
	delay := time.Duration(0)
	for _, pair := range strings.Split(value, ",") {
		key, raw, found := strings.Cut(pair, "=")
		...
		switch strings.TrimSpace(key) {
		case "states":
			states = number
		case "delay":
			delay = time.Duration(number) * time.Millisecond
		}
	}
	return states, delay
}
```

Reusing the reference grammar here isn't nostalgia. Catalog data authored against classic tools drops in unchanged.

## Metadata is the escape hatch

`Metadata` is a raw JSON column the client never sees, for structured configuration that doesn't fit a scalar field. Sit and lay slot geometry for oddly-shaped seating is the recurring example. It exists so that adding one unusual definition doesn't force a schema migration; if a shape of data starts recurring across many definitions, that's the signal to promote it out of `Metadata` into a real column instead.

## Where a definition ends and an item begins

`Definition` is catalog data: one row per furniture *type*, shared by every copy in the game. The thing a player actually owns or places is an *item*: a row referencing a definition plus placement, rotation, `ExtraData` (the wire-visible state a dice roll or a post-it's text lives in), and ownership. That split, and the single-table inventory model built on it, is [[INVENTORY-FURNITURE]]'s subject.
