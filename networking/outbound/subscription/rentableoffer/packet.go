// Package rentableoffer encodes RENTABLE_FURNI_RENT_OR_BUYOUT_OFFER responses.
package rentableoffer

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RENTABLE_FURNI_RENT_OR_BUYOUT_OFFER.
const Header uint16 = 35

// Definition describes the renderer's rentable product offer.
var Definition = codec.Definition{codec.Named("wall", codec.BooleanField), codec.Named("productName", codec.StringField), codec.Named("buyout", codec.BooleanField), codec.Named("credits", codec.Int32Field), codec.Named("points", codec.Int32Field), codec.Named("pointType", codec.Int32Field)}

// Encode creates one rentable product offer response.
func Encode(wall bool, productName string, buyout bool, credits int32, points int32, pointType int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(wall), codec.String(productName), codec.Bool(buyout), codec.Int32(credits), codec.Int32(points), codec.Int32(pointType))
}
