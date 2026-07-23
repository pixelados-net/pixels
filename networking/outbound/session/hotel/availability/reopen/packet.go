// Package reopen contains the HOTEL_CLOSED_AND_OPENS outbound packet.
package reopen

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HOTEL_CLOSED_AND_OPENS packet identifier.
	Header uint16 = 3728
)

// Definition describes the HOTEL_CLOSED_AND_OPENS payload fields.
var Definition = codec.Definition{
	codec.Named("openHour", codec.Int32Field),
	codec.Named("openMinute", codec.Int32Field),
}

// Encode creates a HOTEL_CLOSED_AND_OPENS packet.
func Encode(openHour int32, openMinute int32) (codec.Packet, error) {
	values := make([]codec.Value, 0, 2)
	values = append(values, codec.Int32(openHour))
	values = append(values, codec.Int32(openMinute))

	return codec.NewPacket(Header, Definition, values...)
}
