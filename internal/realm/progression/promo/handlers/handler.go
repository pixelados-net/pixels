// Package handlers adapts Nitro promotional badge claims.
package handlers

import (
	"context"
	"math"
	"strings"

	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	progressionpromo "github.com/niflaot/pixels/internal/realm/progression/promo"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inclaim "github.com/niflaot/pixels/networking/inbound/progression/promo/claim"
	instatus "github.com/niflaot/pixels/networking/inbound/progression/promo/status"
	outbadgeadd "github.com/niflaot/pixels/networking/outbound/progression/achievement/badgeadd"
	outstatus "github.com/niflaot/pixels/networking/outbound/progression/promo/status"
)

// Handler owns promotional badge requests.
type Handler struct {
	// Service owns claim eligibility and persistence.
	Service *progressionpromo.Service
	// Badges resolves awarded badge inventory rows.
	Badges *playerachievement.Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs promotion claim and status handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(inclaim.Header, handler.claim)
	_ = registry.Register(instatus.Header, handler.status)
}

// claim atomically claims a promotion and projects its badge immediately.
func (handler Handler) claim(connection netconn.Context, packet codec.Packet) error {
	code, err := inclaim.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	granted, err := handler.Service.Claim(context.Background(), playerID, code, false)
	if err != nil {
		granted, err = false, nil
	}
	if err != nil {
		return err
	}
	if granted {
		if badgeCode, found := handler.Service.BadgeCode(code); found {
			handler.sendBadge(context.Background(), connection, playerID, badgeCode)
		}
	}
	return handler.sendStatus(connection, code, granted)
}

// status reports whether one player already claimed a promotion.
func (handler Handler) status(connection netconn.Context, packet codec.Packet) error {
	code, err := instatus.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	fulfilled, err := handler.Service.Status(context.Background(), playerID, code)
	if err != nil {
		return err
	}
	return handler.sendStatus(connection, code, fulfilled)
}

// sendStatus sends one native promotion fulfillment response.
func (handler Handler) sendStatus(connection netconn.Context, code string, fulfilled bool) error {
	response, err := outstatus.Encode(code, fulfilled)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// sendBadge projects the newly created badge inventory row.
func (handler Handler) sendBadge(ctx context.Context, connection netconn.Context, playerID int64, requested string) {
	if handler.Badges == nil {
		return
	}
	badges, err := handler.Badges.List(ctx, playerID)
	if err != nil {
		return
	}
	for _, badge := range badges {
		if strings.EqualFold(badge.Code, requested) && badge.ID > 0 && badge.ID <= math.MaxInt32 {
			if packet, encodeErr := outbadgeadd.Encode(int32(badge.ID), badge.Code); encodeErr == nil {
				_ = connection.Send(ctx, packet)
			}
			return
		}
	}
}

// playerID resolves one authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}
