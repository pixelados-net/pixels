// Package club executes subscription membership and monthly gift requests.
package club

import (
	"github.com/niflaot/pixels/internal/command"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies club subscription commands.
	Name command.Name = "subscription.club"
)

// Action identifies one club operation.
type Action uint8

const (
	// Status requests current membership state.
	Status Action = iota
	// Offers requests normal club offers.
	Offers
	// Extension requests one extension deal.
	Extension
	// PurchaseHC purchases the default HC offer.
	PurchaseHC
	// PurchaseVIP purchases the default VIP offer.
	PurchaseVIP
	// GiftInfo requests monthly gifts.
	GiftInfo
	// SelectGift claims one monthly gift.
	SelectGift
	// BuildersCount requests the neutral Builders Club count.
	BuildersCount
	// SMS requests direct SMS purchase availability.
	SMS
	// Kickback requests payday information.
	Kickback
)

// Command contains one decoded club request.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Action identifies requested behavior.
	Action Action
	// VIP reports whether an extension request targets VIP.
	VIP bool
	// GiftName identifies a selected monthly gift.
	GiftName string
	// OfferID identifies a selected club extension offer.
	OfferID int64
	// ProductName identifies a requested subscription product.
	ProductName string
}

// Handler executes club commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps connections to authenticated players.
	Bindings *binding.Registry
	// Subscriptions manages memberships and rewards.
	Subscriptions *core.Service
	// Catalog reads club gift offers.
	Catalog *catalogservice.Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe club command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddUint8("action", uint8(input.Action))

	return nil
}
