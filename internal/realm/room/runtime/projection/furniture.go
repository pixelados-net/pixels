package projection

import (
	"strconv"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	outflooritems "github.com/niflaot/pixels/networking/outbound/room/furniture/flooritems"
	outwallitems "github.com/niflaot/pixels/networking/outbound/room/furniture/wallitems"
)

// GiftSender stores visible gift sender identity for present tags.
type GiftSender struct {
	// Name stores the visible sender name.
	Name string
	// Figure stores the visible sender figure.
	Figure string
}

// FloorItems maps placed furniture items into ROOM_FLOOR_ITEMS records.
func FloorItems(items []furnituremodel.Item, definitions map[int64]furnituremodel.Definition, ownerNames map[int64]string, giftSenders map[int64]GiftSender) ([]outflooritems.Owner, []outflooritems.FloorItem) {
	owners := ownerRecords(items, ownerNames)
	records := make([]outflooritems.FloorItem, 0, len(items))
	for _, item := range items {
		definition, ok := definitions[item.DefinitionID]
		if !ok || definition.Kind == furnituremodel.KindWall || item.X == nil || item.Y == nil || item.Z == nil {
			continue
		}
		records = append(records, floorItemRecord(item, definition, giftSenders))
	}

	return owners, records
}

// WallItems maps placed wall furniture into Nitro owner and item records.
func WallItems(items []furnituremodel.Item, definitions map[int64]furnituremodel.Definition, ownerNames map[int64]string) ([]outwallitems.Owner, []outwallitems.Item) {
	seen := make(map[int64]struct{}, len(items))
	owners := make([]outwallitems.Owner, 0, len(items))
	records := make([]outwallitems.Item, 0, len(items))
	for _, item := range items {
		definition, found := definitions[item.DefinitionID]
		if !found || definition.Kind != furnituremodel.KindWall || item.WallPosition == nil {
			continue
		}
		if _, found = seen[item.OwnerPlayerID]; !found {
			seen[item.OwnerPlayerID] = struct{}{}
			owners = append(owners, outwallitems.Owner{ID: item.OwnerPlayerID, Name: ownerNames[item.OwnerPlayerID]})
		}
		records = append(records, outwallitems.Item{ID: item.ID, SpriteID: definition.SpriteID, WallPosition: *item.WallPosition, ExtraData: WallExtraData(definition, item.ExtraData), UsagePolicy: UsagePolicyValue(definition), OwnerID: item.OwnerPlayerID})
	}

	return owners, records
}

// WallExtraData separates post-it visual color from text while preserving other wall-item state.
func WallExtraData(definition furnituremodel.Definition, extraData string) string {
	if definition.InteractionType == "postit" {
		return roomdecor.PostItColor(extraData)
	}

	return extraData
}

// floorItemRecord maps one persisted item and its definition to a protocol floor item.
func floorItemRecord(item furnituremodel.Item, definition furnituremodel.Definition, giftSenders map[int64]GiftSender) outflooritems.FloorItem {
	sender := giftSenderRecord(item, giftSenders)
	record := outflooritems.FloorItem{
		ID:               item.ID,
		SpriteID:         FurnitureSpriteID(item, definition),
		X:                *item.X,
		Y:                *item.Y,
		Rotation:         int(item.Rotation),
		Z:                FurnitureHeightValue(*item.Z),
		ExtraHeight:      ExtraHeightValue(definition),
		ExtraData:        item.ExtraData,
		UsagePolicy:      UsagePolicyValue(definition),
		Kind:             FurnitureKindValue(item),
		GiftWrapped:      item.GiftWrapped,
		OwnerID:          item.OwnerPlayerID,
		GiftMessage:      giftMessage(item),
		GiftProductCode:  definition.Name,
		GiftSenderName:   sender.Name,
		GiftSenderFigure: sender.Figure,
	}
	record.Data = SpecializedObjectData(definition.InteractionType, item.ExtraData)

	return record
}

// giftMessage returns the optional persisted gift message.
func giftMessage(item furnituremodel.Item) string {
	if item.GiftMessage == nil {
		return ""
	}

	return *item.GiftMessage
}

// giftSenderRecord resolves the optional persisted gift sender.
func giftSenderRecord(item furnituremodel.Item, giftSenders map[int64]GiftSender) GiftSender {
	if item.GiftSenderPlayerID == nil || giftSenders == nil {
		return GiftSender{}
	}

	return giftSenders[*item.GiftSenderPlayerID]
}

// FurnitureSpriteID returns the wrapper sprite for unopened gifts.
func FurnitureSpriteID(item furnituremodel.Item, definition furnituremodel.Definition) int {
	if item.GiftWrapped && item.GiftWrapSpriteID != nil {
		return int(*item.GiftWrapSpriteID)
	}

	return definition.SpriteID
}

// FurnitureKindValue returns the packed box and ribbon variant for unopened gifts.
func FurnitureKindValue(item furnituremodel.Item) int32 {
	if !item.GiftWrapped || item.GiftWrapBoxID == nil || item.GiftWrapRibbonID == nil {
		return 1
	}

	return *item.GiftWrapBoxID*1000 + *item.GiftWrapRibbonID
}

// ownerRecords maps distinct item owners into protocol owner records.
func ownerRecords(items []furnituremodel.Item, ownerNames map[int64]string) []outflooritems.Owner {
	seen := make(map[int64]struct{}, len(items))
	owners := make([]outflooritems.Owner, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item.OwnerPlayerID]; exists {
			continue
		}
		seen[item.OwnerPlayerID] = struct{}{}
		owners = append(owners, outflooritems.Owner{ID: item.OwnerPlayerID, Name: ownerNames[item.OwnerPlayerID]})
	}

	return owners
}

// ExtraHeightValue returns the walkable top height string for walk/sit definitions, matching Arcturus's
// serializeFloorData rule of only reporting it for allowWalk or allowSit items (not allowLay).
func ExtraHeightValue(definition furnituremodel.Definition) string {
	if !definition.AllowWalk && !definition.AllowSit {
		return ""
	}

	return FurnitureHeightValue(definition.StackHeight)
}

// UsagePolicyValue reports whether a definition exposes toggle-style interaction, matching
// Arcturus's isUsable() rule of having more than one interaction mode.
func UsagePolicyValue(definition furnituremodel.Definition) int32 {
	if definition.InteractionModesCount > 1 {
		return 1
	}

	return 0
}

// FurnitureHeightValue formats a persisted decimal height using the room world's rounded height.
func FurnitureHeightValue(value float64) string {
	return strconv.Itoa(int(roomfurniture.RoundHeight(value)))
}
