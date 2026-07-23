package behavior

import (
	"context"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// Bartender serves configured hand items when a nearby player says a keyword.
type Bartender struct{ Generic }

// Type returns the bartender discriminator.
func (Bartender) Type() string { return "bartender" }

// OnUserSay attempts one whole-word serving match.
func (Bartender) OnUserSay(ctx context.Context, bot sdkbot.Bot, message sdkbot.Message, actions sdkbot.Actions) error {
	_, err := actions.ServeKeyword(ctx, bot, message)
	return err
}
