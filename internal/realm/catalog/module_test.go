package catalog

import (
	"context"
	"testing"

	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestProvidersExposeCatalogContracts verifies catalog Fx provider adapters.
func TestProvidersExposeCatalogContracts(t *testing.T) {
	service := &catalogservice.Service{}
	adminService := &catalogadmin.Service{}
	if NewStore(nil) == nil || NewManager(service) == nil || NewReader(service) == nil || NewAdminManager(adminService) == nil {
		t.Fatal("expected catalog providers")
	}
}

// TestRegisterConnectionHandlersAddsCatalogPackets verifies catalog packet wiring.
func TestRegisterConnectionHandlersAddsCatalogPackets(t *testing.T) {
	handlers := &realmconn.Handlers{Inbound: netconn.NewHandlerRegistry()}
	RegisterConnectionHandlers(handlers, HandlerDeps{Players: playerlive.NewRegistry(), Bindings: binding.NewRegistry(),
		Catalog: &catalogservice.Service{}, Log: zap.NewNop()})
	if handlers.Inbound.Len() != 3 {
		t.Fatalf("expected three catalog handlers, got %d", handlers.Inbound.Len())
	}
	RegisterConnectionHandlers(nil, HandlerDeps{})
}

// TestRegisterLifecycleRefreshesCatalog verifies startup cache initialization.
func TestRegisterLifecycleRefreshesCatalog(t *testing.T) {
	store := &lifecycleStore{}
	service := catalogservice.New(store, nil, lifecycleFurniture{}, nil, nil)
	lifecycle := fxtest.NewLifecycle(t)
	RegisterLifecycle(lifecycle, service)
	lifecycle.RequireStart()
	lifecycle.RequireStop()
	if store.reads != 2 {
		t.Fatalf("expected page and item startup reads, got %d", store.reads)
	}
}

// lifecycleFurniture supplies startup furniture definitions.
type lifecycleFurniture struct{}

// FindDefinitionByID finds no startup furniture definition.
func (lifecycleFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions lists no startup furniture definitions.
func (lifecycleFurniture) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}

// Grant supplies unused startup furniture creation behavior.
func (lifecycleFurniture) Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	return nil, nil
}

// lifecycleStore supplies startup catalog rows.
type lifecycleStore struct {
	// Store supplies unused catalog persistence methods.
	catalogrepo.Store

	// reads counts startup reads.
	reads int
}

// ListPages records one startup page read.
func (store *lifecycleStore) ListPages(context.Context) ([]catalogmodel.Page, error) {
	store.reads++

	return nil, nil
}

// ListItems records one startup offer read.
func (store *lifecycleStore) ListItems(context.Context, *int64) ([]catalogmodel.Item, error) {
	store.reads++

	return nil, nil
}
