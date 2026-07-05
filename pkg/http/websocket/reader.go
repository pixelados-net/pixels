package websocket

import (
	"context"
	"errors"
	"time"

	fastws "github.com/fasthttp/websocket"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// readLoop reads WebSocket messages until a terminal reason occurs.
func (socket *socketSession) readLoop(ctx context.Context) netconn.Reason {
	for {
		messageType, data, err := socket.readMessage()
		if err != nil {
			return readReason(err)
		}

		if messageType != fastws.BinaryMessage {
			return netconn.Reason{Code: netconn.DisconnectProtocolError, Message: "non-binary websocket message"}
		}

		if reason, ok := socket.receive(ctx, data); !ok {
			return reason
		}
	}
}

// readMessage reads one WebSocket message with a deadline.
func (socket *socketSession) readMessage() (int, []byte, error) {
	if err := socket.conn.SetReadDeadline(time.Now().Add(socket.config.ReadTimeout)); err != nil {
		return 0, nil, err
	}

	return socket.conn.ReadMessage()
}

// receive opens, decodes, and dispatches bytes.
func (socket *socketSession) receive(ctx context.Context, data []byte) (netconn.Reason, bool) {
	opened, err := socket.session.Open(data)
	if err != nil {
		return netconn.Reason{Code: netconn.DisconnectProtocolError, Message: err.Error()}, false
	}

	socket.buffer = append(socket.buffer, opened...)
	packets, rest, err := codec.DecodeFrames(nil, socket.buffer)
	if err != nil {
		return netconn.Reason{Code: netconn.DisconnectProtocolError, Message: err.Error()}, false
	}

	socket.buffer = append(socket.buffer[:0], rest...)
	for _, packet := range packets {
		if reason, ok := socket.dispatch(ctx, packet); !ok {
			return reason, false
		}
	}

	return netconn.Reason{}, true
}

// dispatch routes one packet to the session.
func (socket *socketSession) dispatch(ctx context.Context, packet codec.Packet) (netconn.Reason, bool) {
	err := socket.session.Receive(ctx, packet)
	if err == nil {
		return netconn.Reason{}, true
	}

	if errors.Is(err, netconn.ErrDisposed) {
		return netconn.Reason{Code: netconn.DisconnectLocalClose, Message: err.Error()}, false
	}

	if errors.Is(err, netconn.ErrHandlerPolicy) || errors.Is(err, netconn.ErrHandlerNotFound) {
		return netconn.Reason{Code: netconn.DisconnectProtocolError, Message: err.Error()}, false
	}

	return netconn.Reason{Code: netconn.DisconnectProtocolError, Message: err.Error()}, false
}

// readReason maps WebSocket read errors to disconnect reasons.
func readReason(err error) netconn.Reason {
	if fastws.IsCloseError(err, fastws.CloseNormalClosure, fastws.CloseGoingAway) {
		return netconn.Reason{Code: netconn.DisconnectRemoteClose, Message: err.Error()}
	}

	return netconn.Reason{Code: netconn.DisconnectTransportError, Message: err.Error()}
}
