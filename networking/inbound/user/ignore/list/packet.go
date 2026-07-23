// Package list decodes USER_IGNORED list requests.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_IGNORED.
const Header uint16 = 3878

// Decode unpacks the requesting username carried by Nitro.
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
