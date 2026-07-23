// Package time contains the AVAILABILITY_TIME outbound packet.
package time

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the AVAILABILITY_TIME packet identifier.
	Header uint16 = 600
)

// Definition describes the AVAILABILITY_TIME payload fields.
var Definition = codec.Definition{
	codec.Named("isOpen", codec.Int32Field),
	codec.Named("minutesUntilChange", codec.Int32Field),
}

// Encode creates a AVAILABILITY_TIME packet.
func Encode(isOpen int32, minutesUntilChange int32) (codec.Packet, error) {
	values := make([]codec.Value, 0, 2)
	values = append(values, codec.Int32(isOpen))
	values = append(values, codec.Int32(minutesUntilChange))

	return codec.NewPacket(Header, Definition, values...)
}
