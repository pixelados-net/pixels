// Package guardian coordinates ephemeral peer chat review.
package guardian

import (
	"time"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// State identifies review lifecycle.
type State uint8

const (
	// StateOffered waits for guardian decisions.
	StateOffered State = iota + 1
	// StateVoting accepts votes.
	StateVoting
	// StateClosed stores a final result.
	StateClosed
)

// Verdict identifies acceptable, bad, horrible, or mixed review.
type Verdict int32

const (
	// VerdictAcceptable reports no actionable behavior.
	VerdictAcceptable Verdict = iota
	// VerdictBad reports actionable behavior.
	VerdictBad
	// VerdictHorrible reports severe behavior.
	VerdictHorrible
	// VerdictMixed reports no strict majority.
	VerdictMixed
)

// Reviewer stores offer and vote state.
type Reviewer struct {
	// PlayerID identifies the guardian.
	PlayerID int64
	// Accepted reports offer acceptance.
	Accepted bool
	// Decided reports an explicit offer decision.
	Decided bool
	// Vote optionally stores the submitted verdict.
	Vote *Verdict
}

// Ticket stores one anonymized review session.
type Ticket struct {
	// ID identifies the runtime review.
	ID int64
	// ReporterPlayerID identifies the requester.
	ReporterPlayerID int64
	// ReportedPlayerID identifies the chat sender.
	ReportedPlayerID int64
	// State stores lifecycle state.
	State State
	// CreatedAt stores creation time.
	CreatedAt time.Time
	// ClosesAt stores the voting deadline.
	ClosesAt time.Time
	// Reviewers stores offered guardians.
	Reviewers map[int64]*Reviewer
	// Chatlog stores anonymized frozen evidence.
	Chatlog []moderationrecord.ChatEntry
	// Result stores the final verdict.
	Result Verdict
}
