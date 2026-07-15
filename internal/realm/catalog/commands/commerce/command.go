// Package commerce executes catalog gift, voucher, bundle, and freshness requests.
package commerce

import (
	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/realm/catalog/gift"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies catalog commerce commands.
	Name command.Name = "catalog.commerce"
)

// Action identifies one catalog commerce operation.
type Action uint8

const (
	// BundleRules requests the bulk-discount acknowledgement.
	BundleRules Action = iota
	// GiftConfig requests gift wrapping options.
	GiftConfig
	// Giftable checks whether one offer may be gifted.
	Giftable
	// BuyGift purchases an offer for another player.
	BuyGift
	// RedeemVoucher redeems one voucher.
	RedeemVoucher
	// MarkNew acknowledges catalog novelty.
	MarkNew
	// PageExpiration requests the current page expiration.
	PageExpiration
	// EarliestExpiration requests the nearest visible page expiration.
	EarliestExpiration
	// NextLimited requests the next scheduled LTD offer.
	NextLimited
)

// Command contains one decoded catalog commerce request.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Action identifies requested behavior.
	Action Action
	// OfferID identifies an optional catalog offer.
	OfferID int64
	// Code stores an optional voucher code.
	Code string
	// ReceiverName identifies an optional gift recipient.
	ReceiverName string
	// Message stores an optional gift message.
	Message string
	// ExtraData stores optional product-specific buyer data.
	ExtraData string
	// SpriteID identifies the selected wrapping furniture sprite.
	SpriteID int32
	// BoxID identifies an optional wrapping box.
	BoxID int32
	// RibbonID identifies an optional wrapping ribbon.
	RibbonID int32
	// ShowMyFace controls sender identity visibility.
	ShowMyFace bool
}

// Handler executes catalog commerce commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps connections to authenticated players.
	Bindings *binding.Registry
	// Catalog manages catalog commerce.
	Catalog *catalogservice.Service
	// Connections stores active transport-agnostic sessions.
	Connections *netconn.Registry
	// GiftOptions stores gift wrapping choices.
	GiftOptions gift.Options
	// Log records unexpected gift delivery and purchase failures.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe commerce command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddUint8("action", uint8(input.Action))
	encoder.AddInt64("offer_id", input.OfferID)

	return nil
}
