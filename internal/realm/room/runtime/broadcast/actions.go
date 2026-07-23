package broadcast

import (
	"context"
	"strconv"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	outdance "github.com/niflaot/pixels/networking/outbound/room/entities/dance"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
	outidle "github.com/niflaot/pixels/networking/outbound/room/entities/idle"
)

// PacketSender sends one encoded room packet to a connection.
type PacketSender interface {
	// Send writes one packet through the underlying connection.
	Send(context.Context, codec.Packet) error
}

// SendRoomActions sends existing dance, effect, and idle projections directly to one room entrant.
func SendRoomActions(ctx context.Context, connection PacketSender, active *roomlive.Room, playerIDs ...int64) error {
	if active == nil {
		return nil
	}
	allowed := make(map[int64]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		allowed[playerID] = struct{}{}
	}
	for _, unit := range active.Units() {
		if len(allowed) > 0 {
			if _, found := allowed[unit.PlayerID]; !found {
				continue
			}
		}
		packets, err := unitActionPackets(unit)
		if err != nil {
			return err
		}
		for _, packet := range packets {
			if err = connection.Send(ctx, packet); err != nil {
				return err
			}
		}
	}
	return nil
}

// unitActionPackets encodes the persistent room actions for one unit snapshot.
func unitActionPackets(unit roomlive.UnitSnapshot) ([]codec.Packet, error) {
	packets := make([]codec.Packet, 0, 3)
	for _, status := range unit.Statuses {
		if status.Key != worldunit.StatusDance {
			continue
		}
		danceID, err := strconv.ParseInt(status.Value, 10, 32)
		if err != nil {
			continue
		}
		packet, err := outdance.Encode(unit.UnitID, int32(danceID))
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}
	if effectID := projection.EffectID(unit); effectID > 0 {
		packet, err := outeffect.Encode(unit.UnitID, effectID, 0)
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}
	if unit.Idle {
		packet, err := outidle.Encode(unit.UnitID, true)
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}
	return packets, nil
}
