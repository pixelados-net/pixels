package achievement

import "context"

// Store persists player badges and idempotent respect grants.
type Store interface {
	// Badges lists a player's durable badge snapshot.
	Badges(context.Context, int64) ([]Badge, error)
	// GrantBadge grants one badge idempotently.
	GrantBadge(context.Context, int64, string, string) (bool, error)
	// ReplaceBadge replaces one badge code while preserving its equipped slot.
	ReplaceBadge(context.Context, int64, string, string, string) (bool, error)
	// RemoveBadge removes one badge regardless of equipped state.
	RemoveBadge(context.Context, int64, string) (bool, error)
	// SetEquipped atomically replaces one player's active badge slots.
	SetEquipped(context.Context, int64, []string) error
	// GrantRespect applies one idempotent positive respect grant.
	GrantRespect(context.Context, int64, int32, string, string) (bool, error)
}
