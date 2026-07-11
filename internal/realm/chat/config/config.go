// Package config controls protocol-backed room communication.
package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	// defaultMaxMessageRunes stores the normal maximum message length.
	defaultMaxMessageRunes = 256
	// defaultAudienceDistance stores the fallback talk radius.
	defaultAudienceDistance = 50
)

// Config controls chat validation, flood protection, and history buffering.
type Config struct {
	// MaxMessageRunes stores the normal Unicode message limit.
	MaxMessageRunes int `env:"PIXELS_CHAT_MAX_MESSAGE_RUNES" envDefault:"256"`
	// Tier0MaxMessages stores the loose room burst budget.
	Tier0MaxMessages int64 `env:"PIXELS_CHAT_FLOOD_TIER0_MAX_MESSAGES" envDefault:"100"`
	// Tier0Window stores the loose room flood window.
	Tier0Window time.Duration `env:"PIXELS_CHAT_FLOOD_TIER0_WINDOW" envDefault:"1s"`
	// Tier1MaxMessages stores the normal room burst budget.
	Tier1MaxMessages int64 `env:"PIXELS_CHAT_FLOOD_TIER1_MAX_MESSAGES" envDefault:"10"`
	// Tier1Window stores the normal room flood window.
	Tier1Window time.Duration `env:"PIXELS_CHAT_FLOOD_TIER1_WINDOW" envDefault:"5s"`
	// Tier2MaxMessages stores the strict room burst budget.
	Tier2MaxMessages int64 `env:"PIXELS_CHAT_FLOOD_TIER2_MAX_MESSAGES" envDefault:"6"`
	// Tier2Window stores the strict room flood window.
	Tier2Window time.Duration `env:"PIXELS_CHAT_FLOOD_TIER2_WINDOW" envDefault:"5s"`
	// LogWhispers reports whether private whispers enter durable history.
	LogWhispers bool `env:"PIXELS_CHAT_LOG_WHISPERS" envDefault:"false"`
	// HistoryRetentionDays stores how many daily partitions remain available.
	HistoryRetentionDays int `env:"PIXELS_CHAT_LOG_RETENTION_DAYS" envDefault:"14"`
	// HistoryFlushInterval stores the maximum history write delay.
	HistoryFlushInterval time.Duration `env:"PIXELS_CHAT_LOG_FLUSH_INTERVAL" envDefault:"2s"`
	// HistoryBatchSize stores the maximum rows written per batch.
	HistoryBatchSize int `env:"PIXELS_CHAT_LOG_FLUSH_BATCH_SIZE" envDefault:"200"`
	// HistoryQueueSize stores the non-blocking history queue capacity.
	HistoryQueueSize int `env:"PIXELS_CHAT_LOG_QUEUE_SIZE" envDefault:"4096"`
}

// Tier describes one flood-control budget.
type Tier struct {
	// MaxMessages stores the accepted burst count.
	MaxMessages int64
	// Window stores the counter expiration window.
	Window time.Duration
}

// LoadConfig reads chat configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid values with conservative defaults.
func (config Config) Normalize() Config {
	if config.MaxMessageRunes <= 0 {
		config.MaxMessageRunes = defaultMaxMessageRunes
	}
	config.Tier0MaxMessages, config.Tier0Window = normalizeTier(config.Tier0MaxMessages, config.Tier0Window, 100, time.Second)
	config.Tier1MaxMessages, config.Tier1Window = normalizeTier(config.Tier1MaxMessages, config.Tier1Window, 10, 5*time.Second)
	config.Tier2MaxMessages, config.Tier2Window = normalizeTier(config.Tier2MaxMessages, config.Tier2Window, 6, 5*time.Second)
	if config.HistoryRetentionDays <= 0 {
		config.HistoryRetentionDays = 14
	}
	if config.HistoryFlushInterval <= 0 {
		config.HistoryFlushInterval = 2 * time.Second
	}
	if config.HistoryBatchSize <= 0 {
		config.HistoryBatchSize = 200
	}
	if config.HistoryQueueSize < config.HistoryBatchSize {
		config.HistoryQueueSize = 4096
	}

	return config
}

// Tier returns the normalized flood budget for one Nitro protection value.
func (config Config) Tier(protection int16) Tier {
	config = config.Normalize()
	switch protection {
	case 0:
		return Tier{MaxMessages: config.Tier0MaxMessages, Window: config.Tier0Window}
	case 1:
		return Tier{MaxMessages: config.Tier1MaxMessages, Window: config.Tier1Window}
	default:
		return Tier{MaxMessages: config.Tier2MaxMessages, Window: config.Tier2Window}
	}
}

// AudienceDistance returns a valid room talk radius.
func AudienceDistance(distance int16) int64 {
	if distance <= 0 {
		return defaultAudienceDistance
	}

	return int64(distance)
}

// normalizeTier fills one invalid flood budget.
func normalizeTier(maxMessages int64, window time.Duration, fallbackMessages int64, fallbackWindow time.Duration) (int64, time.Duration) {
	if maxMessages <= 0 {
		maxMessages = fallbackMessages
	}
	if window <= 0 {
		window = fallbackWindow
	}

	return maxMessages, window
}
