package votes

import (
	"context"
	"fmt"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outscore "github.com/niflaot/pixels/networking/outbound/room/score"
)

// broadcast projects score and per-player eligibility to active occupants.
func (service *Service) broadcast(ctx context.Context, room roommodel.Room, score int) error {
	if service.runtime == nil || service.connections == nil {
		return nil
	}
	active, found := service.runtime.Find(room.ID)
	if !found {
		return nil
	}
	occupants := active.Occupants()
	if len(occupants) == 0 {
		return nil
	}
	ids := occupantIDs(occupants)
	voters, err := service.store.Existing(ctx, room.ID, ids)
	if err != nil {
		return fmt.Errorf("read active room voters: %w", err)
	}
	allowed, denied, err := scorePackets(score)
	if err != nil {
		return err
	}
	for _, occupant := range occupants {
		packet := allowed
		_, voted := voters[occupant.PlayerID]
		if voted || occupant.PlayerID == room.OwnerPlayerID {
			packet = denied
		}
		if connection, exists := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID); exists {
			_ = connection.Send(ctx, packet)
		}
	}

	return nil
}

// occupantIDs extracts player ids into one batch query input.
func occupantIDs(occupants []roomlive.Occupant) []int64 {
	ids := make([]int64, len(occupants))
	for index := range occupants {
		ids[index] = occupants[index].PlayerID
	}

	return ids
}

// scorePackets encodes reusable eligibility variants.
func scorePackets(score int) (codec.Packet, codec.Packet, error) {
	allowed, err := outscore.Encode(int32(score), true)
	if err != nil {
		return codec.Packet{}, codec.Packet{}, err
	}
	denied, err := outscore.Encode(int32(score), false)
	if err != nil {
		return codec.Packet{}, codec.Packet{}, err
	}

	return allowed, denied, nil
}
