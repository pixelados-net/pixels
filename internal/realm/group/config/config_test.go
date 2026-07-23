package config

import (
	"testing"
	"time"
)

// TestLoadConfigUsesDefaultsAndValidOverrides verifies every environment parser class.
func TestLoadConfigUsesDefaultsAndValidOverrides(t *testing.T) {
	t.Setenv("PIXELS_GROUP_CREATION_COST", "25")
	t.Setenv("PIXELS_GROUP_MEMBER_LIMIT", "123")
	t.Setenv("PIXELS_GROUP_FORUM_POST_COOLDOWN", "250ms")
	config := LoadConfig()
	if config.CreationCost != 25 || config.MemberLimit != 123 || config.ForumPostCooldown != 250*time.Millisecond {
		t.Fatalf("config=%#v", config)
	}
}

// TestLoadConfigRejectsUnsafeOverrides verifies invalid values retain safe defaults.
func TestLoadConfigRejectsUnsafeOverrides(t *testing.T) {
	t.Setenv("PIXELS_GROUP_CREATION_COST", "-1")
	t.Setenv("PIXELS_GROUP_MEMBER_LIMIT", "0")
	t.Setenv("PIXELS_GROUP_FORUM_POST_COOLDOWN", "never")
	config := LoadConfig()
	if config.CreationCost != 10 || config.MemberLimit != 50000 || config.ForumPostCooldown != 3*time.Second {
		t.Fatalf("config=%#v", config)
	}
}
