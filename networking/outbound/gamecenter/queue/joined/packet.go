// Package joined encodes JOINEDQUEUEMESSAGE responses.
package joined

import "github.com/niflaot/pixels/networking/codec"

// Header identifies JOINEDQUEUEMESSAGE.
const Header uint16 = 2260

// Encode creates one JOINEDQUEUEMESSAGE response.
func Encode(gameTypeID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(gameTypeID))
}
