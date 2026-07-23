// Package style contains the USER_SETTINGS_CHAT_STYLE inbound packet.
package style

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies USER_SETTINGS_CHAT_STYLE.
	Header uint16 = 1030
)

// Definition describes the selected bubble style.
var Definition = codec.Definition{codec.Named("styleId", codec.Int32Field)}

// Decode decodes a USER_SETTINGS_CHAT_STYLE packet.
func Decode(packet codec.Packet) (int32, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}

	return values[0].Int32, nil
}
