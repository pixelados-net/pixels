package marketplace

import (
	"context"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	marketbrowse "github.com/niflaot/pixels/internal/realm/marketplace/browse"
	marketcommerce "github.com/niflaot/pixels/internal/realm/marketplace/commerce"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	marketdb "github.com/niflaot/pixels/internal/realm/marketplace/database"
	marketlisting "github.com/niflaot/pixels/internal/realm/marketplace/listing"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"go.uber.org/fx"
	"time"
)

// Module provides Marketplace persistence, protocol, and scheduling.
var Module = fx.Module("realm-marketplace", fx.Provide(LoadConfig, marketdb.New, NewStore, marketcore.New), fx.Invoke(RegisterHandlers, RegisterExpiry))

// NewStore exposes the Marketplace persistence contract.
func NewStore(repository *marketdb.Repository) marketrecord.Store { return repository }

// RegisterHandlers installs Marketplace packet adapters.
func RegisterHandlers(handlers *realmconn.Handlers, service *marketcore.Service, bindings *binding.Registry) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	marketbrowse.Register(handlers.Inbound, marketbrowse.Handler{Service: service})
	marketlisting.Register(handlers.Inbound, marketlisting.Handler{Service: service, Bindings: bindings})
	marketcommerce.Register(handlers.Inbound, marketcommerce.Handler{Service: service, Bindings: bindings})
}

// RegisterExpiry starts and stops the Marketplace expiry scheduler.
func RegisterExpiry(lifecycle fx.Lifecycle, config Config, service *marketcore.Service) {
	var cancel context.CancelFunc
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		go func() {
			ticker := time.NewTicker(config.ExpiryInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					_, _ = service.Expire(ctx)
				}
			}
		}()
		return nil
	}, OnStop: func(context.Context) error {
		if cancel != nil {
			cancel()
		}
		return nil
	}})
}
