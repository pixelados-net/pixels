// Package status encodes RENTABLE_SPACE_STATUS responses.
package status

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RENTABLE_SPACE_STATUS.
const Header uint16 = 3559

// Definition describes rentable-space state.
var Definition = codec.Definition{codec.Named("rented", codec.BooleanField), codec.Named("errorCode", codec.Int32Field), codec.Named("renterId", codec.Int32Field), codec.Named("renterName", codec.StringField), codec.Named("secondsRemaining", codec.Int32Field), codec.Named("price", codec.Int32Field)}

// Encode creates one rentable-space status response.
func Encode(rented bool, errorCode int32, renterID int32, renterName string, secondsRemaining int32, price int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(rented), codec.Int32(errorCode), codec.Int32(renterID), codec.String(renterName), codec.Int32(secondsRemaining), codec.Int32(price))
}
