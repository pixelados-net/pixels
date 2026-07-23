package send

import (
	"context"

	roomcontrol "github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outtyping "github.com/niflaot/pixels/networking/outbound/chat/typing"
)

// Typing broadcasts one player's typing state to the other room occupants.
func (service *Service) Typing(ctx context.Context, connection netconn.Context, activeState bool) error {
	player, roomID, err := roomcontrol.Actor(connection, service.bindings, service.players)
	if err != nil {
		return err
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	unit, found := active.Unit(player.ID())
	if !found {
		return roomlive.ErrUnitNotFound
	}
	packet, err := outtyping.Encode(int32(unit.UnitID), activeState)
	if err != nil {
		return err
	}
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID != player.ID() && service.canReceive(occupant.PlayerID, player.ID()) {
			service.sendPresence(ctx, occupant, packet)
		}
	}

	return nil
}
