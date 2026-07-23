// Package effect executes compiled WIRED effects through focused realm boundaries.
package effect

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Status classifies one effect execution result.
type Status uint8

const (
	// Applied reports a successful state change.
	Applied Status = iota + 1
	// Skipped reports an absent actor, target, or no-op state.
	Skipped
	// Blocked reports a safety or authorization rejection.
	Blocked
)

// Result stores one effect result and derived room events.
type Result struct {
	// Status classifies execution outcome.
	Status Status
	// Derived stores events emitted after committed mutation.
	Derived []trigger.Event
	// CallTargets stores selected WIRED stacks to enqueue.
	CallTargets []int64
	// ResetTimers requests a room timer reset.
	ResetTimers bool
}

// FurnitureOperation identifies a furniture mutation.
type FurnitureOperation uint8

const (
	// ToggleState advances furniture state.
	ToggleState FurnitureOperation = iota + 1
	// MatchSnapshot restores captured state and placement.
	MatchSnapshot
	// MoveRotate moves or rotates furniture.
	MoveRotate
	// ChaseActor moves furniture toward the actor.
	ChaseActor
	// FleeActor moves furniture away from the actor.
	FleeActor
	// MoveDirection moves furniture in a configured direction.
	MoveDirection
	// ToggleRandomState selects a distinct valid state.
	ToggleRandomState
	// MoveFurnitureTo moves furniture relative to a target.
	MoveFurnitureTo
)

// AvatarOperation identifies a player-facing room mutation.
type AvatarOperation uint8

const (
	// ShowMessage sends a private message to the actor or every player for actorless triggers.
	ShowMessage AvatarOperation = iota + 1
	// TeleportAvatar moves the actor onto selected furniture.
	TeleportAvatar
	// KickAvatar removes the actor from the room.
	KickAvatar
	// MuteAvatar applies a temporary room mute.
	MuteAvatar
	// GiveRespect grants validated durable respect.
	GiveRespect
	// AlertAvatar sends a localized alert.
	AlertAvatar
	// GiveHanditem changes the actor's room hand item.
	GiveHanditem
	// GiveEffect changes the actor's room-scoped effect.
	GiveEffect
)

// BotOperation identifies a bot realm action.
type BotOperation uint8

const (
	// BotTeleport teleports a named bot.
	BotTeleport BotOperation = iota + 1
	// BotMove requests normal bot movement.
	BotMove
	// BotTalk emits filtered bot speech.
	BotTalk
	// BotGiveHanditem gives the actor a hand item.
	BotGiveHanditem
	// BotFollowAvatar changes a bot follow target.
	BotFollowAvatar
	// BotClothes changes a bot figure.
	BotClothes
	// BotTalkToAvatar emits directed bot speech.
	BotTalkToAvatar
)

// GameOperation identifies an ephemeral game-state mutation.
type GameOperation uint8

const (
	// GiveScore adds score to the actor.
	GiveScore GameOperation = iota + 1
	// JoinTeam assigns the actor to a team.
	JoinTeam
	// LeaveTeam removes the actor from its team.
	LeaveTeam
	// GiveTeamScore adds score to a team.
	GiveTeamScore
	// ResetHighscore clears selected durable boards.
	ResetHighscore
)

// ProgressionOperation identifies a durable progression mutation.
type ProgressionOperation uint8

const (
	// ProgressAchievement advances a named achievement group.
	ProgressAchievement ProgressionOperation = iota + 1
	// ProgressQuest advances the configured active quest.
	ProgressQuest
	// StartQuest activates the configured quest.
	StartQuest
)

// FurnitureService executes validated furniture operations.
type FurnitureService interface {
	// ExecuteFurniture executes an operation using authoritative placement rules.
	ExecuteFurniture(context.Context, FurnitureOperation, *configuration.Node, trigger.Event) (Result, error)
}

// AvatarService executes validated player-facing operations.
type AvatarService interface {
	// ExecuteAvatar executes an operation after re-resolving the actor.
	ExecuteAvatar(context.Context, AvatarOperation, *configuration.Node, trigger.Event) (Result, error)
}

// BotService executes validated bot operations.
type BotService interface {
	// ExecuteBot executes an operation through the room bot runtime.
	ExecuteBot(context.Context, BotOperation, *configuration.Node, trigger.Event) (Result, error)
}

// GameService executes ephemeral game mutations.
type GameService interface {
	// ExecuteGame executes an operation in room-owned game state.
	ExecuteGame(context.Context, GameOperation, *configuration.Node, trigger.Event) (Result, error)
}

// RewardService performs durable idempotent reward claims.
type RewardService interface {
	// Claim selects and delivers one configured reward.
	Claim(context.Context, *configuration.Node, trigger.Event) (Result, error)
}

// ProgressionService executes durable achievement and quest mutations.
type ProgressionService interface {
	// ExecuteProgression executes one configured player mutation.
	ExecuteProgression(context.Context, ProgressionOperation, *configuration.Node, trigger.Event) (Result, error)
}

// Services composes focused effect boundaries.
type Services struct {
	// Furniture mutates room furniture.
	Furniture FurnitureService
	// Avatar mutates player room state.
	Avatar AvatarService
	// Bot controls room bots.
	Bot BotService
	// Game controls room-owned game state.
	Game GameService
	// Reward delivers durable rewards.
	Reward RewardService
	// Progression owns achievement and quest lifecycle.
	Progression ProgressionService
}

// Executor dispatches compiled effects to focused boundaries.
type Executor struct {
	// services stores effect realm boundaries.
	services Services
}

// New creates an effect executor.
func New(services Services) *Executor { return &Executor{services: services} }
