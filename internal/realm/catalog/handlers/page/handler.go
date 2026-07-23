// Package page adapts GET_CATALOG_PAGE packets to catalog commands.
package page

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	pagecmd "github.com/niflaot/pixels/internal/realm/catalog/commands/page"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpage "github.com/niflaot/pixels/networking/inbound/catalog/page/request"
	"go.uber.org/zap"
)

// New creates a catalog page packet handler.
func New(handler pagecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inpage.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[pagecmd.Command]{
			Command:  pagecmd.Command{Connection: connection, PageID: int64(payload.PageID), OfferID: payload.OfferID, Mode: payload.Mode},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the catalog page handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inpage.Header, handler)
}
