// Package delete decodes ROOM_DELETE requests.
package delete

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_DELETE.
const Header uint16 = 532

// Definition describes the room id.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Decode returns the requested room id.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return v[0].Int32, nil
}
