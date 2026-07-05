package connection

import (
	"context"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlatency "github.com/niflaot/pixels/networking/inbound/client/latency"
	inpong "github.com/niflaot/pixels/networking/inbound/client/pong"
	outlatency "github.com/niflaot/pixels/networking/outbound/client/latency"
)

// background returns the command context used by connection handlers.
func background() context.Context {
	return context.Background()
}

// pongHandler handles client heartbeat pongs.
func pongHandler(context netconn.Context, packet codec.Packet) error {
	if _, err := inpong.Decode(packet); err != nil {
		return err
	}

	return context.MarkPong(time.Now())
}

// latencyHandler handles latency echo packets.
func latencyHandler(context netconn.Context, packet codec.Packet) error {
	payload, err := inlatency.Decode(packet)
	if err != nil {
		return err
	}

	response, err := outlatency.Encode(payload.RequestID)
	if err != nil {
		return err
	}

	return context.Send(background(), response)
}
