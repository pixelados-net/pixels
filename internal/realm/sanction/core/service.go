// Package core coordinates global punishment behavior.
package core

import (
	"context"
	"strings"
	"sync"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	sanctionapplied "github.com/niflaot/pixels/internal/realm/sanction/events/applied"
	sanctionrevoked "github.com/niflaot/pixels/internal/realm/sanction/events/revoked"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// Applier executes the immediate side effects for one punishment kind.
type Applier interface {
	// Kind identifies the registered punishment.
	Kind() sanctionrecord.Kind
	// Apply executes the immediate effect after persistence.
	Apply(context.Context, sanctionrecord.Punishment) error
	// Revoke removes the effect after revocation when no overlap remains.
	Revoke(context.Context, sanctionrecord.Punishment) error
}

// Service is the only global punishment mutation entry point.
type Service struct {
	// store persists punishment truth.
	store sanctionrecord.Store
	// permissions authorizes player issuers and immune targets.
	permissions permissionservice.Checker
	// events publishes committed sanction facts.
	events bus.Publisher
	// logger records best-effort effect failures.
	logger *zap.Logger
	// mutex protects behavior registration.
	mutex sync.RWMutex
	// appliers stores behaviors by kind.
	appliers map[sanctionrecord.Kind]Applier
	// now supplies deterministic timestamps.
	now func() time.Time
}

// New creates a global sanction service.
func New(store sanctionrecord.Store, permissions permissionservice.Checker, events bus.Publisher, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{store: store, permissions: permissions, events: events, logger: logger, appliers: make(map[sanctionrecord.Kind]Applier), now: time.Now}
}

// Register installs one punishment behavior.
func (service *Service) Register(applier Applier) error {
	if applier == nil || !applier.Kind().Valid() {
		return ErrInvalidRequest
	}
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if _, exists := service.appliers[applier.Kind()]; exists {
		return ErrApplierExists
	}
	service.appliers[applier.Kind()] = applier
	return nil
}

// Apply validates, persists, executes, and publishes one punishment.
func (service *Service) Apply(ctx context.Context, params sanctionrecord.ApplyParams) (sanctionrecord.Punishment, error) {
	params = normalize(params)
	if !valid(params, service.now()) {
		return sanctionrecord.Punishment{}, ErrInvalidRequest
	}
	if err := service.authorize(ctx, params); err != nil {
		return sanctionrecord.Punishment{}, err
	}
	punishment, err := service.store.Insert(ctx, params)
	if err != nil {
		return sanctionrecord.Punishment{}, err
	}
	if effect := service.applier(punishment.Kind); effect != nil {
		if effectErr := effect.Apply(ctx, punishment); effectErr != nil {
			service.logger.Warn("sanction side effect failed", zap.Int64("punishment_id", punishment.ID), zap.String("kind", string(punishment.Kind)), zap.Error(effectErr))
		}
	}
	service.publishApplied(ctx, punishment)
	return punishment, nil
}

// Revoke marks a punishment revoked and removes its effect when no overlap remains.
func (service *Service) Revoke(ctx context.Context, id int64, actorID int64) (sanctionrecord.Punishment, error) {
	if id <= 0 || actorID <= 0 {
		return sanctionrecord.Punishment{}, ErrInvalidRequest
	}
	allowed, err := service.permissions.HasPermission(ctx, actorID, ApplyNode)
	if err != nil {
		return sanctionrecord.Punishment{}, err
	}
	if !allowed {
		return sanctionrecord.Punishment{}, ErrUnauthorized
	}
	now := service.now()
	punishment, updated, err := service.store.Revoke(ctx, id, &actorID, now)
	return service.finishRevoke(ctx, punishment, updated, err)
}

// RevokeSystem revokes one punishment through an already-authorized system boundary.
func (service *Service) RevokeSystem(ctx context.Context, id int64) (sanctionrecord.Punishment, error) {
	if id <= 0 {
		return sanctionrecord.Punishment{}, ErrInvalidRequest
	}
	punishment, updated, err := service.store.Revoke(ctx, id, nil, service.now())
	return service.finishRevoke(ctx, punishment, updated, err)
}

// finishRevoke applies common side effects after one revocation mutation.
func (service *Service) finishRevoke(ctx context.Context, punishment sanctionrecord.Punishment, updated bool, err error) (sanctionrecord.Punishment, error) {
	if err != nil {
		return sanctionrecord.Punishment{}, err
	}
	if !updated {
		return sanctionrecord.Punishment{}, ErrNotFound
	}
	if effect := service.applier(punishment.Kind); effect != nil {
		if effectErr := effect.Revoke(ctx, punishment); effectErr != nil {
			service.logger.Warn("sanction revoke side effect failed", zap.Int64("punishment_id", punishment.ID), zap.Error(effectErr))
		}
	}
	service.publishRevoked(ctx, punishment)
	return punishment, nil
}

// Active returns the current timestamp-derived projection.
func (service *Service) Active(ctx context.Context, playerID int64) (sanctionrecord.ActiveState, error) {
	return service.store.Active(ctx, playerID, service.now())
}

// History returns recent punishment history.
func (service *Service) History(ctx context.Context, playerID int64, limit int32) ([]sanctionrecord.Punishment, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return service.store.List(ctx, playerID, limit)
}

// Store returns the persistence boundary for focused administrative reads.
func (service *Service) Store() sanctionrecord.Store { return service.store }

// CheckBan reports an active login ban and its visible reason.
func (service *Service) CheckBan(ctx context.Context, playerID int64) (bool, string, error) {
	state, err := service.Active(ctx, playerID)
	if err != nil || state.Ban == nil {
		return false, "", err
	}
	return true, state.Ban.Reason, nil
}

// applier returns a registered behavior.
func (service *Service) applier(kind sanctionrecord.Kind) Applier {
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	return service.appliers[kind]
}

// authorize enforces issuer capability and target immunity.
func (service *Service) authorize(ctx context.Context, params sanctionrecord.ApplyParams) error {
	immune, err := service.permissions.HasPermission(ctx, params.ReceiverPlayerID, ImmuneNode)
	if err != nil {
		return err
	}
	if immune {
		return ErrImmune
	}
	if params.IssuerKind == "system" {
		return nil
	}
	if params.IssuerPlayerID == nil {
		return ErrUnauthorized
	}
	node := ApplyNode
	if params.Kind == sanctionrecord.KindBan {
		node = BanNode
	}
	allowed, err := service.permissions.HasPermission(ctx, *params.IssuerPlayerID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrUnauthorized
	}
	return nil
}

// normalize trims bounded text and fills issuer defaults.
func normalize(params sanctionrecord.ApplyParams) sanctionrecord.ApplyParams {
	params.Reason = strings.TrimSpace(params.Reason)
	params.Source = strings.TrimSpace(params.Source)
	params.IssuerKind = strings.TrimSpace(params.IssuerKind)
	if params.IssuerKind == "" {
		params.IssuerKind = "player"
	}
	if len(params.Reason) > 500 {
		params.Reason = params.Reason[:500]
	}
	if len(params.Source) > 50 {
		params.Source = params.Source[:50]
	}
	if params.Kind.Instant() {
		params.ExpiresAt = nil
	}
	return params
}

// valid reports whether persistence constraints will accept the request.
func valid(params sanctionrecord.ApplyParams, now time.Time) bool {
	validExpiry := params.ExpiresAt == nil || params.ExpiresAt.After(now) && !params.ExpiresAt.After(now.Add(time.Duration(sanctionrecord.MaxDurationHours)*time.Hour))
	return params.ReceiverPlayerID > 0 && params.Kind.Valid() && params.Reason != "" && params.Source != "" && validExpiry && (params.IssuerKind == "system" || params.IssuerKind == "player" && params.IssuerPlayerID != nil)
}

// publishApplied emits one best-effort applied event.
func (service *Service) publishApplied(ctx context.Context, value sanctionrecord.Punishment) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: sanctionapplied.Name, Payload: sanctionapplied.Payload{PunishmentID: value.ID, ReceiverID: value.ReceiverPlayerID, Kind: value.Kind, ExpiresAt: value.ExpiresAt, Source: value.Source}})
	}
}

// publishRevoked emits one best-effort revoked event.
func (service *Service) publishRevoked(ctx context.Context, value sanctionrecord.Punishment) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: sanctionrevoked.Name, Payload: sanctionrevoked.Payload{PunishmentID: value.ID, ReceiverID: value.ReceiverPlayerID, Kind: value.Kind}})
	}
}
