// Package thumbnail contains the RENDER_ROOM_THUMBNAIL inbound packet.
package thumbnail

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies RENDER_ROOM_THUMBNAIL.
	Header uint16 = 1982
	// MaxBytes bounds untrusted thumbnail payloads.
	MaxBytes int32 = 1024 * 1024
)

// Payload contains one client-rendered PNG without copying its bytes.
type Payload struct {
	// PNG stores the raw client-rendered PNG bytes.
	PNG []byte
}

// Decode validates and decodes one bounded thumbnail render.
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
