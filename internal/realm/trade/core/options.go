// Package core implements direct-trade workflows.
package core

import (
	"errors"
	"time"

	"github.com/niflaot/pixels/internal/permission"
)

var (
	// BypassRestrictions bypasses room and global trade restrictions.
	BypassRestrictions = permission.RegisterNode("trade.bypass_restrictions", "")
	// ModerationLock permits changing a player's durable trade lock.
	ModerationLock = permission.RegisterNode("trade.moderation.lock", "")
	// MarketplaceAdmin permits force-closing Marketplace listings.
	MarketplaceAdmin = permission.RegisterNode("marketplace.admin.manage", "")
	// ErrDisabled reports globally disabled direct trading.
	ErrDisabled = errors.New("trade disabled")
	// ErrThrottled reports a repeated trade start within the configured window.
	ErrThrottled = errors.New("trade start throttled")
	// ErrUnavailable reports an unavailable participant or session.
	ErrUnavailable = errors.New("trade participant unavailable")
	// ErrActorNotAllowed reports a trade-locked initiating player.
	ErrActorNotAllowed = errors.New("trade actor not allowed")
	// ErrTargetNotAllowed reports a trade-locked target player.
	ErrTargetNotAllowed = errors.New("trade target not allowed")
	// ErrRoomPolicy reports room settings that reject the trade.
	ErrRoomPolicy = errors.New("room trade policy denied")
	// ErrIgnored reports that the target ignores the initiator.
	ErrIgnored = errors.New("trade initiator ignored")
	// ErrItemUnavailable reports an invalid, locked, or untradeable item.
	ErrItemUnavailable = errors.New("trade item unavailable")
	// ErrNotAccepted reports confirmation before both sides accept.
	ErrNotAccepted = errors.New("trade not accepted")
	// ErrMaximumItems reports an offer above the configured cap.
	ErrMaximumItems = errors.New("trade maximum items reached")
	// ErrAccepted reports item removal before revoking acceptance.
	ErrAccepted = errors.New("trade offer already accepted")
)

// Options controls direct trading.
type Options struct {
	// Enabled reports whether direct trading is globally available.
	Enabled bool
	// StartThrottle stores the minimum interval between trade starts.
	StartThrottle time.Duration
	// MaximumItems stores the per-participant offer cap.
	MaximumItems int
	// AuditEnabled reports whether completed settlements are persisted.
	AuditEnabled bool
}
