// Package player implements bounded dynamic-plugin access to live players and packets.
package player

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// Access implements bounded player operations for one plugin.
type Access struct {
	// players stores connected player state.
	players *playerlive.Registry
	// bindings resolves authenticated connections.
	bindings *binding.Registry
	// connections sends packets and disconnects sessions.
	connections *netconn.Registry
	// inbound stores the shared inbound packet pipeline.
	inbound *netconn.HandlerRegistry
	// permissions resolves real permission nodes.
	permissions permissionservice.Checker
	// timeout bounds interceptor callbacks.
	timeout time.Duration
	// log records isolated failures.
	log *zap.Logger
	// scope stores plugin identity and health.
	scope *pluginruntime.Scope
}

// NewAccess creates one namespace-scoped player facade.
func NewAccess(players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, inbound *netconn.HandlerRegistry, permissions permissionservice.Checker, timeout time.Duration, log *zap.Logger, scope *pluginruntime.Scope) *Access {
	if log == nil {
		log = zap.NewNop()
	}
	return &Access{players: players, bindings: bindings, connections: connections, inbound: inbound, permissions: permissions, timeout: timeout, log: log, scope: scope}
}

// All returns immutable snapshots of connected players in stable id order.
func (access *Access) All() []sdkplugin.Player {
	players := access.players.Snapshot()
	snapshots := make([]sdkplugin.Player, 0, len(players))
	for _, current := range players {
		snapshots = append(snapshots, Snapshot(current))
	}
	sort.Slice(snapshots, func(left int, right int) bool { return snapshots[left].ID < snapshots[right].ID })
	return snapshots
}

// Find returns one immutable connected-player snapshot.
func (access *Access) Find(playerID int64) (sdkplugin.Player, bool) {
	current, found := access.players.Find(playerID)
	if !found {
		return sdkplugin.Player{}, false
	}
	return Snapshot(current), true
}

// Message sends one Nitro system alert to a connected player.
func (access *Access) Message(playerID int64, text string) error {
	connection, err := access.connection(playerID)
	if err != nil {
		return err
	}
	packet, err := outalert.Encode(text)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// Disconnect ends one player's live session with a plugin-authored reason.
func (access *Access) Disconnect(playerID int64, reason string) error {
	current, found := access.bindings.FindByPlayer(playerID)
	if !found {
		return binding.ErrBindingNotFound
	}
	return access.connections.Disconnect(context.Background(), current.ConnectionKind, current.ConnectionID, netconn.Reason{Code: netconn.DisconnectKicked, Message: reason})
}

// HasPermission resolves one real concrete permission node.
func (access *Access) HasPermission(playerID int64, node string) (bool, error) {
	parsed := permission.Node(node)
	if !parsed.Concrete() {
		return false, permission.ErrInvalidRegistration
	}
	return access.permissions.HasPermission(context.Background(), playerID, parsed)
}

// Intercept registers one guarded inbound packet middleware.
func (access *Access) Intercept(interceptor sdkplugin.PacketInterceptor, options sdkplugin.InterceptOptions) error {
	if interceptor == nil {
		return netconn.ErrInvalidHandler
	}
	return access.inbound.Intercept(options.Header, int(options.Priority), func(connection netconn.Context, packet codec.Packet, next netconn.InterceptorNext) error {
		if !access.scope.Enabled() {
			return next()
		}
		pluginPacket := sdkplugin.InterceptContext{Player: access.playerForConnection(connection), Header: packet.Header, Payload: append([]byte(nil), packet.Payload...)}
		return access.invokeInterceptor(context.Background(), pluginPacket, interceptor, next)
	})
}

// invokeInterceptor guards middleware and advances safely after infrastructure failure.
func (access *Access) invokeInterceptor(ctx context.Context, packet sdkplugin.InterceptContext, interceptor sdkplugin.PacketInterceptor, native netconn.InterceptorNext) error {
	var active, called, completed atomic.Bool
	var resultMutex sync.Mutex
	var nextResult error
	active.Store(true)
	pluginNext := func(nextContext context.Context) error {
		if !active.Load() || !called.CompareAndSwap(false, true) {
			return pluginruntime.ErrNextUnavailable
		}
		if err := nextContext.Err(); err != nil {
			return err
		}
		err := native()
		resultMutex.Lock()
		nextResult = err
		resultMutex.Unlock()
		completed.Store(true)
		return err
	}
	err := pluginruntime.InvokeCallback(ctx, access.timeout, access.scope, "packet interceptor", access.log, func(callbackContext context.Context) error {
		return interceptor(callbackContext, packet, pluginNext)
	})
	active.Store(false)
	if !errors.Is(err, pluginruntime.ErrCallbackPanic) && !errors.Is(err, pluginruntime.ErrCallbackTimeout) && !errors.Is(err, pluginruntime.ErrPluginDisabled) {
		return err
	}
	if !called.Load() {
		return native()
	}
	if completed.Load() {
		resultMutex.Lock()
		defer resultMutex.Unlock()
		return nextResult
	}
	return err
}

// connection resolves a player's live transport connection.
func (access *Access) connection(playerID int64) (netconn.Connection, error) {
	current, found := access.bindings.FindByPlayer(playerID)
	if !found {
		return nil, binding.ErrBindingNotFound
	}
	connection, found := access.connections.Get(current.ConnectionKind, current.ConnectionID)
	if !found {
		return nil, netconn.ErrConnectionNotFound
	}
	return connection, nil
}

// playerForConnection returns an authenticated immutable source snapshot.
func (access *Access) playerForConnection(connection netconn.Context) sdkplugin.Player {
	current, found := access.bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return sdkplugin.Player{}
	}
	player, found := access.Find(current.PlayerID)
	if !found {
		return sdkplugin.Player{}
	}
	return player
}

// Snapshot copies safe fields from one live player.
func Snapshot(player interface {
	ID() int64
	Username() string
	CurrentRoom() (int64, bool)
}) sdkplugin.Player {
	roomID, _ := player.CurrentRoom()
	return sdkplugin.Player{ID: player.ID(), Username: player.Username(), RoomID: roomID, Online: true}
}
