package player

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	achievementdb "github.com/niflaot/pixels/internal/realm/player/database/achievement"
	effectdb "github.com/niflaot/pixels/internal/realm/player/database/effect"
	identitydb "github.com/niflaot/pixels/internal/realm/player/database/identity"
	profiledb "github.com/niflaot/pixels/internal/realm/player/database/profile"
	settingsdb "github.com/niflaot/pixels/internal/realm/player/database/settings"
	wardrobedb "github.com/niflaot/pixels/internal/realm/player/database/wardrobe"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerfigure "github.com/niflaot/pixels/internal/realm/player/figure"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	"github.com/niflaot/pixels/internal/realm/player/live"
	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	"github.com/niflaot/pixels/internal/realm/player/repository"
	"github.com/niflaot/pixels/internal/realm/player/service"
	playersettings "github.com/niflaot/pixels/internal/realm/player/settings"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
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
		NewTradeManager,
		achievementdb.New,
		NewAchievementStore,
		playerachievement.New,
		effectdb.New,
		NewEffectStore,
		playereffect.New,
		NewEffectManager,
		settingsdb.New,
		NewSettingsStore,
		playersettings.LoadConfig,
		playersettings.New,
		playersettings.NewWriter,
		profiledb.New,
		NewProfileStore,
		NewProfileClothingFinder,
		playerprofile.LoadConfig,
		playerprofile.NewConfigured,
		playerfigure.LoadConfig,
		playerfigure.NewCatalog,
		wardrobedb.New,
		NewWardrobeStore,
		playerwardrobe.LoadConfig,
		playerwardrobe.NewConfigured,
		identitydb.New,
		NewIdentityStore,
		playeridentity.LoadConfig,
		playeridentity.NewConfigured,
	),
	fx.Invoke(playereffect.RegisterBootstrap, playereffect.RegisterScheduler, playerachievement.RegisterLifecycle, playerwardrobe.RegisterBootstrap, playersettings.RegisterWriter),
)

// NewIdentityStore exposes atomic PostgreSQL renames through their domain boundary.
func NewIdentityStore(repository *identitydb.Repository) playeridentity.Store {
	return repository
}

// NewWardrobeStore exposes PostgreSQL wardrobe persistence through its domain boundary.
func NewWardrobeStore(repository *wardrobedb.Repository) playerwardrobe.Store {
	return repository
}

// NewProfileStore exposes PostgreSQL profile persistence through its domain boundary.
func NewProfileStore(repository *profiledb.Repository) playerprofile.Store {
	return repository
}

// NewProfileClothingFinder exposes clothing unlock reads to figure validation.
func NewProfileClothingFinder(repository *wardrobedb.Repository) playerprofile.ClothingFinder {
	return repository
}

// NewSettingsStore exposes PostgreSQL settings persistence through its domain boundary.
func NewSettingsStore(repository *settingsdb.Repository) playersettings.Store {
	return repository
}

// NewAchievementStore exposes badge and respect persistence through its domain boundary.
func NewAchievementStore(repository *achievementdb.Repository) playerachievement.Store {
	return repository
}

// NewEffectStore exposes PostgreSQL effect persistence through its domain boundary.
func NewEffectStore(repository *effectdb.Repository) playereffect.Store {
	return repository
}

// NewEffectManager exposes player effect behavior through its domain boundary.
func NewEffectManager(service *playereffect.Service) playereffect.Manager {
	return service
}

// NewTradeManager exposes durable trade eligibility changes.
func NewTradeManager(playerService *service.Service) service.TradeManager {
	return playerService
}

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
