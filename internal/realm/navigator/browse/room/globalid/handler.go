// Package globalid resolves external room identifiers without leaking invisible rooms.
package globalid

import (
	"context"
	"strconv"
	"strings"

	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inglobal "github.com/niflaot/pixels/networking/inbound/navigator/browse/globalid"
	outconverted "github.com/niflaot/pixels/networking/outbound/navigator/browse/convertedid"
)

// Handler resolves global room links for authenticated players.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated sessions.
	Bindings *binding.Registry
	// Rooms reads durable room records.
	Rooms roomservice.Manager
	// Rights checks invisible-room access.
	Rights roomrights.Manager
}

// Register installs the global room identifier adapter.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry != nil {
		_ = registry.Register(inglobal.Header, handler.handle)
	}
}

// handle resolves one identifier and always returns a privacy-safe conversion.
func (handler Handler) handle(connection netconn.Context, packet codec.Packet) error {
	payload, err := inglobal.Decode(packet)
	if err != nil {
		return err
	}
	player, _, err := navsession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID := int64(0)
	parsed, valid := parse(payload.GlobalID)
	if valid {
		room, found, findErr := handler.Rooms.FindByID(context.Background(), parsed)
		if findErr != nil {
			return findErr
		}
		if found && handler.visible(context.Background(), room, player.ID()) {
			roomID = parsed
		}
	}
	response, err := outconverted.Encode(payload.GlobalID, int32(roomID))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// parse accepts numeric and room-prefixed identifiers emitted by active link bridges.
func parse(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if prefix, suffix, found := strings.Cut(value, ":"); found {
		if !strings.EqualFold(prefix, "room") {
			return 0, false
		}
		value = suffix
	}
	id, err := strconv.ParseInt(value, 10, 32)
	return id, err == nil && id > 0
}

// visible reports whether an invisible room may be disclosed to the actor.
func (handler Handler) visible(ctx context.Context, room roommodel.Room, playerID int64) bool {
	if room.DoorMode != roommodel.DoorModeInvisible || room.OwnerPlayerID == playerID {
		return true
	}
	allowed, err := handler.Rights.HasRights(ctx, room.ID, playerID)
	return err == nil && allowed
}
