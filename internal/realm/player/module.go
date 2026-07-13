package player

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/player/repository"
	"github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides player realm persistence and runtime state.
var Module = fx.Module(
	"realm-player",
	fx.Provide(
		NewStore,
		live.NewRegistry,
		NewService,
		NewCreator,
		NewFinder,
		NewClubWriter,
		NewManager,
		NewAdminManager,
	),
)

// NewService creates player behavior with default permission assignment.
func NewService(store repository.Store, permissions permissionservice.DefaultAssigner) *service.Service {
	return service.New(store, permissions)
}

// NewClubWriter exposes derived club entitlement updates.
func NewClubWriter(playerService *service.Service) service.ClubWriter {
	return playerService
}

// NewStore creates the player persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewCreator exposes the player creation contract.
func NewCreator(playerService *service.Service) service.Creator {
	return playerService
}

// NewFinder exposes the player lookup contract.
func NewFinder(playerService *service.Service) service.Finder {
	return playerService
}

// NewManager exposes the player management contract.
func NewManager(playerService *service.Service) service.Manager {
	return playerService
}

// NewAdminManager exposes protected player administration behavior.
func NewAdminManager(playerService *service.Service) service.AdminManager {
	return playerService
}
