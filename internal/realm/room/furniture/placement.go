package furniture

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// ResolveWorldItem validates a target placement against a room's live world and the item's definition,
// returning the fully resolved world item ready to apply via Room.ReloadFurniture, along with the
// persisted definition used to build the protocol-facing broadcast afterward.
func ResolveWorldItem(ctx context.Context, active *roomlive.Room, manager furnitureservice.DefinitionFinder, itemID int64, definitionID int64, x int, y int, rotation furnituremodel.Rotation) (worldfurniture.Item, furnituremodel.Definition, error) {
	if !rotation.Valid() {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrInvalidTarget
	}
	point, ok := grid.NewPoint(x, y)
	if !ok {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrInvalidTarget
	}

	definition, found, err := manager.FindDefinitionByID(ctx, definitionID)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}
	if !found {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrDefinitionNotFound
	}

	worldDefinition, err := ToWorldDefinition(definition)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}

	footprint := worldfurniture.Footprint(point, worldDefinition.Width, worldDefinition.Length, worldunit.Rotation(rotation))
	height, err := active.ResolveFurniturePlacement(itemID, footprint)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}

	item := worldfurniture.Item{
		ID:         itemID,
		Definition: worldDefinition,
		Point:      point,
		Z:          height,
		Rotation:   worldunit.Rotation(rotation),
	}

	return item, definition, nil
}
