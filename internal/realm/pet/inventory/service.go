// Package inventory owns cached pet inventory reads.
package inventory

import (
	"context"
	"sync"
	"time"

	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// state stores one immutable owner inventory generation.
type state struct {
	// ready closes after the cold durable read completes.
	ready chan struct{}
	// pets stores stable identifier-ordered records.
	pets []petrecord.Pet
	// err stores the cold read failure for concurrent waiters.
	err error
}

// Service caches immutable inventory generations by owner.
type Service struct {
	// store reads durable pet inventories.
	store petrecord.Store
	// mutex protects owner generations.
	mutex sync.Mutex
	// owners stores cached or loading generations.
	owners map[int64]*state
	// metrics records cold inventory read duration without touching warmed hits.
	metrics *petobservability.Metrics
}

// New creates pet inventory caching behavior.
func New(store petrecord.Store, metrics *petobservability.Metrics) *Service {
	return &Service{store: store, owners: make(map[int64]*state), metrics: metrics}
}

// List returns an immutable identifier-ordered owner inventory.
func (service *Service) List(ctx context.Context, ownerID int64) ([]petrecord.Pet, error) {
	service.mutex.Lock()
	current := service.owners[ownerID]
	if current != nil {
		ready := current.ready
		service.mutex.Unlock()
		if ready != nil {
			select {
			case <-ready:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return current.pets, current.err
	}
	current = &state{ready: make(chan struct{})}
	service.owners[ownerID] = current
	service.mutex.Unlock()

	started := time.Now()
	current.pets, current.err = service.store.Inventory(ctx, ownerID)
	service.metrics.ObserveInventoryList(time.Since(started))
	service.mutex.Lock()
	close(current.ready)
	current.ready = nil
	if current.err != nil && service.owners[ownerID] == current {
		delete(service.owners, ownerID)
	}
	service.mutex.Unlock()
	return current.pets, current.err
}

// Invalidate removes one owner generation after a committed mutation.
func (service *Service) Invalidate(ownerID int64) {
	service.mutex.Lock()
	delete(service.owners, ownerID)
	service.mutex.Unlock()
}
