// Package charge encodes the unconsumed FIREWORK_CHARGE_DATA packet.
package charge

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FIREWORK_CHARGE_DATA.
const Header uint16 = 5210

// Definition documents the provisional golden-only item and charge state.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("charged", codec.BooleanField)}

// Encode creates the golden-only firework charge packet.
func Encode(itemID int32, charged bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(itemID), codec.Bool(charged))
}
