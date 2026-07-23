// Package clothing decodes the retired SET_CLOTHING_CHANGE_DATA request.
package clothing

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies SET_CLOTHING_CHANGE_DATA.
const Header uint16 = 924

// Payload contains the abandoned clothing-change fields.
type Payload struct {
	// ObjectID identifies the abandoned clothing furniture target.
	ObjectID int32
	// Gender stores the requested avatar gender.
	Gender string
	// Look stores the requested avatar figure.
	Look string
}

// Definition describes the retired clothing-change request.
var Definition = codec.Definition{codec.Named("objectId", codec.Int32Field), codec.Named("gender", codec.StringField), codec.Named("look", codec.StringField)}

// Decode returns the compatibility-only fields.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ObjectID: v[0].Int32, Gender: v[1].String, Look: v[2].String}, nil
}
