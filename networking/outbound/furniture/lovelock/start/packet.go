// Package start encodes LOVELOCK_FURNI_START responses.
package start

import "github.com/niflaot/pixels/networking/codec"

// Header identifies LOVELOCK_FURNI_START.
const Header uint16 = 3753

// Definition describes the renderer parser exactly.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("start", codec.BooleanField)}

// Encode creates a lovelock start or reset response.
func Encode(itemID int32, started bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(itemID), codec.Bool(started))
}
