package floorplan

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outblocked "github.com/niflaot/pixels/networking/outbound/room/floorplan/blockedtiles"
)

const (
	// BlockedTilesName identifies the occupied floor plan tiles command.
	BlockedTilesName command.Name = "room.floorplan.blocked_tiles"
)

// BlockedTilesCommand requests currently occupied floor plan tiles.
type BlockedTilesCommand struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// BlockedTilesHandler handles occupied tile requests.
type BlockedTilesHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms reads room records.
	Rooms RoomFinder
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Authorize resolves floor plan capability.
	Authorize *domain.Authorizer
}

// CommandName returns the stable command name.
func (BlockedTilesCommand) CommandName() command.Name { return BlockedTilesName }

// Handle sends every furniture-occupied tile in stable order.
func (handler BlockedTilesHandler) Handle(ctx context.Context, envelope command.Envelope[BlockedTilesCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil || !found {
		return err
	}
	if err = handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
		return sendError(ctx, envelope.Command.Handler, err, nil)
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}
	points := domain.OccupiedTiles(active.FurnitureItems())
	tiles := make([]outblocked.Tile, len(points))
	for index, point := range points {
		tiles[index] = outblocked.Tile{X: int32(point.X), Y: int32(point.Y)}
	}
	packet, err := outblocked.Encode(tiles)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
