package behavior

import (
	"context"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// Generic implements the standard walking and automatic-chat behavior.
type Generic struct{}

// Type returns the generic discriminator.
func (Generic) Type() string { return "generic" }

// OnPlace accepts standard placement without extra state.
func (Generic) OnPlace(context.Context, sdkbot.Bot, sdkbot.Actions) error { return nil }

// OnPickup accepts standard pickup without extra state.
func (Generic) OnPickup(context.Context, sdkbot.Bot, sdkbot.Actions) error { return nil }

// OnCycle performs due generic walking and chat decisions.
func (Generic) OnCycle(ctx context.Context, bot sdkbot.Bot, actions sdkbot.Actions) error {
	if bot.WalkDue && bot.CanWalk && bot.FollowingPlayerID == 0 {
		if err := actions.RandomWalk(ctx, bot); err != nil {
			return err
		}
	}
	if bot.ChatDue && bot.ChatMessage != "" {
		return actions.Talk(ctx, bot, bot.ChatMessage, sdkbot.ScopeTalk, 0)
	}
	return nil
}

// OnUserSay ignores player speech.
func (Generic) OnUserSay(context.Context, sdkbot.Bot, sdkbot.Message, sdkbot.Actions) error {
	return nil
}

// OnUserEnter ignores player entries.
func (Generic) OnUserEnter(context.Context, sdkbot.Bot, int64, sdkbot.Actions) error { return nil }

// SaveCustomSkill rejects skills that are not part of the generic contract.
func (Generic) SaveCustomSkill(context.Context, sdkbot.Bot, int32, string) error {
	return sdkbot.ErrUnsupportedSkill
}
