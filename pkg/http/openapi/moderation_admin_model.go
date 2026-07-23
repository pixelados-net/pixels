package openapi

import "time"

// ModerationPlayerRequest identifies one player punishment history.
type ModerationPlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the target player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// Limit bounds returned history.
	Limit int32 `query:"limit" minimum:"1" maximum:"500" default:"100"`
}

// ModerationPunishmentApplyRequest applies one global punishment.
type ModerationPunishmentApplyRequest struct {
	APIKeyRequest
	// PlayerID identifies the target player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// IssuerPlayerID identifies the authorized moderator.
	IssuerPlayerID int64 `json:"issuerPlayerId" required:"true" minimum:"1"`
	// Kind selects the registered punishment behavior.
	Kind string `json:"kind" required:"true" enum:"ban,mute,warn,trade_lock,kick"`
	// Reason stores the required explanation.
	Reason string `json:"reason" required:"true" maxLength:"500"`
	// DurationHours optionally makes a persistent punishment finite.
	DurationHours *int32 `json:"durationHours,omitempty" minimum:"1" maximum:"87600"`
	// CFHTopicID optionally links a help topic.
	CFHTopicID *int64 `json:"cfhTopicId,omitempty" minimum:"1"`
	// IssueID optionally links a moderation issue.
	IssueID *int64 `json:"issueId,omitempty" minimum:"1"`
}

// ModerationPunishmentRevokeRequest revokes one punishment.
type ModerationPunishmentRevokeRequest struct {
	APIKeyRequest
	// ID identifies the punishment.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// RevokedByPlayerID identifies the authorized moderator.
	RevokedByPlayerID int64 `json:"revokedByPlayerId" required:"true" minimum:"1"`
}

// ModerationPunishmentResponse documents one punishment.
type ModerationPunishmentResponse struct {
	// ID identifies the punishment.
	ID int64 `json:"id"`
	// ReceiverPlayerID identifies the target.
	ReceiverPlayerID int64 `json:"receiverPlayerId"`
	// IssuerPlayerID optionally identifies the moderator.
	IssuerPlayerID *int64 `json:"issuerPlayerId,omitempty"`
	// IssuerKind stores player or system.
	IssuerKind string `json:"issuerKind"`
	// Kind stores the behavior.
	Kind string `json:"kind"`
	// Reason stores the explanation.
	Reason string `json:"reason"`
	// Source stores the workflow.
	Source string `json:"source"`
	// IssuedAt stores creation time.
	IssuedAt time.Time `json:"issuedAt"`
	// ExpiresAt optionally stores expiry.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// RevokedAt optionally stores early revocation time.
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	// Active reports timestamp-derived active state.
	Active bool `json:"active"`
}

// ModerationPunishmentListResponse contains punishment history.
type ModerationPunishmentListResponse struct {
	// Items stores punishment rows.
	Items []ModerationPunishmentResponse `json:"items"`
	// Count stores returned row count.
	Count int `json:"count"`
}

// ModerationIssuesRequest filters external moderation tooling.
type ModerationIssuesRequest struct {
	APIKeyRequest
	// State optionally filters issue lifecycle state.
	State string `query:"state" enum:"open,picked,resolved,deleted"`
	// Limit bounds returned issues.
	Limit int32 `query:"limit" minimum:"1" maximum:"500" default:"100"`
}

// ModerationIssueResponse documents one issue summary.
type ModerationIssueResponse struct {
	// ID identifies the issue.
	ID int64 `json:"id"`
	// ReporterPlayerID identifies the reporter.
	ReporterPlayerID int64 `json:"reporterPlayerId"`
	// ReportedPlayerID optionally identifies the target.
	ReportedPlayerID *int64 `json:"reportedPlayerId,omitempty"`
	// TopicID identifies the selected topic.
	TopicID int64 `json:"topicId"`
	// Kind stores the source workflow.
	Kind string `json:"kind"`
	// Message stores the report description.
	Message string `json:"message"`
	// State stores issue lifecycle state.
	State string `json:"state"`
	// CreatedAt stores report time.
	CreatedAt time.Time `json:"createdAt"`
}

// ModerationIssueListResponse contains issue summaries.
type ModerationIssueListResponse struct {
	// Items stores issue rows.
	Items []ModerationIssueResponse `json:"items"`
	// Count stores returned row count.
	Count int `json:"count"`
}

// ModerationTopicRequest documents a complete topic body.
type ModerationTopicRequest struct {
	APIKeyRequest
	// Category groups topics.
	Category string `json:"category" required:"true"`
	// NameKey stores a localized display key.
	NameKey string `json:"nameKey" required:"true"`
	// Action selects queue, auto_reply, or ignore.
	Action string `json:"action" required:"true" enum:"queue,auto_reply,ignore"`
	// AutoReplyKey optionally stores a localized reply key.
	AutoReplyKey *string `json:"autoReplyKey,omitempty"`
	// DefaultSanctionLadder enables default escalation.
	DefaultSanctionLadder bool `json:"defaultSanctionLadder"`
	// Order stores display order.
	Order int32 `json:"order"`
	// Enabled controls client visibility.
	Enabled bool `json:"enabled"`
}

// ModerationTopicPatchRequest documents an identified topic patch.
type ModerationTopicPatchRequest struct {
	APIKeyRequest
	// ID identifies the topic.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// Category optionally replaces the group.
	Category *string `json:"category,omitempty"`
	// NameKey optionally replaces the localized display key.
	NameKey *string `json:"nameKey,omitempty"`
	// Action optionally replaces topic behavior.
	Action *string `json:"action,omitempty" enum:"queue,auto_reply,ignore"`
	// AutoReplyKey optionally replaces the localized automatic reply.
	AutoReplyKey *string `json:"autoReplyKey,omitempty"`
	// DefaultSanctionLadder optionally changes escalation behavior.
	DefaultSanctionLadder *bool `json:"defaultSanctionLadder,omitempty"`
	// Order optionally replaces display order.
	Order *int32 `json:"order,omitempty"`
	// Enabled optionally changes client visibility.
	Enabled *bool `json:"enabled,omitempty"`
}

// ModerationTopicResponse documents one topic.
type ModerationTopicResponse struct {
	// ID identifies the topic.
	ID int64 `json:"id"`
	ModerationTopicRequest
}

// ModerationTopicListResponse contains topics.
type ModerationTopicListResponse struct {
	// Items stores topic rows.
	Items []ModerationTopicResponse `json:"items"`
	// Count stores returned row count.
	Count int `json:"count"`
}

// ModerationPresetRequest documents one localized preset.
type ModerationPresetRequest struct {
	APIKeyRequest
	// Category groups presets.
	Category string `json:"category" required:"true"`
	// MessageKey stores the localized message key.
	MessageKey string `json:"messageKey" required:"true"`
	// Enabled controls projection.
	Enabled bool `json:"enabled"`
	// Order stores display order.
	Order int32 `json:"order"`
}

// ModerationPresetPatchRequest identifies one preset patch.
type ModerationPresetPatchRequest struct {
	APIKeyRequest
	// ID identifies the preset.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// Category optionally replaces the group.
	Category *string `json:"category,omitempty"`
	// MessageKey optionally replaces the localized message key.
	MessageKey *string `json:"messageKey,omitempty"`
	// Enabled optionally changes projection.
	Enabled *bool `json:"enabled,omitempty"`
	// Order optionally replaces display order.
	Order *int32 `json:"order,omitempty"`
}

// ModerationPresetResponse documents one preset.
type ModerationPresetResponse struct {
	// ID identifies the preset.
	ID int64 `json:"id"`
	ModerationPresetRequest
}

// ModerationPresetListResponse contains presets.
type ModerationPresetListResponse struct {
	// Items stores preset rows.
	Items []ModerationPresetResponse `json:"items"`
	// Count stores returned row count.
	Count int `json:"count"`
}

// ModerationLadderEntry documents one escalation level.
type ModerationLadderEntry struct {
	// Level identifies escalation order.
	Level int32 `json:"level" required:"true" minimum:"1"`
	// Kind selects warn, mute, or ban.
	Kind string `json:"kind" required:"true" enum:"warn,mute,ban"`
	// DurationHours stores zero for permanent punishment.
	DurationHours int32 `json:"durationHours" minimum:"0" maximum:"87600"`
	// ProbationDays stores the repeat window.
	ProbationDays int32 `json:"probationDays" minimum:"1"`
}

// ModerationLadderRequest replaces the complete ladder.
type ModerationLadderRequest struct {
	APIKeyRequest
	// Items stores contiguous escalation levels.
	Items []ModerationLadderEntry `json:"items" required:"true" minItems:"1"`
}

// ModerationLadderResponse contains escalation levels.
type ModerationLadderResponse struct {
	// Items stores escalation levels.
	Items []ModerationLadderEntry `json:"items"`
	// Count stores level count.
	Count int `json:"count"`
}
