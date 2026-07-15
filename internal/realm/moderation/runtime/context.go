// Package runtime provides shared live moderation packet dependencies.
package runtime

import (
	"context"
	"errors"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	"github.com/niflaot/pixels/internal/realm/moderation/guardian"
	"github.com/niflaot/pixels/internal/realm/moderation/guide"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Context composes dependencies shared by moderation protocol capabilities.
type Context struct {
	// Moderation coordinates CFH and issues.
	Moderation *moderationcore.Service
	// Sanctions applies global punishments.
	Sanctions *sanctioncore.Service
	// Guides owns helper sessions.
	Guides *guide.Manager
	// Guardians owns peer reviews.
	Guardians *guardian.Manager
	// Players locates live recipients.
	Players *playerlive.Registry
	// PlayerRecords reads durable player data.
	PlayerRecords playerservice.Finder
	// Rooms reads durable rooms.
	Rooms roomservice.ConfigManager
	// RoomsLive locates active rooms.
	RoomsLive *roomlive.Registry
	// History reads partitioned chat.
	History *history.Service
	// Bindings resolves authenticated actors.
	Bindings *binding.Registry
	// Connections sends push packets.
	Connections *netconn.Registry
	// Permissions authorizes staff and duty actions.
	Permissions permissionservice.Checker
	// Translations localizes hotel-facing text.
	Translations i18n.Translator
	// Events publishes completed guide lifecycle facts.
	Events bus.Publisher
}

// Actor resolves the authenticated source player id.
func (runtime *Context) Actor(connection netconn.Context) (int64, error) {
	value, found := runtime.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, errors.New("moderation player binding not found")
	}
	return value.PlayerID, nil
}

// SendTo performs one best-effort online player delivery.
func (runtime *Context) SendTo(ctx context.Context, playerID int64, packet codec.Packet) error {
	player, found := runtime.Players.Find(playerID)
	if !found {
		return nil
	}
	peer := player.Peer()
	connection, found := runtime.Connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil
	}
	return connection.Send(ctx, packet)
}

// Text resolves a localized key with a stable fallback.
func (runtime *Context) Text(key string) string {
	if runtime.Translations == nil {
		return key
	}
	return runtime.Translations.Default(i18n.Key(key))
}
