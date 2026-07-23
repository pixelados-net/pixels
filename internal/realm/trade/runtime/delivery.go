package runtime

import (
	"context"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// Sender delivers direct-trade packets through existing session bindings.
type Sender struct {
	bindings    *binding.Registry
	connections *netconn.Registry
}

// NewSender creates direct-trade delivery behavior.
func NewSender(bindings *binding.Registry, connections *netconn.Registry) *Sender {
	return &Sender{bindings: bindings, connections: connections}
}

// PlayerID resolves the authenticated player behind a connection.
func (sender *Sender) PlayerID(connection netconn.Context) (int64, error) {
	bound, found := sender.bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, binding.ErrBindingNotFound
	}
	return bound.PlayerID, nil
}

// Send sends one packet to an online player.
func (sender *Sender) Send(ctx context.Context, playerID int64, packet codec.Packet) error {
	bound, found := sender.bindings.FindByPlayer(playerID)
	if !found {
		return binding.ErrBindingNotFound
	}
	connection, found := sender.connections.Get(bound.ConnectionKind, bound.ConnectionID)
	if !found {
		return netconn.ErrConnectionNotFound
	}
	return connection.Send(ctx, packet)
}

// RemoteAddr returns one online player's transport address when exposed.
func (sender *Sender) RemoteAddr(playerID int64) string {
	bound, found := sender.bindings.FindByPlayer(playerID)
	if !found {
		return ""
	}
	connection, found := sender.connections.Get(bound.ConnectionKind, bound.ConnectionID)
	if !found {
		return ""
	}
	addressed, ok := connection.(interface{ RemoteAddr() string })
	if !ok {
		return ""
	}
	return addressed.RemoteAddr()
}

// Both sends one packet to both participants.
func (sender *Sender) Both(ctx context.Context, session *Session, packet codec.Packet) error {
	if err := sender.Send(ctx, session.First.PlayerID, packet); err != nil {
		return err
	}
	return sender.Send(ctx, session.Second.PlayerID, packet)
}
