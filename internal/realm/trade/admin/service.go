// Package admin implements protected direct-trade administration.
package admin

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Service manages trade locks and audit reads.
type Service struct {
	// players persists direct-trade eligibility.
	players playerservice.TradeManager
	// live updates and locates authenticated players.
	live *playerlive.Registry
	// connections delivers an optional lock notice.
	connections *netconn.Registry
	// translations resolves hotel-facing lock text.
	translations i18n.Translator
	// store reads completed trade audits.
	store traderecord.Store
}

// New creates protected trade administration.
func New(players playerservice.TradeManager, live *playerlive.Registry, connections *netconn.Registry, translations i18n.Translator, store traderecord.Store) *Service {
	return &Service{players: players, live: live, connections: connections, translations: translations, store: store}
}

// SetLocked replaces a player's durable trade lock and live projection.
func (service *Service) SetLocked(ctx context.Context, playerID int64, locked bool) error {
	if err := service.players.SetAllowTrade(ctx, playerID, !locked); err != nil {
		return err
	}
	if player, found := service.live.Find(playerID); found {
		player.SetAllowTrade(!locked)
		if locked {
			service.sendLockedNotice(ctx, player)
		}
	}
	return nil
}

// sendLockedNotice delivers a localized best-effort alert after the lock commits.
func (service *Service) sendLockedNotice(ctx context.Context, player *playerlive.Player) {
	if service.connections == nil || player == nil {
		return
	}
	message := "Your direct trading ability has been disabled by hotel staff."
	if service.translations != nil {
		message = service.translations.Default("trade.moderation.locked")
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return
	}
	peer := player.Peer()
	connection, found := service.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if found {
		_ = connection.Send(ctx, packet)
	}
}

// Logs returns recent completed trades involving one player.
func (service *Service) Logs(ctx context.Context, playerID int64) ([]traderecord.Audit, error) {
	return service.store.ListAudits(ctx, playerID, 100)
}
