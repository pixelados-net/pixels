package furniture

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// ResolveWorldItem validates a target placement against a room's live world and the item's definition,
// returning the fully resolved world item ready to apply via Room.ReloadFurniture, along with the
// persisted definition used to build the protocol-facing broadcast afterward.
func ResolveWorldItem(ctx context.Context, active *roomlive.Room, manager furnitureservice.DefinitionFinder, persisted furnituremodel.Item, x int, y int, rotation furnituremodel.Rotation) (worldfurniture.Item, furnituremodel.Definition, error) {
	if !rotation.Valid() {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrInvalidTarget
	}
	point, ok := grid.NewPoint(x, y)
	if !ok {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrInvalidTarget
	}

	definition, found, err := manager.FindDefinitionByID(ctx, persisted.DefinitionID)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}
	if !found {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrDefinitionNotFound
	}
	if definition.Kind == furnituremodel.KindWall {
		return worldfurniture.Item{}, furnituremodel.Definition{}, ErrInvalidTarget
	}

	worldDefinition, err := ToWorldDefinition(definition)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}

	footprint := worldfurniture.Footprint(point, worldDefinition.Width, worldDefinition.Length, worldunit.Rotation(rotation))
	height, err := active.ResolveFurniturePlacement(persisted.ID, footprint)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}

	item := worldfurniture.Item{
		ID:            persisted.ID,
		OwnerPlayerID: persisted.OwnerPlayerID,
		Definition:    worldDefinition,
		Point:         point,
		Z:             height,
		Rotation:      worldunit.Rotation(rotation),
		ExtraData:     persisted.ExtraData,
	}

	return item, definition, nil
}

// ValidateRollerPlacement rejects rollers supported by any other furniture unless rules are disabled.
func ValidateRollerPlacement(active *roomlive.Room, item worldfurniture.Item, noRules bool) error {
	if noRules || item.Definition.InteractionType != "roller" {
		return nil
	}
	for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		for _, placed := range active.FurnitureAt(point) {
			if placed.ID != item.ID {
				return roomlive.ErrCannotStack
			}
		}
	}
	return nil
}
