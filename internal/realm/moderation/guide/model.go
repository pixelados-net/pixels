// Package guide coordinates ephemeral one-to-one helper sessions.
package guide

import "time"

// Duty describes one guide's accepted assistance queues.
type Duty struct {
	// PlayerID identifies the guide.
	PlayerID int64
	// Guide accepts normal help sessions.
	Guide bool
	// Bully accepts bullying help sessions.
	Bully bool
	// Guardian accepts peer chat review.
	Guardian bool
	// Since preserves FIFO duty order.
	Since time.Time
}

// Message stores one filtered guide transcript entry.
type Message struct {
	// SenderPlayerID identifies the sender.
	SenderPlayerID int64
	// Text stores filtered visible content.
	Text string
	// CreatedAt stores send time.
	CreatedAt time.Time
}

// State identifies guide session lifecycle.
type State uint8

const (
	// StateAttached waits for guide decision.
	StateAttached State = iota + 1
	// StateStarted permits chat and room actions.
	StateStarted
	// StateEnded awaits optional feedback.
	StateEnded
)

// Session stores one ephemeral requester-guide pairing.
type Session struct {
	// ID identifies the runtime session.
	ID int64
	// RequesterPlayerID identifies the help requester.
	RequesterPlayerID int64
	// GuidePlayerID identifies the matched guide.
	GuidePlayerID int64
	// Topic stores the client assistance type.
	Topic int32
	// Description stores the filtered problem text.
	Description string
	// State stores lifecycle state.
	State State
	// CreatedAt stores creation time.
	CreatedAt time.Time
	// Transcript stores bounded messages.
	Transcript []Message
}

// Partner returns the other participant for one player.
func (session Session) Partner(playerID int64) (int64, bool) {
	if playerID == session.RequesterPlayerID {
		return session.GuidePlayerID, true
	}
	if playerID == session.GuidePlayerID {
		return session.RequesterPlayerID, true
	}
	return 0, false
}
