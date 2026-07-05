// Package willclose contains the HOTEL_WILL_CLOSE_MINUTES outbound packet.
package willclose

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HOTEL_WILL_CLOSE_MINUTES packet identifier.
	Header uint16 = 1050
)

// Definition describes the HOTEL_WILL_CLOSE_MINUTES payload fields.
var Definition = codec.Definition{
	codec.Named("minutes", codec.Int32Field),
}

// Encode creates a HOTEL_WILL_CLOSE_MINUTES packet.
func Encode(minutes int32) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.Int32(minutes))

	return codec.NewPacket(Header, Definition, values...)
}
