// Package heartbeat contains connection heartbeat packet handlers.
package heartbeat

import (
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpong "github.com/niflaot/pixels/networking/inbound/client/pong"
)

// Register adds heartbeat handlers to a registry.
func Register(registry *netconn.HandlerRegistry) {
	_ = registry.Register(inpong.Header, Handler)
}

// Handler handles client heartbeat pongs.
func Handler(context netconn.Context, packet codec.Packet) error {
	if _, err := inpong.Decode(packet); err != nil {
		return err
	}

	return context.MarkPong(time.Now())
}
