// Package roominvites contains USER_SETTINGS_INVITES.
package roominvites

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_SETTINGS_INVITES.
const Header uint16 = 1086

// Decode unpacks the room-invite blocking preference.
func Decode(packet codec.Packet) (bool, error) {
	if packet.Header != Header {
		return false, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.BooleanField})
	if err != nil {
		return false, err
	}
	return values[0].Boolean, nil
}
