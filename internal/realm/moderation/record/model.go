// Package record defines moderation persistence contracts.
package record

import "time"

// Topic stores one cached call-for-help topic.
type Topic struct {
	// ID identifies the topic.
	ID int64 `json:"id"`
	// Category groups topics in Nitro.
	Category string `json:"category"`
	// NameKey stores the localized display key.
	NameKey string `json:"nameKey"`
	// Action stores queue, auto_reply, or ignore.
	Action string `json:"action"`
	// AutoReplyKey optionally stores a localized automatic response.
	AutoReplyKey *string `json:"autoReplyKey,omitempty"`
	// DefaultSanctionLadder enables escalation when closing with default action.
	DefaultSanctionLadder bool `json:"defaultSanctionLadder"`
	// Order stores display order.
	Order int32 `json:"order"`
	// Enabled reports whether users may select the topic.
	Enabled bool `json:"enabled"`
}

// ChatEntry stores immutable evidence frozen with an issue.
type ChatEntry struct {
	// ID identifies the evidence row.
	ID int64 `json:"id"`
	// PlayerID optionally identifies the speaker.
	PlayerID *int64 `json:"playerId,omitempty"`
	// PatternID stores a client evidence pattern.
	PatternID string `json:"patternId"`
	// Message stores the evidence text.
	Message string `json:"message"`
	// CreatedAt stores the original or capture time.
	CreatedAt time.Time `json:"createdAt"`
}

// Issue stores one global moderation ticket.
type Issue struct {
	// ID identifies the issue.
	ID int64 `json:"id"`
	// ReporterPlayerID identifies the reporting player.
	ReporterPlayerID int64 `json:"reporterPlayerId"`
	// ReporterName stores the reporting player's display name when projected.
	ReporterName string `json:"reporterName,omitempty"`
	// ReportedPlayerID optionally identifies the target.
	ReportedPlayerID *int64 `json:"reportedPlayerId,omitempty"`
	// ReportedName stores the target player's display name when projected.
	ReportedName string `json:"reportedName,omitempty"`
	// RoomID optionally identifies the incident room.
	RoomID *int64 `json:"roomId,omitempty"`
	// PhotoItemID optionally identifies durable photo evidence.
	PhotoItemID *int64 `json:"photoItemId,omitempty"`
	// TopicID identifies the selected topic.
	TopicID int64 `json:"topicId"`
	// Kind identifies CFH, guide, or guardian origin.
	Kind string `json:"kind"`
	// Message stores the report text.
	Message string `json:"message"`
	// State stores open, picked, resolved, or deleted.
	State string `json:"state"`
	// Resolution optionally stores the close code.
	Resolution *int32 `json:"resolution,omitempty"`
	// PickedByPlayerID optionally identifies the assigned moderator.
	PickedByPlayerID *int64 `json:"pickedByPlayerId,omitempty"`
	// PickerName stores the assigned moderator's display name when projected.
	PickerName string `json:"pickerName,omitempty"`
	// PickedAt stores assignment time.
	PickedAt *time.Time `json:"pickedAt,omitempty"`
	// ClosedByPlayerID optionally identifies the resolver.
	ClosedByPlayerID *int64 `json:"closedByPlayerId,omitempty"`
	// ClosedAt stores resolution time.
	ClosedAt *time.Time `json:"closedAt,omitempty"`
	// CreatedAt stores report time.
	CreatedAt time.Time `json:"createdAt"`
	// Chatlog stores frozen evidence when requested.
	Chatlog []ChatEntry `json:"chatlog,omitempty"`
}

// ReportParams describes one call for help.
type ReportParams struct {
	// ReporterPlayerID identifies the reporter.
	ReporterPlayerID int64
	// ReportedPlayerID optionally identifies the target.
	ReportedPlayerID *int64
	// RoomID optionally identifies the incident room.
	RoomID *int64
	// PhotoItemID optionally identifies durable photo evidence.
	PhotoItemID *int64
	// TopicID identifies the selected topic.
	TopicID int64
	// Kind identifies the workflow.
	Kind string
	// Message stores the report description.
	Message string
	// Chatlog stores frozen evidence.
	Chatlog []ChatEntry
}

// Preferences stores modtool window geometry.
type Preferences struct {
	// PlayerID identifies the moderator.
	PlayerID int64 `json:"playerId"`
	// X stores horizontal position.
	X int32 `json:"x"`
	// Y stores vertical position.
	Y int32 `json:"y"`
	// Width stores window width.
	Width int32 `json:"width"`
	// Height stores window height.
	Height int32 `json:"height"`
}

// Preset stores one localized moderator response template.
type Preset struct {
	// ID identifies the preset.
	ID int64 `json:"id"`
	// Category groups presets.
	Category string `json:"category"`
	// MessageKey stores a localized text key.
	MessageKey string `json:"messageKey"`
	// Enabled controls projection.
	Enabled bool `json:"enabled"`
	// Order stores display order.
	Order int32 `json:"order"`
}

// RoomVisit stores one recent room entry.
type RoomVisit struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// RoomName stores its current title.
	RoomName string `json:"roomName"`
	// EnteredAt stores entry time.
	EnteredAt time.Time `json:"enteredAt"`
}
