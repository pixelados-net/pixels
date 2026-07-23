// Package request contains the REQUEST_USER_CREDITS inbound packet.
package request

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the REQUEST_USER_CREDITS packet identifier.
	Header uint16 = 273
)

// Payload contains the unpacked REQUEST_USER_CREDITS fields.
type Payload struct{}

// Definition describes the REQUEST_USER_CREDITS payload fields.
var Definition = codec.Definition{}

// Decode unpacks a REQUEST_USER_CREDITS packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}

	return Payload{}, nil
}
