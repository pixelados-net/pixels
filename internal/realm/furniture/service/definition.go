package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// FindDefinitionByID finds a furniture definition by id.
func (service *Service) FindDefinitionByID(ctx context.Context, id int64) (furnituremodel.Definition, bool, error) {
	if id <= 0 {
		return furnituremodel.Definition{}, false, ErrInvalidDefinitionID
	}

	return service.store.FindDefinitionByID(ctx, id)
}

// ListDefinitions lists furniture definitions.
func (service *Service) ListDefinitions(ctx context.Context) ([]furnituremodel.Definition, error) {
	return service.store.ListDefinitions(ctx)
}
