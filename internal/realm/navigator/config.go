package navigator

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains bounded Navigator behavior settings.
type Config struct {
	// SearchLimit bounds ordinary room search results.
	SearchLimit int `env:"PIXELS_NAVIGATOR_SEARCH_LIMIT" envDefault:"50"`
	// HistoryLimit bounds recent and frequent room results.
	HistoryLimit int `env:"PIXELS_NAVIGATOR_HISTORY_LIMIT" envDefault:"25"`
	// FavoriteLimit bounds ordinary player favorites.
	FavoriteLimit int32 `env:"PIXELS_NAVIGATOR_FAVORITE_LIMIT" envDefault:"30"`
	// HistoryQueueSize bounds asynchronous admission telemetry.
	HistoryQueueSize int `env:"PIXELS_NAVIGATOR_HISTORY_QUEUE_SIZE" envDefault:"1024"`
	// HistoryDedupeWindow prevents rapid reentries from inflating visit counts.
	HistoryDedupeWindow time.Duration `env:"PIXELS_NAVIGATOR_HISTORY_DEDUPE_WINDOW" envDefault:"30s"`
	// PreferenceFlushInterval coalesces repeated Navigator resize writes.
	PreferenceFlushInterval time.Duration `env:"PIXELS_NAVIGATOR_PREFERENCE_FLUSH_INTERVAL" envDefault:"250ms"`
	// PreferencePendingLimit bounds players waiting for a preference flush.
	PreferencePendingLimit int `env:"PIXELS_NAVIGATOR_PREFERENCE_PENDING_LIMIT" envDefault:"4096"`
	// WindowPositionLimit bounds absolute multi-monitor coordinates.
	WindowPositionLimit int `env:"PIXELS_NAVIGATOR_WINDOW_POSITION_LIMIT" envDefault:"32768"`
	// WindowMinimumWidth bounds saved Navigator width.
	WindowMinimumWidth int `env:"PIXELS_NAVIGATOR_WINDOW_MIN_WIDTH" envDefault:"320"`
	// WindowMaximumWidth bounds saved Navigator width.
	WindowMaximumWidth int `env:"PIXELS_NAVIGATOR_WINDOW_MAX_WIDTH" envDefault:"4096"`
	// WindowMinimumHeight bounds saved Navigator height.
	WindowMinimumHeight int `env:"PIXELS_NAVIGATOR_WINDOW_MIN_HEIGHT" envDefault:"240"`
	// WindowMaximumHeight bounds saved Navigator height.
	WindowMaximumHeight int `env:"PIXELS_NAVIGATOR_WINDOW_MAX_HEIGHT" envDefault:"2160"`
}

// LoadConfig loads Navigator policy from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// DefaultConfig returns the documented Navigator defaults.
func DefaultConfig() Config {
	return Config{SearchLimit: 50, HistoryLimit: 25, FavoriteLimit: 30, HistoryQueueSize: 1024, HistoryDedupeWindow: 30 * time.Second, PreferenceFlushInterval: 250 * time.Millisecond, PreferencePendingLimit: 4096, WindowPositionLimit: 32768, WindowMinimumWidth: 320, WindowMaximumWidth: 4096, WindowMinimumHeight: 240, WindowMaximumHeight: 2160}
}
