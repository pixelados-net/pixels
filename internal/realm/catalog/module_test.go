package catalog

import (
	"context"
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	"go.uber.org/fx/fxtest"
)

// TestProvidersExposeCatalogContracts verifies catalog Fx provider adapters.
func TestProvidersExposeCatalogContracts(t *testing.T) {
	service := &catalogservice.Service{}
	if NewStore(nil) == nil || NewManager(service) == nil || NewReader(service) == nil {
		t.Fatal("expected catalog providers")
	}
}

// TestRegisterLifecycleRefreshesCatalog verifies startup cache initialization.
func TestRegisterLifecycleRefreshesCatalog(t *testing.T) {
	store := &lifecycleStore{}
	service := catalogservice.New(store, nil, nil, nil, nil)
	lifecycle := fxtest.NewLifecycle(t)
	RegisterLifecycle(lifecycle, service)
	lifecycle.RequireStart()
	lifecycle.RequireStop()
	if store.reads != 2 {
		t.Fatalf("expected page and item startup reads, got %d", store.reads)
	}
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
