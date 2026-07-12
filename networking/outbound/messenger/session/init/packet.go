// Package initmsg contains the MESSENGER_INIT outbound packet.
package initmsg

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_INIT.
const Header uint16 = 1605

// Encode creates MESSENGER_INIT without deferred friend categories.
func Encode(userLimit int32, normalLimit int32, extendedLimit int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	}, codec.Int32(userLimit), codec.Int32(normalLimit), codec.Int32(extendedLimit), codec.Int32(0))
}
