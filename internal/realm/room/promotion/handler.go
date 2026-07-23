package promotion

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/room/promotion/cancel"
	inclicked "github.com/niflaot/pixels/networking/inbound/room/promotion/clicked"
	inedit "github.com/niflaot/pixels/networking/inbound/room/promotion/edit"
	ininfo "github.com/niflaot/pixels/networking/inbound/room/promotion/info"
	ininitiated "github.com/niflaot/pixels/networking/inbound/room/promotion/initiated"
	inpurchase "github.com/niflaot/pixels/networking/inbound/room/promotion/purchase"
	insearch "github.com/niflaot/pixels/networking/inbound/room/promotion/search"
	inviewed "github.com/niflaot/pixels/networking/inbound/room/promotion/viewed"
	"go.uber.org/zap"
)

// NewHandler creates the grouped room-promotion packet adapter.
func NewHandler(handler CommandHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(connection, packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[Command]{Command: input, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register adds every room-promotion request header.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	if registry == nil || handler == nil {
		return
	}
	for _, header := range []uint16{ininfo.Header, inpurchase.Header, ininitiated.Header, inclicked.Header, inviewed.Header, insearch.Header, inedit.Header, incancel.Header} {
		_ = registry.Register(header, handler)
	}
}

// decode maps one room-promotion request to a command.
func decode(connection netconn.Context, packet codec.Packet) (Command, error) {
	input := Command{Connection: connection}
	switch packet.Header {
	case ininfo.Header:
		input.Action = PurchaseInfo
		return input, ininfo.Decode(packet)
	case inpurchase.Header:
		value, err := inpurchase.Decode(packet)
		input.Action = PurchaseAd
		input.Purchase = PurchaseParams{RoomID: int64(value.RoomID), PageID: int64(value.PageID), OfferID: int64(value.OfferID), Title: value.Title, Description: value.Description, CategoryID: value.CategoryID, Extended: value.Extended}
		return input, err
	case inedit.Header:
		value, err := inedit.Decode(packet)
		input.Action = EditEvent
		input.Edit = EditParams{PromotionID: int64(value.EventID), Title: value.Name, Description: value.Description}
		return input, err
	case incancel.Header:
		_, err := incancel.Decode(packet)
		input.Action = CancelEvent
		return input, err
	case ininitiated.Header:
		input.Action = Telemetry
		return input, ininitiated.Decode(packet)
	case inclicked.Header:
		_, err := inclicked.Decode(packet)
		input.Action = Telemetry
		return input, err
	case inviewed.Header:
		input.Action = Telemetry
		return input, inviewed.Decode(packet)
	case insearch.Header:
		_, err := insearch.Decode(packet)
		input.Action = Telemetry
		return input, err
	default:
		return input, codec.ErrUnexpectedHeader
	}
}
