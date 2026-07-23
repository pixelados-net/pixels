// Package handlers implements explicit camera compatibility stubs.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incompetition "github.com/niflaot/pixels/networking/inbound/camera/competition"
	inthumbnail "github.com/niflaot/pixels/networking/inbound/camera/thumbnailupdate"
	outcompetition "github.com/niflaot/pixels/networking/outbound/camera/competitionstatus"
	outthumbnail "github.com/niflaot/pixels/networking/outbound/camera/thumbnailupdateresult"
)

// Handler responds safely to unsupported legacy camera flows.
type Handler struct{}

// New creates a camera compatibility handler.
func New() *Handler { return &Handler{} }

// Handle decodes and rejects one compatibility request explicitly.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	ctx := context.Background()
	switch packet.Header {
	case inthumbnail.Header:
		payload, err := inthumbnail.Decode(packet)
		if err != nil {
			return err
		}
		response, err := outthumbnail.Encode(payload.FlatID, 0)
		if err != nil {
			return err
		}
		return connection.Send(ctx, response)
	case incompetition.Header:
		if err := incompetition.Decode(packet); err != nil {
			return err
		}
		response, err := outcompetition.Encode(false, "")
		if err != nil {
			return err
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// Register adds camera compatibility headers to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	for _, header := range []uint16{inthumbnail.Header, incompetition.Header} {
		_ = registry.Register(header, handler.Handle)
	}
}
