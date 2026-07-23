// Package admin owns protected pet management workflows.
package admin

import (
	"context"
	"strings"

	petcatalog "github.com/niflaot/pixels/internal/realm/pet/catalog"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpresence "github.com/niflaot/pixels/internal/realm/pet/presence"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// Service coordinates audited protected pet administration.
type Service struct {
	// store persists aggregates and audit records.
	store petrecord.Store
	// catalog validates names and appearance grants.
	catalog *petcatalog.Service
	// presence owns authoritative room transitions.
	presence *petpresence.Service
	// references publishes immutable reference generations.
	references petreference.Reader
	// runtime projects active pet mutations.
	runtime *petruntime.Service
	// rooms resolves active room worlds.
	rooms *roomlive.Registry
}

// New creates protected pet administration.
func New(store petrecord.Store, catalog *petcatalog.Service, presence *petpresence.Service, references petreference.Reader, runtime *petruntime.Service, rooms *roomlive.Registry) *Service {
	return &Service{store: store, catalog: catalog, presence: presence, references: references, runtime: runtime, rooms: rooms}
}

// Metrics returns process-wide low-cardinality pet telemetry.
func (service *Service) Metrics() petobservability.Snapshot {
	return service.runtime.Metrics().Snapshot()
}

// Audit stores required protected mutation attribution.
type Audit struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64
	// Reason stores the human-readable audit reason.
	Reason string
}

// Validate rejects missing protected mutation attribution.
func (audit Audit) Validate() error {
	if audit.ActorPlayerID <= 0 || strings.TrimSpace(audit.Reason) == "" {
		return petrecord.ErrInvalidState
	}
	return nil
}

// audit appends one mutation record inside the active transaction.
func (service *Service) audit(ctx context.Context, petID int64, audit Audit, action string) error {
	if err := audit.Validate(); err != nil {
		return err
	}
	return service.store.AppendAudit(ctx, petID, audit.ActorPlayerID, action, strings.TrimSpace(audit.Reason))
}
