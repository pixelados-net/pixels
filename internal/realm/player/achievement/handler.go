package achievement

import (
	"context"
	"errors"
	"math"

	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incurrent "github.com/niflaot/pixels/networking/inbound/user/badge/current"
	inequip "github.com/niflaot/pixels/networking/inbound/user/badge/equip"
	inlist "github.com/niflaot/pixels/networking/inbound/user/badge/list"
	outcurrent "github.com/niflaot/pixels/networking/outbound/user/badge/current"
	outlist "github.com/niflaot/pixels/networking/outbound/user/badge/list"
)

// Handler adapts Nitro badge requests to achievement behavior.
type Handler struct {
	// Achievements owns badge persistence and equipped snapshots.
	Achievements *Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// RegisterHandlers installs badge inventory, current, and selection packets.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(inlist.Header, handler.inventory)
	_ = registry.Register(incurrent.Header, handler.current)
	_ = registry.Register(inequip.Header, handler.equip)
}

// inventory sends the authenticated player's complete badge inventory.
func (handler Handler) inventory(connection netconn.Context, packet codec.Packet) error {
	if err := inlist.Decode(packet); err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Achievements == nil {
		return nil
	}
	return handler.sendInventory(context.Background(), connection, playerID)
}

// current sends one room user's active badge slots.
func (handler Handler) current(connection netconn.Context, packet codec.Packet) error {
	playerID, err := incurrent.Decode(packet)
	if err != nil {
		return err
	}
	if playerID <= 0 || handler.Achievements == nil {
		return nil
	}
	badges, err := handler.Achievements.List(context.Background(), playerID)
	if err != nil {
		return err
	}
	encoded, err := outcurrent.Encode(playerID, activeBadges(badges))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), encoded)
}

// equip validates and replaces the authenticated player's active badge slots.
func (handler Handler) equip(connection netconn.Context, packet codec.Packet) error {
	requested, err := inequip.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Achievements == nil {
		return nil
	}
	badges, err := handler.Achievements.SetEquipped(context.Background(), playerID, requested[:])
	if err != nil && !errors.Is(err, ErrBadgeNotOwned) && !errors.Is(err, ErrInvalidBadge) {
		return err
	}
	if err != nil {
		return handler.sendInventory(context.Background(), connection, playerID)
	}
	if err = handler.sendBadgeInventory(context.Background(), connection, badges); err != nil {
		return err
	}
	encoded, err := outcurrent.Encode(playerID, activeBadges(badges))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), encoded)
}

// sendInventory loads and projects one player's badge inventory.
func (handler Handler) sendInventory(ctx context.Context, connection netconn.Context, playerID int64) error {
	badges, err := handler.Achievements.List(ctx, playerID)
	if err != nil {
		return err
	}
	return handler.sendBadgeInventory(ctx, connection, badges)
}

// sendBadgeInventory serializes one previously loaded badge snapshot.
func (handler Handler) sendBadgeInventory(ctx context.Context, connection netconn.Context, badges []Badge) error {
	values := make([]outlist.Badge, 0, len(badges))
	for _, badge := range badges {
		if badge.ID <= 0 || badge.ID > math.MaxInt32 {
			continue
		}
		values = append(values, outlist.Badge{ID: int32(badge.ID), Code: badge.Code, Slot: badge.Slot})
	}
	packet, err := outlist.Encode(values)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// playerID resolves an authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}

// activeBadges maps equipped domain badges into their public room projection.
func activeBadges(badges []Badge) []outcurrent.Badge {
	values := make([]outcurrent.Badge, 0, MaxEquippedBadges)
	for _, badge := range badges {
		if badge.Equipped && badge.Slot > 0 {
			values = append(values, outcurrent.Badge{Slot: badge.Slot, Code: badge.Code})
		}
	}
	return values
}
