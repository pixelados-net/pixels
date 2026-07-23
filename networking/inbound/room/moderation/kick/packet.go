// Package kick contains the ROOM_KICK inbound packet.
package kick

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_KICK.
	Header uint16 = 1320
)

// Payload contains unpacked room kick fields.
type Payload struct {
	// PlayerID identifies the target player.
	PlayerID int32
}

// Definition describes ROOM_KICK fields.
var Definition = codec.Definition{codec.Named("playerId", codec.Int32Field)}

// Decode unpacks a ROOM_KICK packet.
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
