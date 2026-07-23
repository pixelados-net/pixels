// Package profile contains USER_PROFILE requests.
package profile

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_PROFILE.
const Header uint16 = 3265

// Payload contains unpacked USER_PROFILE fields.
type Payload struct {
	// PlayerID identifies the requested profile.
	PlayerID int64
	// OpenWindow reports whether Nitro should open the profile window.
	OpenWindow bool
}

// Decode unpacks USER_PROFILE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.BooleanField})
	if err != nil {
		return Payload{}, err
	}
	return Payload{PlayerID: int64(values[0].Int32), OpenWindow: values[1].Boolean}, nil
}
