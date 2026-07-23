// Package record defines durable global sanction contracts.
package record

import "time"

// Kind identifies one global punishment behavior.
type Kind string

const (
	// MaxDurationHours bounds finite punishments to ten years.
	MaxDurationHours int32 = 24 * 365 * 10
	// KindBan prevents authentication and disconnects an active session.
	KindBan Kind = "ban"
	// KindMute prevents hotel chat while active.
	KindMute Kind = "mute"
	// KindWarn delivers one moderator warning.
	KindWarn Kind = "warn"
	// KindTradeLock prevents direct trading while active.
	KindTradeLock Kind = "trade_lock"
	// KindKick removes an online player from the hotel.
	KindKick Kind = "kick"
)

// Valid reports whether the kind is supported by the sanction engine.
func (kind Kind) Valid() bool {
	switch kind {
	case KindBan, KindMute, KindWarn, KindTradeLock, KindKick:
		return true
	default:
		return false
	}
}

// Instant reports whether the punishment has no persistent active effect.
func (kind Kind) Instant() bool { return kind == KindWarn || kind == KindKick }

// Punishment stores one immutable sanction plus optional revocation state.
type Punishment struct {
	// ID identifies the punishment.
	ID int64
	// ReceiverPlayerID identifies the sanctioned player.
	ReceiverPlayerID int64
	// IssuerPlayerID identifies the moderator when IssuerKind is player.
	IssuerPlayerID *int64
	// IssuerKind identifies player or system issuance.
	IssuerKind string
	// Kind selects the registered effect.
	Kind Kind
	// Reason stores the moderator-visible reason.
	Reason string
	// CFHTopicID optionally links the originating topic.
	CFHTopicID *int64
	// IssueID optionally links the originating issue.
	IssueID *int64
	// Source identifies the issuing workflow.
	Source string
	// IssuedAt stores creation time.
	IssuedAt time.Time
	// ExpiresAt optionally stores active-effect expiry.
	ExpiresAt *time.Time
	// RevokedAt optionally stores early revocation time.
	RevokedAt *time.Time
	// RevokedByPlayerID optionally identifies the revoking moderator.
	RevokedByPlayerID *int64
}

// ActiveAt reports whether the punishment has an active persistent effect.
func (punishment Punishment) ActiveAt(now time.Time) bool {
	return !punishment.Kind.Instant() && punishment.RevokedAt == nil && (punishment.ExpiresAt == nil || punishment.ExpiresAt.After(now))
}

// ApplyParams describes one sanction request.
type ApplyParams struct {
	// ReceiverPlayerID identifies the target.
	ReceiverPlayerID int64
	// IssuerPlayerID identifies a player issuer.
	IssuerPlayerID *int64
	// IssuerKind identifies player or system issuance.
	IssuerKind string
	// Kind selects the punishment behavior.
	Kind Kind
	// Reason stores the required explanation.
	Reason string
	// CFHTopicID links an optional help topic.
	CFHTopicID *int64
	// IssueID links an optional moderation issue.
	IssueID *int64
	// Source identifies the workflow.
	Source string
	// ExpiresAt optionally bounds the punishment.
	ExpiresAt *time.Time
}

// ActiveState summarizes hot-path sanctions for one player.
type ActiveState struct {
	// Ban stores the newest active ban.
	Ban *Punishment
	// MuteUntil stores the latest finite mute or nil for no finite mute.
	MuteUntil *time.Time
	// MutedPermanently reports an active permanent mute.
	MutedPermanently bool
	// TradeLockUntil stores the latest finite trade lock.
	TradeLockUntil *time.Time
	// TradeLockedPermanently reports an active permanent trade lock.
	TradeLockedPermanently bool
}

// LadderEntry configures one escalation level.
type LadderEntry struct {
	// Level identifies escalation order.
	Level int32
	// Kind selects the resulting punishment.
	Kind Kind
	// DurationHours stores zero for permanent punishment.
	DurationHours int32
	// ProbationDays stores the repeat window.
	ProbationDays int32
}

// Alert stores one pending offline warning.
type Alert struct {
	// ID identifies the pending alert.
	ID int64
	// PlayerID identifies its recipient.
	PlayerID int64
	// PunishmentID optionally links its source.
	PunishmentID *int64
	// Message stores already-localized visible text.
	Message string
}
