// Package render contains the RENDER_ROOM inbound packet.
package render

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies RENDER_ROOM.
	Header uint16 = 3226
	// MaxBytes bounds untrusted photo payloads before domain policy runs.
	MaxBytes int32 = 2 * 1024 * 1024
)

// Payload contains one client-rendered PNG without copying its bytes.
type Payload struct {
	// PNG stores the raw client-rendered PNG bytes.
	PNG []byte
}

// Decode validates and decodes one bounded room render.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	size := values[0].Int32
	if size <= 0 || size > MaxBytes || len(rest) != int(size) {
		return Payload{}, codec.ErrInvalidField
	}
	return Payload{PNG: rest}, nil
}
