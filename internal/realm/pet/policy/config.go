// Package policy owns pet limits, timing, and permission capabilities.
package policy

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config stores bounded pet realm settings.
type Config struct {
	// Enabled reports whether pet creation and room behavior are available.
	Enabled bool `env:"PIXELS_PET_ENABLED" envDefault:"true"`
	// MaxPerRoom stores the ordinary room pet limit.
	MaxPerRoom int `env:"PIXELS_PET_MAX_PER_ROOM" envDefault:"25"`
	// MaxPerOwnerRoom stores one owner's room pet limit.
	MaxPerOwnerRoom int `env:"PIXELS_PET_MAX_PER_OWNER_ROOM" envDefault:"10"`
	// MaxInventory stores the ordinary per-player inventory limit.
	MaxInventory int `env:"PIXELS_PET_MAX_INVENTORY" envDefault:"25"`
	// InventoryFragmentSize stores the maximum pets per USER_PETS fragment.
	InventoryFragmentSize int `env:"PIXELS_PET_INVENTORY_FRAGMENT_SIZE" envDefault:"100"`
	// WalkRadius stores the autonomous random-walk radius.
	WalkRadius int `env:"PIXELS_PET_WALK_RADIUS" envDefault:"5"`
	// DecisionMinimum stores the minimum autonomous decision interval.
	DecisionMinimum time.Duration `env:"PIXELS_PET_DECISION_MINIMUM" envDefault:"5s"`
	// DecisionMaximum stores the bounded autonomous jitter ceiling.
	DecisionMaximum time.Duration `env:"PIXELS_PET_DECISION_MAXIMUM" envDefault:"15s"`
	// PositionFlushInterval stores coalesced position persistence cadence.
	PositionFlushInterval time.Duration `env:"PIXELS_PET_POSITION_FLUSH_INTERVAL" envDefault:"5s"`
	// StatDecayInterval stores the durable energy and happiness materialization cadence.
	StatDecayInterval time.Duration `env:"PIXELS_PET_STAT_DECAY_INTERVAL" envDefault:"30m"`
	// EnergyDecay stores energy removed per elapsed materialization interval.
	EnergyDecay int32 `env:"PIXELS_PET_ENERGY_DECAY" envDefault:"1"`
	// HappinessDecay stores happiness removed per elapsed materialization interval.
	HappinessDecay int32 `env:"PIXELS_PET_HAPPINESS_DECAY" envDefault:"1"`
	// RespectMinimumAge stores required age before respect.
	RespectMinimumAge time.Duration `env:"PIXELS_PET_RESPECT_MINIMUM_AGE" envDefault:"72h"`
	// RespectExperience stores experience awarded by respect.
	RespectExperience int32 `env:"PIXELS_PET_RESPECT_EXPERIENCE" envDefault:"10"`
	// RespectDailyLimit stores the ordinary player-wide respect budget.
	RespectDailyLimit int `env:"PIXELS_PET_RESPECT_DAILY_LIMIT" envDefault:"3"`
	// AllowRespectOwn reports whether owners may respect their pets.
	AllowRespectOwn bool `env:"PIXELS_PET_ALLOW_RESPECT_OWN" envDefault:"false"`
	// PlantRewardDefinitionID identifies the furniture seed returned by harvest.
	PlantRewardDefinitionID int64 `env:"PIXELS_PET_PLANT_REWARD_DEFINITION_ID" envDefault:"4582"`
	// PlantCompostDefinitionID identifies the RIP furniture created from a dead plant.
	PlantCompostDefinitionID int64 `env:"PIXELS_PET_PLANT_COMPOST_DEFINITION_ID" envDefault:"4830"`
	// PlantGrowDuration stores the absolute seed-to-maturity interval.
	PlantGrowDuration time.Duration `env:"PIXELS_PET_PLANT_GROW_DURATION" envDefault:"168h"`
	// PlantLifeDuration stores the mature plant lifetime before death.
	PlantLifeDuration time.Duration `env:"PIXELS_PET_PLANT_LIFE_DURATION" envDefault:"168h"`
	// PackageTimeout bounds a package naming handshake.
	PackageTimeout time.Duration `env:"PIXELS_PET_PACKAGE_TIMEOUT" envDefault:"2m"`
	// BreedingTimeout bounds one durable nest session.
	BreedingTimeout time.Duration `env:"PIXELS_PET_BREEDING_TIMEOUT" envDefault:"2m"`
	// BreedingMinimumAge stores the adult parent age gate.
	BreedingMinimumAge time.Duration `env:"PIXELS_PET_BREEDING_MINIMUM_AGE" envDefault:"72h"`
	// UnloadFlushTimeout bounds room-close persistence draining.
	UnloadFlushTimeout time.Duration `env:"PIXELS_PET_UNLOAD_FLUSH_TIMEOUT" envDefault:"3s"`
}

// LoadConfig loads pet settings from environment variables.
func LoadConfig() (Config, error) {
	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}
	return config.Normalize(), nil
}

// Normalize clamps unsafe pet configuration values.
func (config Config) Normalize() Config {
	if config.MaxPerRoom <= 0 {
		config.MaxPerRoom = 25
	}
	if config.MaxPerOwnerRoom <= 0 || config.MaxPerOwnerRoom > config.MaxPerRoom {
		config.MaxPerOwnerRoom = 10
	}
	if config.MaxInventory <= 0 {
		config.MaxInventory = 25
	}
	if config.InventoryFragmentSize <= 0 || config.InventoryFragmentSize > 500 {
		config.InventoryFragmentSize = 100
	}
	if config.WalkRadius <= 0 || config.WalkRadius > 20 {
		config.WalkRadius = 5
	}
	if config.DecisionMinimum < time.Second {
		config.DecisionMinimum = 5 * time.Second
	}
	if config.DecisionMaximum < config.DecisionMinimum {
		config.DecisionMaximum = 15 * time.Second
	}
	if config.PositionFlushInterval < time.Second {
		config.PositionFlushInterval = 5 * time.Second
	}
	if config.StatDecayInterval < time.Minute {
		config.StatDecayInterval = 30 * time.Minute
	}
	if config.EnergyDecay <= 0 {
		config.EnergyDecay = 1
	}
	if config.HappinessDecay <= 0 {
		config.HappinessDecay = 1
	}
	if config.RespectMinimumAge < 0 {
		config.RespectMinimumAge = 72 * time.Hour
	}
	if config.RespectExperience <= 0 {
		config.RespectExperience = 10
	}
	if config.RespectDailyLimit <= 0 {
		config.RespectDailyLimit = 3
	}
	if config.PlantRewardDefinitionID <= 0 {
		config.PlantRewardDefinitionID = 4582
	}
	if config.PlantCompostDefinitionID <= 0 {
		config.PlantCompostDefinitionID = 4830
	}
	if config.PlantGrowDuration <= 0 {
		config.PlantGrowDuration = 7 * 24 * time.Hour
	}
	if config.PlantLifeDuration <= 0 {
		config.PlantLifeDuration = 7 * 24 * time.Hour
	}
	if config.PackageTimeout <= 0 {
		config.PackageTimeout = 2 * time.Minute
	}
	if config.BreedingTimeout <= 0 {
		config.BreedingTimeout = 2 * time.Minute
	}
	if config.BreedingMinimumAge < 0 {
		config.BreedingMinimumAge = 72 * time.Hour
	}
	if config.UnloadFlushTimeout <= 0 {
		config.UnloadFlushTimeout = 3 * time.Second
	}
	return config
}
