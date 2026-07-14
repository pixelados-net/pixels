package effect

import (
	"context"
	"errors"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	effectgranted "github.com/niflaot/pixels/internal/realm/player/events/effectgranted"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outactivated "github.com/niflaot/pixels/networking/outbound/user/effect/activated"
	outremove "github.com/niflaot/pixels/networking/outbound/user/effect/remove"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
)

var (
	// ErrInvalidEffect reports malformed effect input.
	ErrInvalidEffect = errors.New("invalid player effect")
	// ErrEffectNotFound reports an unavailable effect.
	ErrEffectNotFound = errors.New("player effect not found")
)

// Service coordinates durable effects and live Nitro projections.
type Service struct {
	// store persists effects.
	store Store
	// permissions resolves the synthetic rank effect.
	permissions permissionservice.Manager
	// players resolves online connections.
	players *playerlive.Registry
	// connections sends direct packets.
	connections *netconn.Registry
	// rooms resolves current room presence.
	rooms *roomlive.Registry
	// events publishes committed lifecycle events.
	events bus.Publisher
	// now supplies wall time for deterministic tests.
	now func() time.Time
}

// New creates an effect service.
func New(store Store, permissions permissionservice.Manager, players *playerlive.Registry, connections *netconn.Registry, rooms *roomlive.Registry, events bus.Publisher) *Service {
	return &Service{store: store, permissions: permissions, players: players, connections: connections, rooms: rooms, events: events, now: time.Now}
}

// List returns durable effects plus the player's synthetic primary-group effect.
func (service *Service) List(ctx context.Context, playerID int64) ([]Effect, error) {
	effects, err := service.store.List(ctx, playerID)
	if err != nil {
		return nil, err
	}
	rank, found, err := service.rankEffect(ctx, playerID)
	if err != nil || !found {
		return effects, err
	}
	for _, item := range effects {
		if item.ID == rank.ID {
			return effects, nil
		}
	}
	return append(effects, rank), nil
}

// Grant adds one durable effect charge.
func (service *Service) Grant(ctx context.Context, playerID int64, effectID int32, durationSeconds int32, source Source) (Effect, error) {
	if playerID <= 0 || effectID <= 0 || durationSeconds < 0 {
		return Effect{}, ErrInvalidEffect
	}
	var granted Effect
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var grantErr error
		granted, grantErr = service.store.Grant(txCtx, playerID, effectID, durationSeconds)
		if grantErr != nil {
			return grantErr
		}
		service.deferProjection(txCtx, func(projectionCtx context.Context) {
			_ = service.sendAdded(projectionCtx, granted)
			service.publish(projectionCtx, effectgranted.Name, effectgranted.Payload{PlayerID: playerID, EffectID: effectID, Source: string(source)})
		})
		return nil
	})
	return granted, err
}

// GrantEnabled atomically adds, activates, and selects one durable effect charge.
func (service *Service) GrantEnabled(ctx context.Context, playerID int64, effectID int32, durationSeconds int32, source Source) (Effect, error) {
	var granted Effect
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var grantErr error
		granted, grantErr = service.Grant(txCtx, playerID, effectID, durationSeconds, source)
		if grantErr != nil {
			return grantErr
		}
		return service.enable(txCtx, playerID, effectID, source)
	})
	return granted, err
}

// Activate starts one effect charge without selecting it.
func (service *Service) Activate(ctx context.Context, playerID int64, effectID int32) (Effect, error) {
	if effectID <= 0 {
		return Effect{}, ErrInvalidEffect
	}
	effect, err := service.activate(ctx, playerID, effectID)
	if err != nil {
		return Effect{}, err
	}
	packet, err := outactivated.Encode(effect.ID, wireDuration(effect), effect.Permanent())
	if err == nil {
		err = service.send(ctx, playerID, packet)
	}
	return effect, err
}

// activate starts one available durable or synthetic effect without projection.
func (service *Service) activate(ctx context.Context, playerID int64, effectID int32) (Effect, error) {
	if rank, found, err := service.rankEffect(ctx, playerID); err != nil || found && rank.ID == effectID {
		return rank, err
	}
	effect, found, err := service.store.Activate(ctx, playerID, effectID, service.now().UTC())
	if err != nil {
		return Effect{}, err
	}
	if !found {
		return Effect{}, ErrEffectNotFound
	}
	return effect, nil
}

// Enable activates and selects one effect, or disables it with id zero.
func (service *Service) Enable(ctx context.Context, playerID int64, effectID int32) error {
	return service.enable(ctx, playerID, effectID, SourceAdmin)
}

// enable activates and selects one effect while preserving its source.
func (service *Service) enable(ctx context.Context, playerID int64, effectID int32, source Source) error {
	if effectID < 0 {
		return ErrInvalidEffect
	}
	var selected *int32
	var activated Effect
	if effectID > 0 {
		selected = &effectID
	}
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if effectID > 0 {
			effect, activateErr := service.activate(txCtx, playerID, effectID)
			if activateErr != nil {
				return activateErr
			}
			activated = effect
			if effect.Synthetic {
				source = SourceRank
			}
		}
		if setErr := service.store.SetActive(txCtx, playerID, selected); setErr != nil {
			return setErr
		}
		service.deferProjection(txCtx, func(projectionCtx context.Context) {
			if effectID > 0 {
				if packet, encodeErr := outactivated.Encode(effectID, wireDuration(activated), activated.Permanent()); encodeErr == nil {
					_ = service.send(projectionCtx, playerID, packet)
				}
			}
			_ = service.projectSelection(projectionCtx, playerID, effectID, source)
		})
		return nil
	})
	return err
}

// Revoke removes one durable effect and disables it when selected.
func (service *Service) Revoke(ctx context.Context, playerID int64, effectID int32) error {
	var disabled bool
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		removed, revokeErr := service.store.Revoke(txCtx, playerID, effectID)
		if revokeErr != nil {
			return revokeErr
		}
		if !removed {
			return ErrEffectNotFound
		}
		active, activeErr := service.store.Active(txCtx, playerID)
		if activeErr != nil {
			return activeErr
		}
		if active != nil && *active == effectID {
			disabled = true
			return service.store.SetActive(txCtx, playerID, nil)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if disabled {
		_ = service.projectSelection(ctx, playerID, 0, SourceAdmin)
	}
	packet, encodeErr := outremove.Encode(effectID)
	if encodeErr != nil {
		return encodeErr
	}
	return service.send(ctx, playerID, packet)
}

// deferProjection runs after a transaction commit or immediately outside a scope.
func (service *Service) deferProjection(ctx context.Context, callback func(context.Context)) {
	if !postgres.AfterCommit(ctx, callback) {
		callback(ctx)
	}
}

// publish emits a best-effort effect event.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
