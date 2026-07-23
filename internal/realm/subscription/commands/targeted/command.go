// Package targeted executes personalized catalog offer requests.
package targeted

import (
	"github.com/niflaot/pixels/internal/command"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies targeted-offer commands.
	Name command.Name = "subscription.targeted"
)

// Action identifies one targeted-offer operation.
type Action uint8

const (
	// Current requests the current targeted offer.
	Current Action = iota
	// Next requests the next targeted offer.
	Next
	// Purchase purchases a targeted offer.
	Purchase
	// State records a viewed or dismissed state.
	State
	// Product requests one catalog product offer.
	Product
)

// Command contains one targeted-offer request.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Action identifies requested behavior.
	Action Action
	// OfferID identifies the targeted or catalog offer.
	OfferID int64
	// Quantity stores requested purchase units.
	Quantity int32
	// Dismissed reports a dismissed offer state.
	Dismissed bool
}

// Handler executes targeted-offer commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections.
	Bindings *binding.Registry
	// Subscriptions manages targeted offer state.
	Subscriptions *core.Service
	// Catalog reads catalog offer data.
	Catalog *catalogservice.Service
	// Translations localizes offer copy.
	Translations i18n.Translator
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddUint8("action", uint8(input.Action))
	encoder.AddInt64("offer_id", input.OfferID)

	return nil
}
