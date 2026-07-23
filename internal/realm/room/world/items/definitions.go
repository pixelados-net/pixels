// Package furniture bridges persisted furniture records into room world furniture.
package furniture

import (
	"context"
	"fmt"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
)

// DefinitionsByID indexes furniture definitions by id.
func DefinitionsByID(ctx context.Context, manager furnitureservice.Manager) (map[int64]furnituremodel.Definition, error) {
	definitions, err := manager.ListDefinitions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list furniture definitions: %w", err)
	}

	indexed := make(map[int64]furnituremodel.Definition, len(definitions))
	for _, definition := range definitions {
		indexed[definition.ID] = definition
	}

	return indexed, nil
}
