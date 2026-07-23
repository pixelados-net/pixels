package furniture

import (
	"context"
	"fmt"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// LoadRoomFurniture loads placed furniture items for a room as room world furniture items.
func LoadRoomFurniture(ctx context.Context, manager furnitureservice.Manager, roomID int64) ([]worldfurniture.Item, error) {
	persisted, err := manager.ListRoomItems(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room furniture items: %w", err)
	}
	if len(persisted) == 0 {
		return nil, nil
	}

	definitions, err := DefinitionsByID(ctx, manager)
	if err != nil {
		return nil, err
	}

	items := make([]worldfurniture.Item, 0, len(persisted))
	for _, item := range persisted {
		worldItem, ok, err := ToWorldItem(item, definitions)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		items = append(items, worldItem)
	}

	return items, nil
}
