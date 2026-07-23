// Package init contains INIT_CAMERA.
package init

import "github.com/niflaot/pixels/networking/codec"

// Header identifies INIT_CAMERA.
const Header uint16 = 3878

// Definition describes camera price fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}

// Encode creates current camera prices.
func Encode(creditPrice int32, pointsPrice int32, publishPointsPrice int32) (codec.Packet, error) {
	if creditPrice < 0 || pointsPrice < 0 || publishPointsPrice < 0 {
		return codec.Packet{}, codec.ErrInvalidField
	}
	return codec.NewPacket(Header, Definition, codec.Int32(creditPrice), codec.Int32(pointsPrice), codec.Int32(publishPointsPrice))
}
