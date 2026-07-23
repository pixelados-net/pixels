// Package bot exposes controlled bot behavior extension contracts.
package bot

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrUnsupportedSkill reports a skill outside a behavior contract.
	ErrUnsupportedSkill = errors.New("unsupported bot skill")
)

// Scope identifies bot chat delivery semantics.
type Scope uint8

const (
	// ScopeTalk broadcasts ordinary room chat.
	ScopeTalk Scope = iota + 1
	// ScopeShout broadcasts room-wide shout chat.
	ScopeShout
	// ScopeWhisper delivers chat only to one player.
	ScopeWhisper
)

// Bot is an immutable behavior-facing bot snapshot.
type Bot struct {
	// ID identifies the bot.
	ID int64
	// OwnerPlayerID identifies the bot owner.
	OwnerPlayerID int64
	// RoomID identifies the active room.
	RoomID int64
	// Name stores the visible name.
	Name string
	// BehaviorType stores the registered discriminator.
	BehaviorType string
	// CanWalk reports whether random movement is enabled.
	CanWalk bool
	// FollowingPlayerID identifies an active follow target.
	FollowingPlayerID int64
	// WalkDue reports whether the core schedule permits another walk attempt.
	WalkDue bool
	// ChatDue reports whether automatic chat is due.
	ChatDue bool
	// ChatMessage stores the already-expanded and filtered automatic line.
	ChatMessage string
}

// Message describes player speech delivered to a bot behavior.
type Message struct {
	// PlayerID identifies the speaker.
	PlayerID int64
	// Text stores the visible filtered text.
	Text string
}

// Visit describes one prior room entry.
type Visit struct {
	// PlayerID identifies the visitor.
	PlayerID int64
	// PlayerName stores the visitor name.
	PlayerName string
	// EnteredAt stores visit time.
	EnteredAt time.Time
}

// Actions exposes bounded room operations to registered behaviors.
type Actions interface {
	// RandomWalk attempts one bounded in-memory movement decision.
	RandomWalk(context.Context, Bot) error
	// Talk sends filtered bot chat using the requested scope.
	Talk(context.Context, Bot, string, Scope, int64) error
	// ServeKeyword resolves and delivers one configured bartender item.
	ServeKeyword(context.Context, Bot, Message) (bool, error)
	// Visits returns bounded room visit history.
	Visits(context.Context, Bot, int64) ([]Visit, error)
}

// Behavior defines an independently registered bot behavior.
type Behavior interface {
	// Type returns the persisted behavior discriminator.
	Type() string
	// OnPlace runs after a bot enters a room world.
	OnPlace(context.Context, Bot, Actions) error
	// OnPickup runs before a bot leaves its room world.
	OnPickup(context.Context, Bot, Actions) error
	// OnCycle runs from the room's single 500ms owner loop.
	OnCycle(context.Context, Bot, Actions) error
	// OnUserSay reacts asynchronously to delivered player chat.
	OnUserSay(context.Context, Bot, Message, Actions) error
	// OnUserEnter reacts asynchronously to a player room entry.
	OnUserEnter(context.Context, Bot, int64, Actions) error
	// SaveCustomSkill handles behavior-owned skill identifiers.
	SaveCustomSkill(context.Context, Bot, int32, string) error
}

// SpeechInterceptor provides the explicit integration boundary for Wired speech triggers.
type SpeechInterceptor interface {
	// Intercept examines bot speech before normal filtered delivery.
	Intercept(context.Context, Bot, string, Scope, int64) (string, bool, error)
}

// NoopSpeechInterceptor preserves normal speech until Wired behavior is implemented.
type NoopSpeechInterceptor struct{}

// Intercept returns the original message without consuming it.
func (NoopSpeechInterceptor) Intercept(_ context.Context, _ Bot, message string, _ Scope, _ int64) (string, bool, error) {
	return message, false, nil
}
