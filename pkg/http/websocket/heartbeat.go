package websocket

import (
	"context"
	"time"

	netconn "github.com/niflaot/pixels/networking/connection"
	outping "github.com/niflaot/pixels/networking/outbound/client/ping"
	"go.uber.org/zap"
)

// heartbeatLoop sends pings and closes idle sessions.
func (socket *socketSession) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(socket.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !socket.heartbeat(ctx) {
				return
			}
		case <-socket.done:
			return
		}
	}
}

// heartbeat performs one heartbeat check.
func (socket *socketSession) heartbeat(ctx context.Context) bool {
	if time.Since(socket.session.LastPongAt()) > socket.config.PongTimeout {
		_ = socket.session.Disconnect(ctx, netconn.Reason{Code: netconn.DisconnectIdleTimeout, Message: "pong timeout"})
		return false
	}

	packet, err := outping.Encode()
	if err != nil {
		socket.log.Debug("websocket ping encode failed", zap.Error(err))
		return true
	}

	if err := socket.session.Send(ctx, packet); err != nil {
		socket.log.Debug("websocket ping send failed", zap.Error(err))
		return false
	}

	return true
}
