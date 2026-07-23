// Package info encodes BONUS_RARE_INFO responses.
package info

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BONUS_RARE_INFO.
const Header uint16 = 1533

// Definition describes Bonus Rare reward and progress fields.
var Definition = codec.Definition{
	codec.Named("productType", codec.StringField),
	codec.Named("productClassId", codec.Int32Field),
	codec.Named("totalCoinsForBonus", codec.Int32Field),
	codec.Named("coinsStillRequiredToBuy", codec.Int32Field),
}

// Encode creates one BONUS_RARE_INFO packet.
func Encode(productType string, productClassID int32, threshold int32, remaining int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(productType), codec.Int32(productClassID), codec.Int32(threshold), codec.Int32(remaining))
}
