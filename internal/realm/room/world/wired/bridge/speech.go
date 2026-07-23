package bridge

import (
	"context"
	"time"

	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// speechDepthKey identifies recursion depth inside effect-generated bot speech.
type speechDepthKey struct{}

// SpeechBridge feeds filtered player and bot speech into WIRED.
type SpeechBridge struct {
	// rooms resolves room lifecycle task queues.
	rooms *roomlive.Registry
	// engine matches and executes SAY triggers.
	engine *wiredruntime.Engine
}

// NewSpeechBridge creates the shared player and bot speech bridge.
func NewSpeechBridge(rooms *roomlive.Registry, engine *wiredruntime.Engine) *SpeechBridge {
	return &SpeechBridge{rooms: rooms, engine: engine}
}

// Intercept examines filtered player talk and shout before normal delivery.
func (bridge *SpeechBridge) Intercept(ctx context.Context, roomID int64, playerID int64, username string, message string) (bool, error) {
	event := trigger.Event{Kind: trigger.Say, RoomID: roomID, ActorKind: trigger.ActorPlayer, ActorID: playerID, PlayerID: playerID, Username: username, Message: message}
	return bridge.schedule(ctx, event)
}

// Intercept examines filtered bot speech before normal delivery.
func (bridge *SpeechBridge) InterceptBot(ctx context.Context, bot sdkbot.Bot, message string, _ sdkbot.Scope, _ int64) (string, bool, error) {
	event := trigger.Event{Kind: trigger.Say, RoomID: bot.RoomID, ActorKind: trigger.ActorBot, ActorID: botcore.EntityKey(bot.ID), Username: bot.Name, Message: message}
	consumed, err := bridge.schedule(ctx, event)
	return message, consumed, err
}

// InterceptPet examines localized pet speech before normal delivery.
func (bridge *SpeechBridge) InterceptPet(ctx context.Context, roomID int64, entityKey int64, message string) (bool, error) {
	event := trigger.Event{Kind: trigger.Say, RoomID: roomID, ActorKind: trigger.ActorPet, ActorID: entityKey, Message: message}
	return bridge.schedule(ctx, event)
}

// schedule queues one matching speech event on its active room.
func (bridge *SpeechBridge) schedule(ctx context.Context, event trigger.Event) (bool, error) {
	depth, _ := ctx.Value(speechDepthKey{}).(int)
	if depth >= 4 || !bridge.engine.Matches(event) {
		return false, nil
	}
	active, found := bridge.rooms.Find(event.RoomID)
	if !found {
		return false, nil
	}
	executionContext := context.WithValue(ctx, speechDepthKey{}, depth+1)
	active.Schedule(0, func(now time.Time) {
		_, _ = bridge.engine.Process(executionContext, event, now)
	})
	return true, nil
}

// BotSpeechAdapter maps the SDK method name without weakening either interface.
type BotSpeechAdapter struct {
	// bridge stores the shared speech behavior.
	bridge *SpeechBridge
}

// NewBotSpeechInterceptor adapts the shared bridge to the bot SDK contract.
func NewBotSpeechInterceptor(bridge *SpeechBridge) sdkbot.SpeechInterceptor {
	return BotSpeechAdapter{bridge: bridge}
}

// Intercept implements sdkbot.SpeechInterceptor.
func (adapter BotSpeechAdapter) Intercept(ctx context.Context, bot sdkbot.Bot, message string, scope sdkbot.Scope, targetPlayerID int64) (string, bool, error) {
	return adapter.bridge.InterceptBot(ctx, bot, message, scope, targetPlayerID)
}
