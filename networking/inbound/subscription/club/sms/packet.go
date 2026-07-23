// Package sms contains the GET_DIRECT_CLUB_BUY_AVAILABLE inbound packet.
package sms

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_DIRECT_CLUB_BUY_AVAILABLE.
	Header uint16 = 801
)

// Payload contains the requested SMS membership duration.
type Payload struct {
	// DurationDays stores requested membership days.
	DurationDays int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a GET_DIRECT_CLUB_BUY_AVAILABLE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{DurationDays: v[0].Int32}, nil
}
