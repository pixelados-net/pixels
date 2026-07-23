// Package grant contains the ROOM_RIGHTS_GIVE inbound packet.
package grant

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_GIVE.
	Header uint16 = 808
)

// Payload contains unpacked rights grant fields.
type Payload struct {
	// PlayerID identifies the target player.
	PlayerID int32
}

// Definition describes ROOM_RIGHTS_GIVE fields.
var Definition = codec.Definition{codec.Named("playerId", codec.Int32Field)}

// Decode unpacks a ROOM_RIGHTS_GIVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{PlayerID: values[0].Int32}, nil
}
