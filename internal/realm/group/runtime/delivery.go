package runtime

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Delivery resolves authenticated actors and sends targeted group packets.
type Delivery struct {
	// bindings maps players to authenticated connections.
	bindings *binding.Registry
	// connections stores active transport-agnostic sessions.
	connections *netconn.Registry
}

// NewDelivery creates group packet delivery behavior.
func NewDelivery(bindings *binding.Registry, connections *netconn.Registry) *Delivery {
	return &Delivery{bindings: bindings, connections: connections}
}

// PlayerID resolves the authenticated actor behind one connection.
func (delivery *Delivery) PlayerID(connection netconn.Context) (int64, error) {
	if delivery == nil || delivery.bindings == nil {
		return 0, binding.ErrBindingNotFound
	}
	bound, found := delivery.bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, binding.ErrBindingNotFound
	}
	return bound.PlayerID, nil
}

// Send delivers one packet to an online player.
func (delivery *Delivery) Send(ctx context.Context, playerID int64, packet codec.Packet) (bool, error) {
	if delivery == nil || delivery.bindings == nil || delivery.connections == nil {
		return false, nil
	}
	bound, found := delivery.bindings.FindByPlayer(playerID)
	if !found {
		return false, nil
	}
	connection, found := delivery.connections.Get(bound.ConnectionKind, bound.ConnectionID)
	if !found {
		return false, nil
	}
	return true, connection.Send(ctx, packet)
}

// SendError sends one localized hotel-facing group error without disconnecting.
func SendError(ctx context.Context, connection netconn.Context, translations i18n.Translator, key i18n.Key) error {
	message := string(key)
	if translations != nil {
		message = translations.Default(key)
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}
