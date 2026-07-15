package furniture

import (
	"encoding/json"
	"fmt"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// metadataSlots stores the metadata.slots JSON shape declared on furniture definitions.
type metadataSlots struct {
	// Slots stores declared sit/lay slots in unrotated local coordinates.
	Slots []struct {
		// DX stores the local x offset within the footprint at rotation 0.
		DX int `json:"dx"`

		// DY stores the local y offset within the footprint at rotation 0.
		DY int `json:"dy"`

		// Status stores "sit" or "lay".
		Status string `json:"status"`

		// BodyRotation stores the forced body facing at rotation 0.
		BodyRotation int `json:"body_rotation"`
	} `json:"slots"`
}

// ToWorldItem converts a persisted furniture item into a room world furniture item.
func ToWorldItem(item furnituremodel.Item, definitions map[int64]furnituremodel.Definition) (worldfurniture.Item, bool, error) {
	if item.X == nil || item.Y == nil || item.Z == nil {
		return worldfurniture.Item{}, false, nil
	}
	definition, found := definitions[item.DefinitionID]
	if !found {
		return worldfurniture.Item{}, false, nil
	}
	if definition.Kind == furnituremodel.KindWall {
		return worldfurniture.Item{}, false, nil
	}
	point, ok := grid.NewPoint(*item.X, *item.Y)
	if !ok {
		return worldfurniture.Item{}, false, nil
	}

	worldDefinition, err := ToWorldDefinition(definition)
	if err != nil {
		return worldfurniture.Item{}, false, err
	}

	worldItem := worldfurniture.Item{
		ID:            item.ID,
		OwnerPlayerID: item.OwnerPlayerID,
		Definition:    worldDefinition,
		Point:         point,
		Z:             RoundHeight(*item.Z),
		Rotation:      worldunit.Rotation(item.Rotation),
		ExtraData:     item.ExtraData,
	}
	worldItem.Definition.StackHeight = worldItem.Definition.HeightAtState(worldItem.ExtraData)

	return worldItem, true, nil
}

// ToWorldDefinition converts a persisted furniture definition into a room world definition.
func ToWorldDefinition(definition furnituremodel.Definition) (worldfurniture.Definition, error) {
	slots, err := parseSlots(definition.Metadata)
	if err != nil {
		return worldfurniture.Definition{}, fmt.Errorf("parse furniture definition %d metadata: %w", definition.ID, err)
	}

	return worldfurniture.Definition{
		SpriteID:              definition.SpriteID,
		InteractionType:       definition.InteractionType,
		InteractionModesCount: definition.InteractionModesCount,
		Multiheight:           definition.Multiheight,
		CustomParams:          definition.CustomParams,
		EffectPool:            definition.EffectPool,
		EffectMale:            definition.EffectMale,
		EffectFemale:          definition.EffectFemale,
		Width:                 definition.Width,
		Length:                definition.Length,
		StackHeight:           RoundHeight(definition.StackHeight),
		AllowStack:            definition.AllowStack,
		AllowWalk:             definition.AllowWalk,
		AllowSit:              definition.AllowSit,
		AllowLay:              definition.AllowLay,
		Slots:                 slots,
	}, nil
}

// parseSlots parses declared sit/lay slots from definition metadata.
func parseSlots(metadata json.RawMessage) ([]worldfurniture.SlotDefinition, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	var parsed metadataSlots
	if err := json.Unmarshal(metadata, &parsed); err != nil {
		return nil, err
	}

	slots := make([]worldfurniture.SlotDefinition, 0, len(parsed.Slots))
	for _, slot := range parsed.Slots {
		status := worldfurniture.SlotStatusSit
		if slot.Status == string(worldfurniture.SlotStatusLay) {
			status = worldfurniture.SlotStatusLay
		}
		slots = append(slots, worldfurniture.SlotDefinition{
			DX:           slot.DX,
			DY:           slot.DY,
			Status:       status,
			BodyRotation: worldunit.Rotation(slot.BodyRotation),
		})
	}

	return slots, nil
}

// RoundHeight rounds a persisted decimal height into a compact grid height.
func RoundHeight(value float64) grid.Height {
	return grid.HeightFromUnits(value)
}
