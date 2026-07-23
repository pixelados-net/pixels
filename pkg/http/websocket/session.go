package websocket

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap"
)

// socketSession binds one Fiber WebSocket to a connection session.
type socketSession struct {
	// conn is the underlying Fiber WebSocket connection.
	conn *fiberws.Conn
	// config controls transport timeouts and queue sizes.
	config Config
	// registry tracks the active protocol session.
	registry *netconn.Registry
	// log records transport events.
	log *zap.Logger
	// id identifies the protocol session.
	id netconn.ID
	// session owns protocol state and handlers.
	session *netconn.Session
	// queue serializes outbound writer operations.
	queue chan writeItem
	// stop signals writer and heartbeat loops.
	stop chan struct{}
	// done closes after transport cleanup completes.
	done chan struct{}
	// buffer stores partial inbound frame bytes.
	buffer []byte
	// closed reports whether disposal started.
	closed atomic.Bool
	// finishOnce guards terminal cleanup.
	finishOnce sync.Once
}

// newSocketSession creates and registers a socket session.
func newSocketSession(adapter *Adapter, conn *fiberws.Conn) (*socketSession, error) {
	socket := &socketSession{
		conn:     conn,
		config:   adapter.config,
		registry: adapter.registry,
		log:      adapter.log,
		id:       netconn.ID(uuid.NewString()),
		queue:    make(chan writeItem, adapter.config.QueueSize),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:                socket.id,
		Kind:              kind,
		RemoteAddr:        conn.IP(),
		Inbound:           adapter.handlers.Inbound,
		Outbound:          adapter.handlers.Outbound,
		SecurityPolicy:    netconn.SecurityPolicyForEnvironment(adapter.app.Environment),
		PacketLogger:      packetLoggerForEnvironment(adapter.app.Environment, adapter.log, adapter.logger),
		Sender:            socket.send,
		Disposer:          socket.dispose,
		SecurityActivator: socket.activate,
	})
	if err != nil {
		return nil, err
	}

	socket.session = session
	if err := adapter.registry.Register(session); err != nil {
		return nil, err
	}

	return socket, nil
}

// run starts transport loops and blocks until the read loop exits.
func (socket *socketSession) run(ctx context.Context) {
	go socket.writeLoop()
	go socket.heartbeatLoop(ctx)

	reason := socket.readLoop(ctx)
	if err := socket.session.Disconnect(ctx, reason); err != nil && err != netconn.ErrDisposed {
		socket.log.Debug("websocket disconnect failed", zap.Error(err))
	}

	socket.wait()
}

// send enqueues an outbound packet.
func (socket *socketSession) send(ctx context.Context, packet codec.Packet) error {
	if socket.closed.Load() {
		return netconn.ErrDisposed
	}

	item := writeItem{kind: writePacket, packet: packet}
	select {
	case socket.queue <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrQueueFull
	}
}

// activate enqueues a security activation barrier.
func (socket *socketSession) activate(ctx context.Context, channel netconn.SecureChannel) error {
	if socket.closed.Load() {
		return netconn.ErrDisposed
	}

	item := writeItem{kind: writeActivate, channel: channel}
	select {
	case socket.queue <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrQueueFull
	}
}

// dispose starts graceful transport disposal.
func (socket *socketSession) dispose(ctx context.Context, reason netconn.Reason) error {
	if socket.closed.Swap(true) {
		return netconn.ErrDisposed
	}

	socket.enqueueClose(ctx, reason)
	socket.wait()

	return nil
}

// wait blocks until the writer has stopped.
func (socket *socketSession) wait() {
	timer := time.NewTimer(socket.config.CloseGrace)
	defer timer.Stop()

	select {
	case <-socket.done:
	case <-timer.C:
		socket.finish()
	}
}

// finish closes the underlying WebSocket once.
func (socket *socketSession) finish() {
	socket.finishOnce.Do(func() {
		close(socket.stop)
		if socket.conn != nil {
			_ = socket.conn.Close()
		}
		close(socket.done)
	})
}
