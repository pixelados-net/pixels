// Package notfound contains the GIFT_RECEIVER_NOT_FOUND outbound packet.
package notfound

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GIFT_RECEIVER_NOT_FOUND.
	Header uint16 = 1517
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.StringField}

// Encode creates a GIFT_RECEIVER_NOT_FOUND packet.
func Encode(username string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(username))
}
