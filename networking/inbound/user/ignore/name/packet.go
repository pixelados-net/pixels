// Package name decodes USER_IGNORE username requests.
package name

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_IGNORE.
const Header uint16 = 1117

// Decode unpacks the target username.
func Decode(packet codec.Packet) (string, error) {
	if packet.Header != Header {
		return "", codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField})
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
