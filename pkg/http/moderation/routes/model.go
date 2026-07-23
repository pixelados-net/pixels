// Package routes exposes protected global moderation administration.
package routes

import "time"

// ApplyPunishmentRequest describes one global punishment mutation.
type ApplyPunishmentRequest struct {
	// IssuerPlayerID identifies the authorized moderator applying the punishment.
	IssuerPlayerID int64 `json:"issuerPlayerId"`
	// Kind selects ban, mute, warn, trade_lock, or kick.
	Kind string `json:"kind"`
	// Reason stores the required moderator explanation.
	Reason string `json:"reason"`
	// DurationHours optionally creates a finite punishment.
	DurationHours *int32 `json:"durationHours,omitempty"`
	// CFHTopicID optionally links the originating topic.
	CFHTopicID *int64 `json:"cfhTopicId,omitempty"`
	// IssueID optionally links the originating issue.
	IssueID *int64 `json:"issueId,omitempty"`
}

// RevokePunishmentRequest identifies the moderator revoking a punishment.
type RevokePunishmentRequest struct {
	// RevokedByPlayerID identifies the authorized moderator.
	RevokedByPlayerID int64 `json:"revokedByPlayerId"`
}

// PunishmentResponse contains one durable global punishment.
type PunishmentResponse struct {
	// ID identifies the punishment.
	ID int64 `json:"id"`
	// ReceiverPlayerID identifies the target.
	ReceiverPlayerID int64 `json:"receiverPlayerId"`
	// IssuerPlayerID optionally identifies the moderator.
	IssuerPlayerID *int64 `json:"issuerPlayerId,omitempty"`
	// IssuerKind stores player or system.
	IssuerKind string `json:"issuerKind"`
	// Kind stores the punishment behavior.
	Kind string `json:"kind"`
	// Reason stores its explanation.
	Reason string `json:"reason"`
	// CFHTopicID optionally identifies the source topic.
	CFHTopicID *int64 `json:"cfhTopicId,omitempty"`
	// IssueID optionally identifies the source issue.
	IssueID *int64 `json:"issueId,omitempty"`
	// Source identifies the issuing workflow.
	Source string `json:"source"`
	// IssuedAt stores creation time.
	IssuedAt time.Time `json:"issuedAt"`
	// ExpiresAt optionally stores expiry.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// RevokedAt optionally stores early revocation time.
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	// RevokedByPlayerID optionally identifies the revoking moderator.
	RevokedByPlayerID *int64 `json:"revokedByPlayerId,omitempty"`
	// Active reports timestamp-derived active state.
	Active bool `json:"active"`
}

// TopicRequest describes a new call-for-help topic.
type TopicRequest struct {
	// Category groups topics in Nitro.
	Category string `json:"category"`
	// NameKey stores the localized display key.
	NameKey string `json:"nameKey"`
	// Action stores queue, auto_reply, or ignore.
	Action string `json:"action"`
	// AutoReplyKey optionally stores a localized automatic reply.
	AutoReplyKey *string `json:"autoReplyKey,omitempty"`
	// DefaultSanctionLadder enables escalation on default closure.
	DefaultSanctionLadder bool `json:"defaultSanctionLadder"`
	// Order stores display order.
	Order int32 `json:"order"`
	// Enabled controls client visibility.
	Enabled bool `json:"enabled"`
}

// TopicPatchRequest describes optional call-for-help topic changes.
type TopicPatchRequest struct {
	// Category optionally replaces the group.
	Category *string `json:"category,omitempty"`
	// NameKey optionally replaces the localized display key.
	NameKey *string `json:"nameKey,omitempty"`
	// Action optionally replaces behavior.
	Action *string `json:"action,omitempty"`
	// AutoReplyKey optionally replaces or clears the reply key.
	AutoReplyKey **string `json:"autoReplyKey,omitempty"`
	// DefaultSanctionLadder optionally changes escalation behavior.
	DefaultSanctionLadder *bool `json:"defaultSanctionLadder,omitempty"`
	// Order optionally replaces display order.
	Order *int32 `json:"order,omitempty"`
	// Enabled optionally changes visibility.
	Enabled *bool `json:"enabled,omitempty"`
}

// PresetRequest describes one localized moderator response.
type PresetRequest struct {
	// Category groups presets.
	Category string `json:"category"`
	// MessageKey stores the localized message key.
	MessageKey string `json:"messageKey"`
	// Enabled controls modtool projection.
	Enabled bool `json:"enabled"`
	// Order stores display order.
	Order int32 `json:"order"`
}

// PresetPatchRequest describes optional preset changes.
type PresetPatchRequest struct {
	// Category optionally replaces the group.
	Category *string `json:"category,omitempty"`
	// MessageKey optionally replaces the localized message key.
	MessageKey *string `json:"messageKey,omitempty"`
	// Enabled optionally changes visibility.
	Enabled *bool `json:"enabled,omitempty"`
	// Order optionally replaces display order.
	Order *int32 `json:"order,omitempty"`
}

// LadderEntryRequest configures one sanction escalation level.
type LadderEntryRequest struct {
	// Level identifies escalation order.
	Level int32 `json:"level"`
	// Kind selects warn, mute, or ban.
	Kind string `json:"kind"`
	// DurationHours stores zero for permanent punishment.
	DurationHours int32 `json:"durationHours"`
	// ProbationDays stores the repeat window.
	ProbationDays int32 `json:"probationDays"`
}

// LadderRequest replaces the complete escalation ladder.
type LadderRequest struct {
	// Items stores every ordered escalation level.
	Items []LadderEntryRequest `json:"items"`
}
