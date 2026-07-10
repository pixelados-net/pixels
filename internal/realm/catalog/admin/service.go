package admin

import (
	"context"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
)

// Service implements catalog administration behavior.
type Service struct {
	// store persists catalog records.
	store catalogrepo.Store
	// catalog refreshes player-facing catalog data.
	catalog catalogservice.Manager
	// definitions validates furniture references.
	definitions furnitureservice.DefinitionGranter
}

// New creates a catalog administration service.
func New(store catalogrepo.Store, catalog catalogservice.Manager, definitions furnitureservice.DefinitionGranter) *Service {
	return &Service{store: store, catalog: catalog, definitions: definitions}
}

// Refresh reloads the player-facing catalog cache.
func (service *Service) Refresh(ctx context.Context) error {
	return service.catalog.Refresh(ctx)
}

// Pages lists all active catalog pages without player access filtering.
func (service *Service) Pages(ctx context.Context) ([]catalogmodel.Page, error) {
	return service.store.ListPages(ctx)
}

// Items lists active offers with an optional page filter.
func (service *Service) Items(ctx context.Context, pageID *int64) ([]catalogmodel.Item, error) {
	return service.store.ListItems(ctx, pageID)
}

// SanitizeList lists furniture definitions without active offers.
func (service *Service) SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error) {
	return service.store.SanitizeList(ctx)
}

// refresh reloads player-facing catalog data after a mutation.
func (service *Service) refresh(ctx context.Context) error {
	return service.catalog.Refresh(ctx)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
