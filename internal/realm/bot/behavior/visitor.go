package behavior

import (
	"context"
	"strconv"
	"strings"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// VisitorLog reports bounded visit history to the room owner.
type VisitorLog struct{ Generic }

// Type returns the visitor-log discriminator.
func (VisitorLog) Type() string { return "visitor_log" }

// OnUserEnter offers a summary only to the bot owner.
func (VisitorLog) OnUserEnter(ctx context.Context, bot sdkbot.Bot, playerID int64, actions sdkbot.Actions) error {
	if playerID != bot.OwnerPlayerID {
		return nil
	}
	visits, err := actions.Visits(ctx, bot, playerID)
	if err != nil {
		return err
	}
	message := "bots.visitor.no_visits"
	if len(visits) > 0 {
		message = "bots.visitor.visits:" + strconv.Itoa(len(visits))
	}
	return actions.Talk(ctx, bot, message, sdkbot.ScopeWhisper, playerID)
}

// OnUserSay presents the visit list once when the owner responds affirmatively.
func (VisitorLog) OnUserSay(ctx context.Context, bot sdkbot.Bot, message sdkbot.Message, actions sdkbot.Actions) error {
	if message.PlayerID != bot.OwnerPlayerID || !affirmative(message.Text) {
		return nil
	}
	visits, err := actions.Visits(ctx, bot, message.PlayerID)
	if err != nil || len(visits) == 0 {
		return err
	}
	names := make([]string, 0, len(visits))
	for _, visit := range visits {
		names = append(names, visit.PlayerName+" ("+visit.EnteredAt.Format("2006-01-02 15:04")+")")
	}
	return actions.Talk(ctx, bot, strings.Join(names, ", "), sdkbot.ScopeWhisper, message.PlayerID)
}

// affirmative recognizes the localized baseline affirmative forms.
func affirmative(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "yes" || value == "si" || value == "sí"
}
