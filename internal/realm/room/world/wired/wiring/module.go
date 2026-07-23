// Package wiring composes the room WIRED capability without coupling domain packages to PostgreSQL.
package wiring

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	wiredrepo "github.com/niflaot/pixels/internal/realm/room/database/wired"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/wired"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/bridge"
	conditionroom "github.com/niflaot/pixels/internal/realm/room/world/wired/condition/room"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	avatareffect "github.com/niflaot/pixels/internal/realm/room/world/wired/effect/avatar"
	boteffect "github.com/niflaot/pixels/internal/realm/room/world/wired/effect/bot"
	furnitureeffect "github.com/niflaot/pixels/internal/realm/room/world/wired/effect/furniture"
	progressioneffect "github.com/niflaot/pixels/internal/realm/room/world/wired/effect/progression"
	rewardeffect "github.com/niflaot/pixels/internal/realm/room/world/wired/effect/reward"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides WIRED persistence, compilation, execution, and protocol wiring.
var Module = fx.Module("room-wired",
	fx.Provide(
		roomwired.LoadConfig, NewRepository, NewStore, NewRewardStore, NewHighscoreStore, NewRegistry, NewCompiler,
		NewGamesWithStore, conditionroom.New, wiredruntime.NewRoomScheduler,
		NewAvatarEffects, NewFurnitureEffects, NewBotEffects, progressioneffect.New,
		rewardeffect.New, NewEffectsWithProgression, NewEngine, game.NewCoordinator, bridge.NewSpeechBridge, NewCommandHandler,
	),
	fx.Invoke(RegisterRuntime, RegisterHandlers),
)

// NewGames creates WIRED game state with room team-color projection.
func NewGames(rooms *roomlive.Registry, connections *netconn.Registry) *game.Service {
	return game.NewProjected(rooms, connections)
}

// NewGamesWithStore creates projected game state with durable highscore reset support.
func NewGamesWithStore(rooms *roomlive.Registry, connections *netconn.Registry, highscores record.HighscoreStore) *game.Service {
	resetter, _ := highscores.(record.HighscoreResetter)
	return game.NewProjected(rooms, connections, resetter)
}

// NewRepository creates WIRED PostgreSQL persistence.
func NewRepository(pool *postgres.Pool) *wiredrepo.Repository { return wiredrepo.New(pool) }

// NewStore exposes WIRED persistence through its domain contract.
func NewStore(repository *wiredrepo.Repository) record.Store { return repository }

// NewRewardStore exposes WIRED reward persistence through its focused contract.
func NewRewardStore(repository *wiredrepo.Repository) record.RewardStore { return repository }

// NewHighscoreStore exposes WIRED highscore persistence through its focused contract.
func NewHighscoreStore(repository *wiredrepo.Repository) record.HighscoreStore { return repository }

// NewRegistry creates and audits the canonical WIRED manifest.
func NewRegistry() (*registry.Registry, error) { return registry.Canonical() }

// NewCompiler creates the immutable WIRED configuration compiler.
func NewCompiler(registered *registry.Registry, config roomwired.Config) *configuration.Compiler {
	return configuration.NewCompiler(registered, config)
}

// NewAvatarEffects creates player-facing WIRED effects.
func NewAvatarEffects(rooms *roomlive.Registry, players *playerlive.Registry, connections *netconn.Registry, moderation *roommoderation.Service, achievements *playerachievement.Service) *avatareffect.Service {
	return avatareffect.New(rooms, players, connections, moderation, achievements)
}

// NewFurnitureEffects creates authoritative furniture WIRED effects.
func NewFurnitureEffects(rooms *roomlive.Registry, furniture *furnitureservice.Service, connections *netconn.Registry) *furnitureeffect.Service {
	return furnitureeffect.New(rooms, furniture, connections)
}

// NewBotEffects creates bot WIRED effects.
func NewBotEffects(rooms *roomlive.Registry, bots *botcore.Service) *boteffect.Service {
	return boteffect.New(rooms, bots)
}

// NewEffects composes focused effect services.
func NewEffects(furniture *furnitureeffect.Service, avatar *avatareffect.Service, bot *boteffect.Service, games *game.Service, rewards *rewardeffect.Service) *effect.Executor {
	return effect.New(effect.Services{Furniture: furniture, Avatar: avatar, Bot: bot, Game: games, Reward: rewards})
}

// NewEffectsWithProgression composes every focused effect service.
func NewEffectsWithProgression(furniture *furnitureeffect.Service, avatar *avatareffect.Service, bot *boteffect.Service, games *game.Service, rewards *rewardeffect.Service, progression *progressioneffect.Service) *effect.Executor {
	return effect.New(effect.Services{Furniture: furniture, Avatar: avatar, Bot: bot, Game: games, Reward: rewards, Progression: progression})
}

// NewEngine creates the room WIRED engine.
func NewEngine(config roomwired.Config, store record.Store, compiler *configuration.Compiler, effects *effect.Executor, views *conditionroom.Provider, scheduler *wiredruntime.RoomScheduler, furniture *furnitureeffect.Service) *wiredruntime.Engine {
	return wiredruntime.New(config, store, compiler, effects, views, scheduler, furniture)
}

// NewCommandHandler composes WIRED editor command behavior.
func NewCommandHandler(config roomwired.Config, players *playerlive.Registry, bindings *binding.Registry, rooms *roomlive.Registry, store record.Store, registered *registry.Registry, compiler *configuration.Compiler, engine *wiredruntime.Engine, permissions permissionservice.Checker, translations i18n.Translator) wiredcmd.Handler {
	return wiredcmd.Handler{Config: config, Players: players, Bindings: bindings, Rooms: rooms, Store: store, Registry: registered, Compiler: compiler, Engine: engine, Permissions: permissions, ConfigureAny: "room.wired.configure.any", Superwired: "room.wired.admin", Translations: translations}
}
