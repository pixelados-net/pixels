// Package room owns social-group room policy and group-furniture packet adapters.
package room

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininfo "github.com/niflaot/pixels/networking/inbound/room/furniture/group/info"
	outcontext "github.com/niflaot/pixels/networking/outbound/room/furniture/group/context"
)

// Handler contains group-furniture context dependencies.
type Handler struct {
	// Cache resolves warmed room, player, and item group generations.
	Cache *groupruntime.Cache
	// Delivery resolves authenticated actors.
	Delivery *groupruntime.Delivery
	// Rooms resolves the actor's authoritative active room.
	Rooms *roomlive.Registry
}

// RegisterHandlers registers group-furniture context requests.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(ininfo.Header, handler.info)
}

// info sends one warmed group-furniture context response.
func (handler Handler) info(connection netconn.Context, packet codec.Packet) error {
	payload, err := ininfo.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	if handler.Cache == nil || handler.Rooms == nil {
		return nil
	}
	active, found := handler.Rooms.FindByPlayer(playerID)
	if !found {
		return nil
	}
	room, found := handler.Cache.Room(active.ID())
	if !found {
		return nil
	}
	groupID, found := handler.Cache.FurnitureGroup(payload.ObjectID)
	if !found || groupID != room.Group.Group.ID || payload.GroupID > 0 && payload.GroupID != groupID {
		return nil
	}
	player, loaded := handler.Cache.Player(playerID)
	member := false
	role := grouprecord.Member
	if loaded {
		role, member = player.Role(groupID)
	}
	group := room.Group.Group
	forumReadable := group.ForumEnabled && (group.ReadPolicy == grouprecord.Everyone || member && (group.ReadPolicy == grouprecord.Members || group.ReadPolicy == grouprecord.Admins && role <= grouprecord.Admin || group.ReadPolicy == grouprecord.Owners && role == grouprecord.Owner))
	response, err := outcontext.Encode(payload.ObjectID, group.ID, group.Name, group.HomeRoomID, member, forumReadable)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}
