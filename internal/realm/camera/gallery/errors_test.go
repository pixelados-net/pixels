package gallery

import (
	"context"
	"errors"
	"testing"

	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
)

// TestPurchaseExpectedFailures verifies disabled, missing, and balance failures.
func TestPurchaseExpectedFailures(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*galleryStore, *galleryCurrencies)
		expected  error
	}{
		{name: "disabled", configure: func(store *galleryStore, _ *galleryCurrencies) { store.settings.Enabled = false }, expected: camerarecord.ErrDisabled},
		{name: "missing capture", configure: func(store *galleryStore, _ *galleryCurrencies) { store.active.ID = 0 }, expected: camerarecord.ErrNoPendingCapture},
		{name: "points", configure: func(_ *galleryStore, currencies *galleryCurrencies) { currencies.failType = 5 }, expected: camerarecord.ErrInsufficientPoints},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := galleryStoreForTest()
			currencies := &galleryCurrencies{}
			test.configure(store, currencies)
			service := New(store, &galleryFurniture{}, currencies, galleryPlayers{}, nil, camerametrics.New())
			if _, err := service.Purchase(context.Background(), 7); !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
			if store.purchases != 0 {
				t.Fatalf("failed purchase linked %d items", store.purchases)
			}
		})
	}
}

// TestPublishRetryReturnsExistingPublicationWithoutCharge verifies idempotency.
func TestPublishRetryReturnsExistingPublicationWithoutCharge(t *testing.T) {
	store := galleryStoreForTest()
	store.publication = camerarecord.Publication{ID: 12, CaptureID: store.active.ID, URL: store.active.URL}
	store.hasPublication = true
	currencies := &galleryCurrencies{}
	service := New(store, &galleryFurniture{}, currencies, galleryPlayers{}, nil, camerametrics.New())
	publication, remaining, err := service.Publish(context.Background(), 7)
	if err != nil || publication.ID != 12 || remaining != 0 || len(currencies.mutations) != 0 {
		t.Fatalf("publication=%+v remaining=%s mutations=%+v err=%v", publication, remaining, currencies.mutations, err)
	}
}

// TestPublishExpectedFailures verifies disabled, missing, and balance failures.
func TestPublishExpectedFailures(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*galleryStore, *galleryCurrencies)
		expected  error
	}{
		{name: "disabled", configure: func(store *galleryStore, _ *galleryCurrencies) { store.settings.Enabled = false }, expected: camerarecord.ErrDisabled},
		{name: "missing capture", configure: func(store *galleryStore, _ *galleryCurrencies) { store.active.ID = 0 }, expected: camerarecord.ErrNoPendingCapture},
		{name: "points", configure: func(_ *galleryStore, currencies *galleryCurrencies) { currencies.failType = 5 }, expected: camerarecord.ErrInsufficientPoints},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := galleryStoreForTest()
			currencies := &galleryCurrencies{}
			test.configure(store, currencies)
			service := New(store, &galleryFurniture{}, currencies, galleryPlayers{}, nil, camerametrics.New())
			if _, _, err := service.Publish(context.Background(), 7); !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
			if store.hasPublication {
				t.Fatal("failed publication was persisted")
			}
		})
	}
}
