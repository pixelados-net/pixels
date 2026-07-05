package websocket

import (
	"context"

	fiberws "github.com/gofiber/contrib/websocket"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
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
	adapter.log.Debug("websocket disconnected", zap.String("id", string(socket.id)))
}
