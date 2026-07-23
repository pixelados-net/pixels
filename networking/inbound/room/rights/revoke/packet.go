// Package revoke contains the ROOM_RIGHTS_REMOVE inbound packet.
package revoke

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_REMOVE.
	Header uint16 = 2064
	// maxPlayers caps one rights revoke packet.
	maxPlayers int32 = 100
)

// Payload contains unpacked rights revoke fields.
type Payload struct {
	// PlayerIDs identifies targets to revoke.
	PlayerIDs []int32
}

// CountDefinition describes the target count.
var CountDefinition = codec.Definition{codec.Named("playerCount", codec.Int32Field)}

// PlayerDefinition describes one target player id.
var PlayerDefinition = codec.Definition{codec.Named("playerId", codec.Int32Field)}

// Decode unpacks a ROOM_RIGHTS_REMOVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, CountDefinition)
	if err != nil || values[0].Int32 < 0 || values[0].Int32 > maxPlayers {
		return Payload{}, packetError(err)
	}
	ids := make([]int32, 0, values[0].Int32)
	for range values[0].Int32 {
		values, rest, err = codec.DecodePayload(values[:0], PlayerDefinition, rest)
		if err != nil {
			return Payload{}, err
		}
		ids = append(ids, values[0].Int32)
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}

	return Payload{PlayerIDs: ids}, nil
}

// packetError returns a decode or invalid-field error.
func packetError(err error) error {
	if err != nil {
		return err
	}

	return codec.ErrInvalidField
}
