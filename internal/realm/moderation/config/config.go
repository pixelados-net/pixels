// Package config loads global moderation runtime policy.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config stores global moderation runtime policy.
type Config struct {
	// Enabled controls call-for-help availability.
	Enabled bool
	// ReportLimit bounds reports in one throttle window.
	ReportLimit int64
	// ReportWindow stores the distributed throttle duration.
	ReportWindow time.Duration
	// ContextWindow bounds frozen history entries.
	ContextWindow int
	// GuardianCount stores reviewers offered per ticket.
	GuardianCount int
	// GuardianVoteWindow bounds peer review voting.
	GuardianVoteWindow time.Duration
	// GuardianIgnoreLimit stores ignored offers before exclusion.
	GuardianIgnoreLimit int
	// GuardianExclusion stores ignored-offer exclusion duration.
	GuardianExclusion time.Duration
}

// Load reads moderation environment settings with safe defaults.
func Load() Config {
	return Config{Enabled: envBool("PIXELS_MODERATION_ENABLED", true), ReportLimit: int64(envInt("PIXELS_MODERATION_REPORT_LIMIT", 3)), ReportWindow: envDuration("PIXELS_MODERATION_REPORT_WINDOW", 10*time.Minute), ContextWindow: envInt("PIXELS_MODERATION_CONTEXT_WINDOW", 50), GuardianCount: envInt("PIXELS_GUARDIAN_COUNT", 3), GuardianVoteWindow: envDuration("PIXELS_GUARDIAN_VOTE_WINDOW", time.Minute), GuardianIgnoreLimit: envInt("PIXELS_GUARDIAN_IGNORE_LIMIT", 3), GuardianExclusion: envDuration("PIXELS_GUARDIAN_EXCLUSION", 30*time.Minute)}
}

// envBool reads one boolean environment value.
func envBool(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

// envInt reads one integer environment value.
func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// envDuration reads one duration environment value.
func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
