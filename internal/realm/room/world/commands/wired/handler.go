package wired

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler executes WIRED editor commands.
type Handler struct {
	// Config stores WIRED editor limits.
	Config roomwired.Config
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated session bindings.
	Bindings *binding.Registry
	// Rooms stores active rooms.
	Rooms *roomlive.Registry
	// Store persists configurations.
	Store record.Store
	// Registry resolves behavior descriptors.
	Registry *registry.Registry
	// Compiler validates configurations.
	Compiler *configuration.Compiler
	// Engine reloads active generations.
	Engine *wiredruntime.Engine
	// Permissions resolves staff overrides.
	Permissions permissionservice.Checker
	// ConfigureAny stores the staff override node.
	ConfigureAny string
	// Superwired stores the administrative compatibility node.
	Superwired string
	// Translations resolves localized feedback.
	Translations i18n.Translator
}

// authorizeSuperwired restricts privileged progression effects to explicit staff policy.
func (handler Handler) authorizeSuperwired(ctx context.Context, playerID int64) error {
	if handler.Permissions == nil || handler.Superwired == "" {
		return ErrNoRights
	}
	allowed, err := handler.Permissions.HasPermission(ctx, playerID, permission.Node(handler.Superwired))
	if err != nil {
		return err
	}
	if !allowed {
		return ErrNoRights
	}
	return nil
}

// actor resolves the authenticated player and active room.
func (handler Handler) actor(connection netconn.Context) (*playerlive.Player, *roomlive.Room, int64, error) {
	bindingValue, found := handler.Bindings.FindByConnection(binding.ConnectionKey{ID: connection.ConnectionID, Kind: connection.ConnectionKind})
	if !found {
		return nil, nil, 0, binding.ErrBindingNotFound
	}
	player, found := handler.Players.Find(bindingValue.PlayerID)
	if !found {
		return nil, nil, 0, binding.ErrBindingNotFound
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil, nil, 0, ErrNotInRoom
	}
	active, found := handler.Rooms.Find(roomID)
	if !found {
		return nil, nil, 0, roomlive.ErrRoomNotFound
	}
	return player, active, roomID, nil
}

// authorize verifies local rights or a global staff override.
func (handler Handler) authorize(ctx context.Context, playerID int64, active *roomlive.Room) error {
	if active.CanManageFurniture(playerID) {
		return nil
	}
	if handler.Permissions != nil && handler.ConfigureAny != "" {
		allowed, err := handler.Permissions.HasPermission(ctx, playerID, permission.Node(handler.ConfigureAny))
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
	}
	return ErrNoRights
}
