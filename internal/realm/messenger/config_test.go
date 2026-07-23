package messenger

import (
	"testing"
	"time"
)

// TestNormalizeUsesDefaults verifies messenger defensive defaults.
func TestNormalizeUsesDefaults(t *testing.T) {
	config := (Config{}).Normalize()
	if config.MaxFriends != 200 || config.MaxFriendsClub != 500 || config.SearchCacheTTL != 30*time.Second || config.FriendCacheTTL != 30*time.Second || config.ChatThrottle != 750*time.Millisecond {
		t.Fatalf("unexpected config %#v", config)
	}
}

// TestLoadConfigUsesEnvironment verifies representative overrides.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_MESSENGER_MAX_FRIENDS", "300")
	t.Setenv("PIXELS_MESSENGER_SEARCH_CACHE_TTL", "45s")
	t.Setenv("PIXELS_MESSENGER_FRIEND_CACHE_TTL", "20s")
	t.Setenv("PIXELS_MESSENGER_CHAT_LOG_ENABLED", "true")
	config, err := LoadConfig()
	if err != nil || config.MaxFriends != 300 || config.SearchCacheTTL != 45*time.Second || config.FriendCacheTTL != 20*time.Second || !config.ChatLogEnabled {
		t.Fatalf("unexpected config %#v err=%v", config, err)
	}
}
