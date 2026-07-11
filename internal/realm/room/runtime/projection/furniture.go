package projection

import (
	"strconv"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	outflooritems "github.com/niflaot/pixels/networking/outbound/room/furniture/flooritems"
)

// FloorItems maps placed furniture items into ROOM_FLOOR_ITEMS records.
func FloorItems(items []furnituremodel.Item, definitions map[int64]furnituremodel.Definition, ownerNames map[int64]string) ([]outflooritems.Owner, []outflooritems.FloorItem) {
	owners := ownerRecords(items, ownerNames)
	records := make([]outflooritems.FloorItem, 0, len(items))
	for _, item := range items {
		definition, ok := definitions[item.DefinitionID]
		if !ok || item.X == nil || item.Y == nil || item.Z == nil {
			continue
		}
		records = append(records, floorItemRecord(item, definition))
	}

	return owners, records
}

// floorItemRecord maps one persisted item and its definition to a protocol floor item.
func floorItemRecord(item furnituremodel.Item, definition furnituremodel.Definition) outflooritems.FloorItem {
	return outflooritems.FloorItem{
		ID:          item.ID,
		SpriteID:    definition.SpriteID,
		X:           *item.X,
		Y:           *item.Y,
		Rotation:    int(item.Rotation),
		Z:           FurnitureHeightValue(*item.Z),
		ExtraHeight: ExtraHeightValue(definition),
		ExtraData:   item.ExtraData,
		UsagePolicy: UsagePolicyValue(definition),
		OwnerID:     item.OwnerPlayerID,
	}
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
