// Package blockedtiles contains the GET_OCCUPIED_TILES inbound packet.
package blockedtiles

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_OCCUPIED_TILES.
	Header uint16 = 1687
)

// Payload contains the empty occupied tile request.
type Payload struct{}

// Decode validates a GET_OCCUPIED_TILES packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}

	return Payload{}, nil
}
