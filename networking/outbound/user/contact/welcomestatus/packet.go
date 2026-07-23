// Package welcomestatus contains the retired WELCOME_GIFT_STATUS packet.
package welcomestatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WELCOME_GIFT_STATUS.
const Header uint16 = 2707

// Definition describes WELCOME_GIFT_STATUS fields.
var Definition = codec.Definition{
	codec.Named("email", codec.StringField),
	codec.Named("verified", codec.BooleanField),
	codec.Named("allowChange", codec.BooleanField),
	codec.Named("furnitureId", codec.Int32Field),
	codec.Named("requestedByUser", codec.BooleanField),
}

// Encode creates a retired WELCOME_GIFT_STATUS packet.
//
// Deprecated: the legacy welcome-gift journey is intentionally retired.
func Encode(email string, verified bool, allowChange bool, furnitureID int32, requestedByUser bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(email), codec.Bool(verified), codec.Bool(allowChange), codec.Int32(furnitureID), codec.Bool(requestedByUser))
}
