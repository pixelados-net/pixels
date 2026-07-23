// Package id decodes USER_IGNORE_ID requests.
package id

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_IGNORE_ID.
const Header uint16 = 3314

// Decode unpacks the target player id.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return int64(values[0].Int32), nil
}
