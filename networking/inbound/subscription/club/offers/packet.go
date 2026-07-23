// Package offers contains the GET_CLUB_OFFERS inbound packet.
package offers

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_CLUB_OFFERS.
	Header uint16 = 3285
)

// Payload contains the requesting client window id.
type Payload struct {
	// WindowID identifies the originating client window.
	WindowID int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a GET_CLUB_OFFERS packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{WindowID: v[0].Int32}, nil
}
