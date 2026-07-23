// Package latency contains client latency packet handlers.
package latency

import (
	"context"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlatency "github.com/niflaot/pixels/networking/inbound/client/latency"
	outlatency "github.com/niflaot/pixels/networking/outbound/client/latency"
)

// Register adds latency handlers to a registry.
func Register(registry *netconn.HandlerRegistry) {
	_ = registry.Register(inlatency.Header, Handler)
}

// Handler handles latency echo packets.
func Handler(handler netconn.Context, packet codec.Packet) error {
	payload, err := inlatency.Decode(packet)
	if err != nil {
		return err
	}

	response, err := outlatency.Encode(payload.RequestID)
	if err != nil {
		return err
	}

	return handler.Send(context.Background(), response)
}
