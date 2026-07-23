// Package config loads social-group operational policy.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config contains social-group limits and operational policy.
type Config struct {
	// CreationCost stores the credits charged for one group.
	CreationCost int64
	// RequireClub requires an active HC entitlement for creation.
	RequireClub bool
	// OwnedLimit caps active groups owned by one player.
	OwnedLimit int
	// MembershipLimit caps active memberships held by one player.
	MembershipLimit int
	// MemberLimit caps active members in one group.
	MemberLimit int
	// PendingLimit caps pending requests in one group.
	PendingLimit int
	// MemberPageSize stores Nitro's fixed member page size.
	MemberPageSize int
	// MaxSearchLength bounds member search input before allocation.
	MaxSearchLength int
	// BulkApproveLimit bounds one approve-all transaction.
	BulkApproveLimit int
	// FurnitureCleanupLimit bounds one atomic HQ return operation.
	FurnitureCleanupLimit int
	// ForumPageSize caps forum list and message requests.
	ForumPageSize int
	// ForumSubjectLimit caps thread subjects in Unicode runes.
	ForumSubjectLimit int
	// ForumMessageLimit caps post bodies in Unicode runes.
	ForumMessageLimit int
	// ForumPostCooldown throttles forum writes per player.
	ForumPostCooldown time.Duration
	// ForumActiveWindow controls the active-forum ranking window.
	ForumActiveWindow time.Duration
	// CacheTTL is a safety net for missed invalidation events.
	CacheTTL time.Duration
	// ForumCursorTTL bounds header-only forum CFH context.
	ForumCursorTTL time.Duration
	// DeactivationRetention controls retained audit/content lifetime.
	DeactivationRetention time.Duration
}

// LoadConfig loads social-group configuration with safe defaults.
func LoadConfig() Config {
	return Config{
		CreationCost: envInt64("PIXELS_GROUP_CREATION_COST", 10),
		RequireClub:  envBool("PIXELS_GROUP_REQUIRE_CLUB", true), OwnedLimit: envInt("PIXELS_GROUP_OWNED_LIMIT", 100),
		MembershipLimit: envInt("PIXELS_GROUP_MEMBERSHIP_LIMIT", 100), MemberLimit: envInt("PIXELS_GROUP_MEMBER_LIMIT", 50000),
		PendingLimit: envInt("PIXELS_GROUP_PENDING_LIMIT", 100), MemberPageSize: envInt("PIXELS_GROUP_MEMBER_PAGE_SIZE", 14),
		MaxSearchLength: envInt("PIXELS_GROUP_MAX_SEARCH_LENGTH", 64), BulkApproveLimit: envInt("PIXELS_GROUP_BULK_APPROVE_LIMIT", 100),
		FurnitureCleanupLimit: envInt("PIXELS_GROUP_FURNITURE_CLEANUP_LIMIT", 1000), ForumPageSize: envInt("PIXELS_GROUP_FORUM_PAGE_SIZE", 50),
		ForumSubjectLimit: envInt("PIXELS_GROUP_FORUM_SUBJECT_LIMIT", 120), ForumMessageLimit: envInt("PIXELS_GROUP_FORUM_MESSAGE_LIMIT", 4000),
		ForumPostCooldown: envDuration("PIXELS_GROUP_FORUM_POST_COOLDOWN", 3*time.Second), ForumActiveWindow: envDuration("PIXELS_GROUP_FORUM_ACTIVE_WINDOW", 7*24*time.Hour),
		CacheTTL: envDuration("PIXELS_GROUP_CACHE_TTL", 10*time.Minute), ForumCursorTTL: envDuration("PIXELS_GROUP_FORUM_CURSOR_TTL", 5*time.Minute),
		DeactivationRetention: envDuration("PIXELS_GROUP_DEACTIVATION_RETENTION", 365*24*time.Hour),
	}
}

// envBool reads one boolean environment setting.
func envBool(name string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	if err != nil {
		return fallback
	}
	return value
}

// envInt reads one integer environment setting.
func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// envInt64 reads one non-negative integer environment setting.
func envInt64(name string, fallback int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

// envDuration reads one positive duration environment setting.
func envDuration(name string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
