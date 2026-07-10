// Package buy adapts PURCHASE_FROM_CATALOG packets to catalog commands.
package buy

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	buycmd "github.com/niflaot/pixels/internal/realm/catalog/commands/buy"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbuy "github.com/niflaot/pixels/networking/inbound/catalog/item/buy"
	"go.uber.org/zap"
)

// New creates a catalog purchase packet handler.
func New(handler buycmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inbuy.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[buycmd.Command]{
			Command:  buycmd.Command{Connection: connection, PageID: int64(payload.PageID), OfferID: int64(payload.OfferID), ExtraData: payload.ExtraData, Amount: payload.Amount},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the catalog purchase handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inbuy.Header, handler)
}
