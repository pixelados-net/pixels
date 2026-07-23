package votes

import (
	"context"
	"errors"
	"fmt"

	roomvotecast "github.com/niflaot/pixels/internal/realm/room/control/events/votecast"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Service coordinates durable votes and live score projection.
type Service struct {
	// store persists votes.
	store Store
	// rooms reads room metadata.
	rooms RoomFinder
	// runtime reads active room occupants.
	runtime *roomlive.Registry
	// connections resolves occupant transports.
	connections *netconn.Registry
	// events publishes committed votes.
	events bus.Publisher
}

// New creates room vote behavior.
func New(store Store, rooms RoomFinder, runtime *roomlive.Registry, connections *netconn.Registry, events bus.Publisher) *Service {
	return &Service{store: store, rooms: rooms, runtime: runtime, connections: connections, events: events}
}

// Cast permanently upvotes a room once per player.
func (service *Service) Cast(ctx context.Context, roomID int64, playerID int64) (Mutation, error) {
	room, err := service.room(ctx, roomID)
	if err != nil {
		return Mutation{}, err
	}
	if playerID <= 0 {
		return Mutation{}, ErrInvalidPlayerID
	}
	if room.OwnerPlayerID == playerID {
		result := Mutation{Score: room.Score}
		return result, service.broadcast(ctx, room, result.Score)
	}
	result, err := service.store.Cast(ctx, roomID, playerID)
	if err != nil {
		return Mutation{}, fmt.Errorf("cast room vote: %w", err)
	}
	room.Score = result.Score
	var eventErr error
	if result.Inserted && service.events != nil {
		eventErr = service.events.Publish(ctx, bus.Event{Name: roomvotecast.Name, Payload: roomvotecast.Payload{RoomID: roomID, PlayerID: playerID, Score: result.Score}})
	}
	projectionErr := service.broadcast(ctx, room, result.Score)

	return result, errors.Join(eventErr, projectionErr)
}

// State reads one player's room score and eligibility.
func (service *Service) State(ctx context.Context, roomID int64, playerID int64) (State, error) {
	room, err := service.room(ctx, roomID)
	if err != nil {
		return State{}, err
	}
	if playerID <= 0 {
		return State{}, ErrInvalidPlayerID
	}
	if room.OwnerPlayerID == playerID {
		return State{Score: room.Score}, nil
	}
	voted, err := service.store.HasVote(ctx, roomID, playerID)
	if err != nil {
		return State{}, fmt.Errorf("read room vote state: %w", err)
	}

	return State{Score: room.Score, CanVote: !voted, Voted: voted}, nil
}

// List returns durable room votes.
func (service *Service) List(ctx context.Context, query Query) ([]Vote, error) {
	normalized, err := query.Normalize()
	if err != nil {
		return nil, err
	}
	if _, err := service.room(ctx, normalized.RoomID); err != nil {
		return nil, err
	}

	return service.store.List(ctx, normalized)
}

// room validates and resolves one room.
func (service *Service) room(ctx context.Context, roomID int64) (roommodel.Room, error) {
	if roomID <= 0 {
		return roommodel.Room{}, ErrInvalidRoomID
	}
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return roommodel.Room{}, fmt.Errorf("find vote room: %w", err)
	}
	if !found {
		return roommodel.Room{}, ErrRoomNotFound
	}

	return room, nil
}
