package furniture

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/furniture/service"
)

// TestProvidersExposeContracts verifies module helper providers return contracts.
func TestProvidersExposeContracts(t *testing.T) {
	furnitureService := service.New(nil)

	if NewStore(nil) == nil {
		t.Fatal("expected furniture store")
	}
	if NewManager(furnitureService) == nil {
		t.Fatal("expected furniture manager")
	}
}
