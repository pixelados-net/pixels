// Package gamecenter wires the external Game Center lobby.
package gamecenter

import (
	"context"

	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	gamecenterconfig "github.com/niflaot/pixels/internal/realm/gamecenter/config"
	gamecenterdb "github.com/niflaot/pixels/internal/realm/gamecenter/database"
	gamecenterlobby "github.com/niflaot/pixels/internal/realm/gamecenter/lobby"
	gamehandlers "github.com/niflaot/pixels/internal/realm/gamecenter/lobby/handlers"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	"go.uber.org/fx"
)

// Module provides Game Center persistence, caching, and packet routing.
var Module = fx.Module("realm-gamecenter", fx.Provide(
	gamecenterconfig.Load,
	gamecenterdb.New,
	NewStore,
	gamecenterlobby.New,
), fx.Invoke(RegisterHandlers, RegisterLifecycle))

// NewStore exposes PostgreSQL Game Center persistence.
func NewStore(repository *gamecenterdb.Repository) gamecenterrecord.Store { return repository }

// RegisterHandlers installs Game Center packet adapters.
func RegisterHandlers(handlers *realmconn.Handlers, lobby *gamecenterlobby.Service) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	gamehandlers.Register(handlers.Inbound, gamehandlers.Handler{Lobby: lobby})
}

// RegisterLifecycle loads the immutable lobby cache before serving traffic.
func RegisterLifecycle(lifecycle fx.Lifecycle, lobby *gamecenterlobby.Service) {
	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error { return lobby.Reload(ctx) }})
}
