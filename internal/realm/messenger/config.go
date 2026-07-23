// Package messenger owns durable and live social communication behavior.
package messenger

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls messenger limits, caches, and private chat behavior.
type Config struct {
	// MaxFriends stores the normal friend-list capacity.
	MaxFriends int `env:"PIXELS_MESSENGER_MAX_FRIENDS" envDefault:"200"`
	// MaxFriendsClub stores the active club friend-list capacity.
	MaxFriendsClub int `env:"PIXELS_MESSENGER_MAX_FRIENDS_CLUB" envDefault:"500"`
	// MaxSearchResults stores the maximum prefix-search result count.
	MaxSearchResults int `env:"PIXELS_MESSENGER_SEARCH_MAX_RESULTS" envDefault:"50"`
	// SearchCacheTTL stores shared search-result cache lifetime.
	SearchCacheTTL time.Duration `env:"PIXELS_MESSENGER_SEARCH_CACHE_TTL" envDefault:"30s"`
	// FriendCacheTTL stores shared durable friend-card cache lifetime.
	FriendCacheTTL time.Duration `env:"PIXELS_MESSENGER_FRIEND_CACHE_TTL" envDefault:"30s"`
	// SearchThrottle stores the per-player search interval.
	SearchThrottle time.Duration `env:"PIXELS_MESSENGER_SEARCH_THROTTLE" envDefault:"3s"`
	// ChatThrottle stores the per-sender private-message interval.
	ChatThrottle time.Duration `env:"PIXELS_MESSENGER_CHAT_THROTTLE" envDefault:"750ms"`
	// ChatFilterEnabled reports whether the global word filter applies to private messages.
	ChatFilterEnabled bool `env:"PIXELS_MESSENGER_CHAT_FILTER_ENABLED" envDefault:"false"`
	// ChatLogEnabled reports whether accepted private messages are persisted.
	ChatLogEnabled bool `env:"PIXELS_MESSENGER_CHAT_LOG_ENABLED" envDefault:"false"`
}

// LoadConfig reads messenger configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid messenger limits with conservative defaults.
func (config Config) Normalize() Config {
	if config.MaxFriends <= 0 {
		config.MaxFriends = 200
	}
	if config.MaxFriendsClub < config.MaxFriends {
		config.MaxFriendsClub = 500
	}
	if config.MaxSearchResults <= 0 {
		config.MaxSearchResults = 50
	}
	if config.SearchCacheTTL <= 0 {
		config.SearchCacheTTL = 30 * time.Second
	}
	if config.FriendCacheTTL <= 0 {
		config.FriendCacheTTL = 30 * time.Second
	}
	if config.SearchThrottle <= 0 {
		config.SearchThrottle = 3 * time.Second
	}
	if config.ChatThrottle <= 0 {
		config.ChatThrottle = 750 * time.Millisecond
	}

	return config
}
