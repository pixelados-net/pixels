package gallery

import (
	"context"
	"sync"
	"testing"

	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
)

// TestConcurrentPurchasesCreateIndependentCopies verifies native buy-another semantics.
func TestConcurrentPurchasesCreateIndependentCopies(t *testing.T) {
	store := galleryStoreForTest()
	service := New(store, &galleryFurniture{}, &galleryCurrencies{}, galleryPlayers{}, nil, camerametrics.New())
	errorsByCall := runConcurrent(t, func() error {
		_, err := service.Purchase(context.Background(), 7)
		return err
	}, func() error {
		_, err := service.Purchase(context.Background(), 7)
		return err
	})
	assertNoErrors(t, errorsByCall)
	if store.purchases != 2 {
		t.Fatalf("purchases=%d", store.purchases)
	}
}

// TestConcurrentPurchaseAndPublishBothSucceed verifies actions are no longer exclusive.
func TestConcurrentPurchaseAndPublishBothSucceed(t *testing.T) {
	store := galleryStoreForTest()
	service := New(store, &galleryFurniture{}, &galleryCurrencies{}, galleryPlayers{}, nil, camerametrics.New())
	errorsByCall := runConcurrent(t, func() error {
		_, err := service.Purchase(context.Background(), 7)
		return err
	}, func() error {
		_, _, err := service.Publish(context.Background(), 7)
		return err
	})
	assertNoErrors(t, errorsByCall)
	if store.purchases != 1 || !store.hasPublication {
		t.Fatalf("unexpected store state: %+v", store)
	}
}

// TestConcurrentPublishIsIdempotent verifies one charge and one publication.
func TestConcurrentPublishIsIdempotent(t *testing.T) {
	store := galleryStoreForTest()
	currencies := &galleryCurrencies{}
	metrics := camerametrics.New()
	service := New(store, &galleryFurniture{}, currencies, galleryPlayers{}, nil, metrics)
	errorsByCall := runConcurrent(t, func() error {
		_, _, err := service.Publish(context.Background(), 7)
		return err
	}, func() error {
		_, _, err := service.Publish(context.Background(), 7)
		return err
	})
	assertNoErrors(t, errorsByCall)
	if len(currencies.mutations) != 1 || metrics.Snapshot().Publications != 1 {
		t.Fatalf("mutations=%+v metrics=%+v", currencies.mutations, metrics.Snapshot())
	}
}

// runConcurrent executes two workflows after a shared start signal.
func runConcurrent(t *testing.T, first func() error, second func() error) <-chan error {
	t.Helper()
	errorsByCall := make(chan error, 2)
	start := make(chan struct{})
	var calls sync.WaitGroup
	for _, call := range []func() error{first, second} {
		calls.Add(1)
		go func(work func() error) {
			defer calls.Done()
			<-start
			errorsByCall <- work()
		}(call)
	}
	close(start)
	calls.Wait()
	close(errorsByCall)
	return errorsByCall
}

// assertNoErrors requires every concurrent workflow to succeed.
func assertNoErrors(t *testing.T, errorsByCall <-chan error) {
	t.Helper()
	for err := range errorsByCall {
		if err != nil {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
}
