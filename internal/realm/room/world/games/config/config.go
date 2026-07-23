// Package config loads room-game mechanics policy.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config stores global room-game policy.
type Config struct {
	// Enabled controls all room-game engines.
	Enabled bool
	// Freeze stores Freeze mechanics policy.
	Freeze Freeze
	// Banzai stores Battle Banzai scoring.
	Banzai Banzai
}

// Freeze stores bounded Freeze mechanics values.
type Freeze struct {
	// PointsFreeze rewards freezing another player.
	PointsFreeze int
	// PointsBlock rewards breaking a block.
	PointsBlock int
	// PointsEffect rewards collecting a power-up.
	PointsEffect int
	// PowerupChance stores the percentage chance of a dropped power-up.
	PowerupChance int
	// MaxSnowballs bounds simultaneous ammunition.
	MaxSnowballs int
	// MaxLives bounds player lives.
	MaxLives int
	// LooseSnowballs stores ammunition lost when frozen.
	LooseSnowballs int
	// LooseBoost stores range lost when frozen.
	LooseBoost int
	// FrozenDuration controls immobilization.
	FrozenDuration time.Duration
	// ProtectionDuration controls shield duration.
	ProtectionDuration time.Duration
	// ProtectionStack allows repeated shield extensions.
	ProtectionStack bool
}

// Banzai stores tile scoring values.
type Banzai struct {
	// PointsSteal rewards taking an enemy tile.
	PointsSteal int
	// PointsFill rewards an enclosed-area tile.
	PointsFill int
	// PointsLock rewards locking a tile.
	PointsLock int
}

// Load reads every documented room-game environment setting.
func Load() Config {
	return Config{
		Enabled: envBool("PIXELS_GAMES_ENABLED", true),
		Freeze: Freeze{
			PointsFreeze: envInt("PIXELS_GAMES_FREEZE_POINTS_FREEZE", 10), PointsBlock: envInt("PIXELS_GAMES_FREEZE_POINTS_BLOCK", 1),
			PointsEffect: envInt("PIXELS_GAMES_FREEZE_POINTS_EFFECT", 3), PowerupChance: envInt("PIXELS_GAMES_FREEZE_POWERUP_CHANCE", 33),
			MaxSnowballs: envInt("PIXELS_GAMES_FREEZE_MAX_SNOWBALLS", 5), MaxLives: envInt("PIXELS_GAMES_FREEZE_MAX_LIVES", 3),
			LooseSnowballs: envInt("PIXELS_GAMES_FREEZE_LOOSE_SNOWBALLS", 5), LooseBoost: envInt("PIXELS_GAMES_FREEZE_LOOSE_BOOST", 3),
			FrozenDuration:     time.Duration(envInt("PIXELS_GAMES_FREEZE_FROZEN_SECONDS", 5)) * time.Second,
			ProtectionDuration: time.Duration(envInt("PIXELS_GAMES_FREEZE_PROTECTION_SECONDS", 10)) * time.Second,
			ProtectionStack:    envBool("PIXELS_GAMES_FREEZE_PROTECTION_STACK", true),
		},
		Banzai: Banzai{PointsSteal: envNonNegative("PIXELS_GAMES_BANZAI_POINTS_STEAL", 0), PointsFill: envNonNegative("PIXELS_GAMES_BANZAI_POINTS_FILL", 0), PointsLock: envInt("PIXELS_GAMES_BANZAI_POINTS_LOCK", 1)},
	}
}

// envBool reads one boolean or its fallback.
func envBool(name string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	if err != nil {
		return fallback
	}
	return value
}

// envInt reads one positive integer or its fallback.
func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// envNonNegative reads one non-negative integer or its fallback.
func envNonNegative(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}
