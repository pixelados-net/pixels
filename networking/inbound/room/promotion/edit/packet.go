// Package edit decodes EDIT_ROOM_EVENT requests.
package edit

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies EDIT_ROOM_EVENT.
const Header uint16 = 3991

// Payload contains editable promotion copy.
type Payload struct {
	EventID     int32
	Name        string
	Description string
}

// Definition describes edit fields.
var Definition = codec.Definition{codec.Named("eventId", codec.Int32Field), codec.Named("name", codec.StringField), codec.Named("description", codec.StringField)}

// Decode returns one promotion edit.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{EventID: v[0].Int32, Name: v[1].String, Description: v[2].String}, nil
}
