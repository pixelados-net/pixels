package websocket

import (
	"context"
	"strings"

	fiberws "github.com/gofiber/contrib/websocket"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/config/app"
	logconfig "github.com/niflaot/pixels/pkg/logger"
	"go.uber.org/zap"
)

const kind netconn.Kind = "websocket"

const shortConnectionIDSize = 8

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
	// logger configures WebSocket log shape.
	logger logconfig.Config
}

// New creates a WebSocket adapter.
func New(config Config, app app.Config, registry *netconn.Registry, handlers *realmconn.Handlers, log *zap.Logger, logger logconfig.Config) *Adapter {
	return &Adapter{
		config:   config.Normalize(),
		app:      app,
		registry: registry,
		handlers: handlers,
		log:      log,
		logger:   logger,
	}
}

// Handle runs one WebSocket connection.
func (adapter *Adapter) Handle(conn *fiberws.Conn) {
	socket, err := newSocketSession(adapter, conn)
	if err != nil {
		_ = conn.Close()
		return
	}

	adapter.log.Debug("websocket connected", adapter.connectionIDField(socket.id))
	socket.run(context.Background())
	adapter.registry.Remove(kind, socket.id)
	adapter.handlers.Disconnected(context.Background(), kind, socket.id)
	adapter.log.Debug("websocket disconnected", adapter.connectionIDField(socket.id))
}

// packetLogger records development packet traffic.
type packetLogger struct {
	// log records packet traffic.
	log *zap.Logger
	// toon enables compact packet fields.
	toon bool
}

// packetLoggerForEnvironment returns a packet logger for development.
func packetLoggerForEnvironment(environment string, log *zap.Logger, logger logconfig.Config) netconn.PacketLogger {
	if !strings.EqualFold(environment, "development") || log == nil {
		return nil
	}

	return packetLogger{log: log, toon: logger.ToonConsole}
}

// Received records an inbound packet.
func (logger packetLogger) Received(context netconn.Context, packet codec.Packet) {
	logger.log.Debug("packet received", logger.packetFields(context, packet)...)
}

// Sent records an outbound packet.
func (logger packetLogger) Sent(context netconn.Context, packet codec.Packet) {
	logger.log.Debug("packet sent", logger.packetFields(context, packet)...)
}

// Unhandled records an inbound packet without a registered handler.
func (logger packetLogger) Unhandled(context netconn.Context, packet codec.Packet) {
	fields := append(logger.packetFields(context, packet), zap.String("error", "packet has no registered handler"))
	logger.log.Warn("packet unhandled", fields...)
}

// Disconnected records a connection disposal reason.
func (logger packetLogger) Disconnected(context netconn.Context, reason netconn.Reason) {
	fields := append(logger.connectionFields(context), zap.String("disconnect_code", reason.Code.String()))
	fields = append(fields, logger.errorField(reason.Message))
	logger.log.Debug("websocket disconnecting", fields...)
}

// packetFields returns structured packet log fields.
func (logger packetLogger) packetFields(context netconn.Context, packet codec.Packet) []zap.Field {
	fields := logger.connectionFields(context)

	if logger.toon {
		return append(fields,
			zap.Uint16("header", packet.Header),
			zap.Int("bytes", len(packet.Payload)),
			zap.Binary("payload", packet.Payload),
		)
	}

	return append(fields,
		zap.Uint16("packet_header", packet.Header),
		zap.Int("packet_payload_size", len(packet.Payload)),
		zap.Binary("packet_payload", packet.Payload),
	)
}

// connectionFields returns structured connection log fields.
func (logger packetLogger) connectionFields(context netconn.Context) []zap.Field {
	if logger.toon {
		return []zap.Field{
			zap.String("cid", shortConnectionID(context.ConnectionID)),
			zap.String("state", context.State.String()),
		}
	}

	return []zap.Field{
		zap.String("connection_id", string(context.ConnectionID)),
		zap.String("connection_kind", string(context.ConnectionKind)),
		zap.String("state", context.State.String()),
	}
}

// errorField returns the preferred error field for the active log shape.
func (logger packetLogger) errorField(message string) zap.Field {
	if logger.toon {
		return zap.String("error", message)
	}

	return zap.String("disconnect_message", message)
}

// connectionIDField returns the adapter lifecycle connection identifier field.
func (adapter *Adapter) connectionIDField(id netconn.ID) zap.Field {
	if adapter.logger.ToonConsole {
		return zap.String("cid", shortConnectionID(id))
	}

	return zap.String("id", string(id))
}

// shortConnectionID returns the visible connection identifier prefix.
func shortConnectionID(id netconn.ID) string {
	value := string(id)
	if len(value) <= shortConnectionIDSize {
		return value
	}

	return value[:shortConnectionIDSize]
}
