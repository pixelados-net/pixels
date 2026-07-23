package poll

import (
	"context"

	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
)

// Definition describes one DB-backed poll.
type Definition struct {
	// ID identifies the poll.
	ID int32
	// Title stores its internal title.
	Title string
	// Headline stores offer headline text.
	Headline string
	// Summary stores offer summary text.
	Summary string
	// StartMessage stores introductory text.
	StartMessage string
	// ThanksMessage stores completion text.
	ThanksMessage string
	// RoomID identifies the assigned room.
	RoomID int64
	// RewardBadge stores an optional completion badge.
	RewardBadge string
	// Version stores optimistic administration state.
	Version int64
	// Enabled reports whether players may answer the poll.
	Enabled bool
	// Questions stores ordered content.
	Questions []outcontents.Question
}

// Store persists DB-backed polls and responses.
type Store interface {
	// Polls returns every enabled poll with ordered questions.
	Polls(context.Context) ([]Definition, error)
	// Poll returns one enabled poll with ordered questions.
	Poll(context.Context, int32) (Definition, bool, error)
	// PollForRoom returns one enabled room assignment.
	PollForRoom(context.Context, int64) (Definition, bool, error)
	// Completed reports whether a player answered every question.
	Completed(context.Context, int64, int32) (bool, error)
	// SaveAnswer inserts one answer and reports first-time completion.
	SaveAnswer(context.Context, int64, int32, int32, []string) (bool, string, error)
	// Reject records a poll rejection idempotently.
	RejectPoll(context.Context, int64, int32) error
}

// BadgeGranter grants optional completion badges.
type BadgeGranter interface {
	// GrantBadge grants one durable badge.
	GrantBadge(context.Context, int64, string, string) (bool, error)
}
