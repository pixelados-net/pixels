package websocket

import (
	"context"
	"strconv"
	"time"

	fastws "github.com/fasthttp/websocket"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outreason "github.com/niflaot/pixels/networking/outbound/client/disconnect/reason"
	outerror "github.com/niflaot/pixels/networking/outbound/connection/error"
	"go.uber.org/zap"
)

// writeKind names queued writer item behavior.
type writeKind uint8

const (
	// writePacket sends a protocol packet.
	writePacket writeKind = iota + 1
	// writeActivate activates a secure channel.
	writeActivate
	// writeClose sends a WebSocket close frame.
	writeClose
)

// writeItem is one serialized WebSocket writer operation.
type writeItem struct {
	// kind selects how the writer handles this item.
	kind writeKind
	// packet stores the outbound protocol packet.
	packet codec.Packet
	// channel stores the secure channel activation target.
	channel netconn.SecureChannel
	// reason stores the close reason for terminal writes.
	reason netconn.Reason
}

// writeLoop serializes all WebSocket writes.
func (socket *socketSession) writeLoop() {
	for {
		select {
		case item := <-socket.queue:
			if !socket.handleWrite(item) {
				return
			}
		case <-socket.stop:
			return
		}
	}
}

// errorPacket returns a protocol error packet for a close reason.
func errorPacket(reason netconn.Reason) (codec.Packet, bool) {
	code, ok := pixelErrorCode(reason.Code)
	if !ok {
		return codec.Packet{}, false
	}

	packet, err := outerror.Encode(0, code, "")
	if err != nil {
		return codec.Packet{}, false
	}

	return packet, true
}

// disconnectPacket returns Nitro's protocol-native disconnect reason packet.
func disconnectPacket(reason netconn.Reason) (codec.Packet, bool) {
	code, ok := pixelDisconnectCode(reason.Code)
	if !ok {
		return codec.Packet{}, false
	}

	packet, err := outreason.Encode(outreason.WithReason(code))
	if err != nil {
		return codec.Packet{}, false
	}

	return packet, true
}

// handleWrite applies one queued write item.
func (socket *socketSession) handleWrite(item writeItem) bool {
	switch item.kind {
	case writePacket:
		return socket.writePacket(item.packet)
	case writeActivate:
		return socket.activateSecurity(item.channel)
	case writeClose:
		socket.writeClose(item.reason)
		socket.finish()
		return false
	default:
		return true
	}
}

// writePacket writes one encoded packet.
func (socket *socketSession) writePacket(packet codec.Packet) bool {
	frame, err := codec.AppendFrame(nil, packet)
	if err == nil {
		frame, err = socket.session.Seal(frame)
	}
	if err == nil {
		err = socket.conn.SetWriteDeadline(time.Now().Add(socket.config.WriteTimeout))
	}
	if err == nil {
		err = socket.conn.WriteMessage(fastws.BinaryMessage, frame)
	}
	if err != nil {
		socket.log.Debug("websocket write failed", zap.Error(err))
		socket.finish()
		return false
	}

	return true
}

// activateSecurity activates a queued secure channel.
func (socket *socketSession) activateSecurity(channel netconn.SecureChannel) bool {
	if err := socket.session.ActivateSecurity(channel); err != nil {
		socket.log.Debug("websocket security activation failed", zap.Error(err))
		socket.finish()
		return false
	}

	return true
}

// enqueueClose enqueues final protocol and close frames.
func (socket *socketSession) enqueueClose(ctx context.Context, reason netconn.Reason) {
	if packet, ok := disconnectPacket(reason); ok {
		socket.enqueueBestEffort(ctx, writeItem{kind: writePacket, packet: packet})
	}
	if packet, ok := errorPacket(reason); ok {
		socket.enqueueBestEffort(ctx, writeItem{kind: writePacket, packet: packet})
	}

	socket.enqueueBestEffort(ctx, writeItem{kind: writeClose, reason: reason})
}

// enqueueBestEffort queues an item or closes directly.
func (socket *socketSession) enqueueBestEffort(ctx context.Context, item writeItem) {
	timer := time.NewTimer(socket.config.CloseGrace)
	defer timer.Stop()

	select {
	case socket.queue <- item:
	case <-ctx.Done():
		socket.finish()
	case <-timer.C:
		socket.finish()
	}
}

// pixelErrorCode maps disconnect codes to client error codes.
func pixelErrorCode(code netconn.DisconnectCode) (int32, bool) {
	switch code {
	case netconn.DisconnectRemoteClose, netconn.DisconnectTransportError:
		return 0, false
	default:
		return 0, true
	}
}

// pixelDisconnectCode maps server disconnects to Nitro's reason enumeration.
func pixelDisconnectCode(code netconn.DisconnectCode) (int32, bool) {
	switch code {
	case netconn.DisconnectRemoteClose, netconn.DisconnectTransportError:
		return 0, false
	case netconn.DisconnectBanned:
		return 1, true
	case netconn.DisconnectDuplicateSession:
		return 2, true
	default:
		return 0, true
	}
}

// websocketCloseCode maps disconnect codes to WebSocket close codes.
func websocketCloseCode(code netconn.DisconnectCode) int {
	switch code {
	case netconn.DisconnectLocalClose, netconn.DisconnectRemoteClose, netconn.DisconnectDuplicateSession, netconn.DisconnectKicked:
		return fastws.CloseNormalClosure
	case netconn.DisconnectProtocolError:
		return fastws.CloseProtocolError
	case netconn.DisconnectAuthenticationFailed, netconn.DisconnectAuthenticationTimeout:
		return fastws.ClosePolicyViolation
	case netconn.DisconnectIdleTimeout:
		return fastws.CloseGoingAway
	case netconn.DisconnectRateLimited, netconn.DisconnectPolicyViolation, netconn.DisconnectBanned:
		return fastws.ClosePolicyViolation
	case netconn.DisconnectServerShutdown:
		return fastws.CloseServiceRestart
	default:
		return fastws.CloseInternalServerErr
	}
}

// writeClose writes the WebSocket close control frame.
func (socket *socketSession) writeClose(reason netconn.Reason) {
	code := websocketCloseCode(reason.Code)
	text := reason.Message
	if text == "" {
		text = reason.Code.String()
	}
	if len(text) > 80 {
		text = text[:80]
	}

	message := fastws.FormatCloseMessage(code, strconv.Itoa(int(reason.Code))+":"+text)
	_ = socket.conn.WriteControl(fastws.CloseMessage, message, time.Now().Add(socket.config.WriteTimeout))
}
