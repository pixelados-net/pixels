package routes

import (
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
)

// AuditRequest attributes one administrative mutation.
type AuditRequest struct {
	// ActorPlayerID identifies the administrator.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason stores the required audit reason.
	Reason string `json:"reason"`
}

// CenterRequest describes one external game registration mutation.
type CenterRequest struct {
	// AuditRequest stores administrative attribution.
	AuditRequest
	// Version stores optimistic concurrency state.
	Version int64 `json:"version"`
	// Name stores the client display key.
	Name string `json:"name"`
	// BackgroundColor stores six RGB hexadecimal digits.
	BackgroundColor string `json:"bgColor"`
	// TextColor stores six RGB hexadecimal digits.
	TextColor string `json:"textColor"`
	// AssetURL stores lobby artwork.
	AssetURL string `json:"assetUrl"`
	// SupportURL stores support documentation.
	SupportURL string `json:"supportUrl"`
	// LaunchURL stores the external game entry point.
	LaunchURL string `json:"launchUrl"`
	// LaunchKind chooses url or params launch packets.
	LaunchKind gamecenterrecord.LaunchKind `json:"launchKind"`
	// Enabled controls player visibility.
	Enabled bool `json:"enabled"`
}

// ScoreRequest describes one manual weekly score update.
type ScoreRequest struct {
	// AuditRequest stores administrative attribution.
	AuditRequest
	// Year stores the ISO week year.
	Year int32 `json:"year"`
	// Week stores the ISO week number.
	Week int32 `json:"week"`
	// Score stores the submitted best score.
	Score int64 `json:"score"`
}

// PollRequest describes one poll and its nested questions.
type PollRequest struct {
	// AuditRequest stores administrative attribution.
	AuditRequest
	// Version stores optimistic concurrency state.
	Version int64 `json:"version"`
	// Title stores the internal poll title.
	Title string `json:"title"`
	// Headline stores the room offer headline.
	Headline string `json:"headline"`
	// Summary stores the room offer summary.
	Summary string `json:"summary"`
	// StartMessage stores introductory text.
	StartMessage string `json:"startMessage"`
	// ThanksMessage stores completion text.
	ThanksMessage string `json:"thanksMessage"`
	// RoomID optionally assigns a room.
	RoomID int64 `json:"roomId"`
	// RewardBadge optionally grants a badge once.
	RewardBadge string `json:"rewardBadge"`
	// Enabled controls room offers and answers.
	Enabled bool `json:"enabled"`
	// Questions stores ordered Nitro poll questions.
	Questions []outcontents.Question `json:"questions"`
}

// PollRoomRequest assigns or clears one room.
type PollRoomRequest struct {
	// AuditRequest stores administrative attribution.
	AuditRequest
	// Version stores optimistic concurrency state.
	Version int64 `json:"version"`
	// RoomID assigns a room or clears it with zero.
	RoomID int64 `json:"roomId"`
}
