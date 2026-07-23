// Package hotelview executes subscription-backed hotel-view requests.
package hotelview

import (
	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap/zapcore"
)

// Name identifies hotel-view subscription commands.
const Name command.Name = "subscription.hotelview"

// Action identifies a hotel-view request.
type Action uint8

const (
	// BonusRare requests configured currency progress.
	BonusRare Action = iota
	// Countdown requests the seconds remaining to a supplied date.
	Countdown
	// StartCampaign is an explicit compatibility NOOP because Nitro React has no caller and reference emulators define no campaign semantics for header 1697.
	StartCampaign
)

// Command contains one hotel-view request.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Action identifies requested behavior.
	Action Action
	// Value stores the date or retired campaign code.
	Value string
}

// Handler executes hotel-view subscription commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections.
	Bindings *binding.Registry
	// Subscriptions manages Bonus Rare progress.
	Subscriptions *core.Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddUint8("action", uint8(input.Action))
	encoder.AddString("value", input.Value)
	return nil
}
