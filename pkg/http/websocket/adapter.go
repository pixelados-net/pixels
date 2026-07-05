package websocket

import (
	"context"
	"strings"

	fiberws "github.com/gofiber/contrib/websocket"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/config/app"
	"go.uber.org/zap"
)

const kind netconn.Kind = "websocket"

// Adapter accepts WebSocket connections.
type Adapter struct {
	// config controls WebSocket transport behavior.
	config Config
	// app stores application environment settings.
	app app.Config
	// registry tracks active WebSocket sessions.
	registry *netconn.Registry
	// handlers routes connection-realm packets.
	handlers *realmconn.Handlers
	// log records WebSocket lifecycle events.
	log *zap.Logger
}

// New creates a WebSocket adapter.
func New(config Config, app app.Config, registry *netconn.Registry, handlers *realmconn.Handlers, log *zap.Logger) *Adapter {
	return &Adapter{
		config:   config.Normalize(),
		app:      app,
		registry: registry,
		handlers: handlers,
		log:      log,
	}
}

// Handle runs one WebSocket connection.
func (adapter *Adapter) Handle(conn *fiberws.Conn) {
	socket, err := newSocketSession(adapter, conn)
	if err != nil {
		_ = conn.Close()
		return
	}

	adapter.log.Debug("websocket connected", zap.String("id", string(socket.id)))
	socket.run(context.Background())
	adapter.registry.Remove(kind, socket.id)
	adapter.handlers.Disconnected(context.Background(), kind, socket.id)
	adapter.log.Debug("websocket disconnected", zap.String("id", string(socket.id)))
}

// packetLogger records development packet traffic.
type packetLogger struct {
	// log records packet traffic.
	log *zap.Logger
}

// packetLoggerForEnvironment returns a packet logger for development.
func packetLoggerForEnvironment(environment string, log *zap.Logger) netconn.PacketLogger {
	if !strings.EqualFold(environment, "development") || log == nil {
		return nil
	}

	return packetLogger{log: log}
}

// Received records an inbound packet.
func (logger packetLogger) Received(context netconn.Context, packet codec.Packet) {
	logger.log.Debug("packet received", packetFields(context, packet)...)
}

// Sent records an outbound packet.
func (logger packetLogger) Sent(context netconn.Context, packet codec.Packet) {
	logger.log.Debug("packet sent", packetFields(context, packet)...)
}

// Unhandled records an inbound packet without a registered handler.
func (logger packetLogger) Unhandled(context netconn.Context, packet codec.Packet) {
	logger.log.Warn("packet unhandled", packetFields(context, packet)...)
}

// Disconnected records a connection disposal reason.
func (logger packetLogger) Disconnected(context netconn.Context, reason netconn.Reason) {
	fields := append(connectionFields(context),
		zap.String("disconnect_code", reason.Code.String()),
		zap.String("disconnect_message", reason.Message),
	)
	logger.log.Debug("websocket disconnecting", fields...)
}

// packetFields returns structured packet log fields.
func packetFields(context netconn.Context, packet codec.Packet) []zap.Field {
	fields := connectionFields(context)

	return append(fields,
		zap.Uint16("packet_header", packet.Header),
		zap.Int("packet_payload_size", len(packet.Payload)),
		zap.Binary("packet_payload", packet.Payload),
	)
}

// connectionFields returns structured connection log fields.
func connectionFields(context netconn.Context) []zap.Field {
	return []zap.Field{
		zap.String("connection_id", string(context.ConnectionID)),
		zap.String("connection_kind", string(context.ConnectionKind)),
		zap.String("state", context.State.String()),
	}
}
