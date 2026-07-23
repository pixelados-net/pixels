package recipe

import (
	"context"
	"sync"

	craftingconfig "github.com/niflaot/pixels/internal/realm/crafting/config"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	"github.com/niflaot/pixels/pkg/bus"
)

// AchievementGranter grants optional post-commit crafting badges.
type AchievementGranter interface {
	GrantBadge(context.Context, int64, string, string) (bool, error)
}

// Service coordinates altar sessions and atomic crafting.
type Service struct {
	// config stores crafting runtime policy.
	config craftingconfig.Config
	// store persists recipes and discoveries.
	store craftingrecord.Store
	// furniture validates, consumes, and grants inventory items.
	furniture furnitureservice.TradingManager
	// granter creates reward items in the shared transaction.
	granter furnitureservice.Granter
	// achievements receives optional non-blocking hooks.
	achievements AchievementGranter
	// events publishes committed crafting outcomes.
	events bus.Publisher
	// mutex protects current altar sessions.
	mutex sync.RWMutex
	// altars stores the most recently opened valid altar per player.
	altars map[int64]int64
}

// Result stores one committed craft projection.
type Result struct {
	// Recipe stores the crafted recipe.
	Recipe craftingrecord.Recipe
	// Removed stores consumed inventory item identifiers.
	Removed []int64
	// Granted stores the created reward item.
	Granted furnituremodel.Item
	// Definition stores the reward definition.
	Definition furnituremodel.Definition
	// Discovered reports a first secret discovery.
	Discovered bool
	// Exhausted reports that limited stock reached zero.
	Exhausted bool
}

// New creates a crafting recipe service.
func New(config craftingconfig.Config, store craftingrecord.Store, furniture furnitureservice.TradingManager, granter furnitureservice.Granter, achievements *playerachievement.Service, events bus.Publisher) *Service {
	return &Service{config: config, store: store, furniture: furniture, granter: granter, achievements: achievements, events: events, altars: make(map[int64]int64)}
}
