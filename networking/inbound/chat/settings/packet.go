// Package settings contains the GET_SOUND_SETTINGS inbound packet.
package settings

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_SOUND_SETTINGS.
	Header uint16 = 2388
)

// Definition describes the empty settings request payload.
var Definition = codec.Definition{}

// Decode validates a GET_SOUND_SETTINGS packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)

	return err
}
