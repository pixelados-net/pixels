// Package stackheight encodes ITEM_STACK_HELPER responses.
package stackheight

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ITEM_STACK_HELPER.
const Header uint16 = 2816

// Definition describes the effective height in centimeters.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("height", codec.Int32Field)}

// Encode creates one effective stack-height response.
func Encode(itemID int32, height int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(itemID), codec.Int32(height))
}
