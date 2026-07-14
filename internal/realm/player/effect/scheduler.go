package effect

import (
	"context"
	"time"

	effectexpired "github.com/niflaot/pixels/internal/realm/player/events/effectexpired"
	outremove "github.com/niflaot/pixels/networking/outbound/user/effect/remove"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	// expiryInterval stores the global effect sweep cadence.
	expiryInterval = time.Second
	// expiryBatchSize bounds work and row locks per sweep.
	expiryBatchSize int32 = 500
)

// RegisterScheduler starts one global effect expiration scheduler.
func RegisterScheduler(lifecycle fx.Lifecycle, service *Service, log *zap.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go service.runExpiry(ctx, log)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}

// runExpiry consumes expired charges on one global ticker.
func (service *Service) runExpiry(ctx context.Context, log *zap.Logger) {
	ticker := time.NewTicker(expiryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := service.expire(ctx, now.UTC()); err != nil && log != nil {
				log.Error("expire player effects", zap.Error(err))
			}
		}
	}
}

// expire consumes one bounded batch and projects committed removals.
func (service *Service) expire(ctx context.Context, now time.Time) error {
	var expired []Expiration
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var expireErr error
		expired, expireErr = service.store.Expire(txCtx, now, expiryBatchSize)
		return expireErr
	})
	if err != nil {
		return err
	}
	for _, item := range expired {
		if item.Selected {
			_ = service.projectSelection(ctx, item.PlayerID, 0, SourceAdmin)
		}
		if item.RemainingCharges == 0 {
			packet, encodeErr := outremove.Encode(item.EffectID)
			if encodeErr != nil {
				return encodeErr
			}
			_ = service.send(ctx, item.PlayerID, packet)
		} else {
			_ = service.SendInventory(ctx, item.PlayerID)
		}
		service.publish(ctx, effectexpired.Name, effectexpired.Payload{PlayerID: item.PlayerID, EffectID: item.EffectID,
			RemainingCharges: item.RemainingCharges, Source: string(SourceAdmin)})
	}
	return nil
}

// eventTypeAssertion keeps the event contract linked at compile time.
var _ bus.Name = effectexpired.Name
