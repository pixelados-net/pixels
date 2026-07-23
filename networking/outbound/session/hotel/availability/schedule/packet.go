// Package schedule contains the HOTEL_CLOSES_AND_OPENS_AT outbound packet.
package schedule

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HOTEL_CLOSES_AND_OPENS_AT packet identifier.
	Header uint16 = 2771
)

// Definition describes the HOTEL_CLOSES_AND_OPENS_AT payload fields.
var Definition = codec.Definition{
	codec.Named("openHour", codec.Int32Field),
	codec.Named("openMinute", codec.Int32Field),
	codec.Named("userThrownOutAtClose", codec.BooleanField),
}

// Encode creates a HOTEL_CLOSES_AND_OPENS_AT packet.
func Encode(openHour int32, openMinute int32, userThrownOutAtClose bool) (codec.Packet, error) {
	values := make([]codec.Value, 0, 3)
	values = append(values, codec.Int32(openHour))
	values = append(values, codec.Int32(openMinute))
	values = append(values, codec.Bool(userThrownOutAtClose))

	return codec.NewPacket(Header, Definition, values...)
}
