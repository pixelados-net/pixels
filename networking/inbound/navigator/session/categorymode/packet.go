// Package categorymode contains the NAVIGATOR_CATEGORY_LIST_MODE inbound packet.
package categorymode

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_CATEGORY_LIST_MODE packet identifier.
	Header uint16 = 1202
)

// Payload contains the unpacked NAVIGATOR_CATEGORY_LIST_MODE fields.
type Payload struct {
	// Category identifies the category code.
	Category string
	// ListMode stores the requested display mode.
	ListMode int32
}

// Definition describes the NAVIGATOR_CATEGORY_LIST_MODE payload fields.
var Definition = codec.Definition{
	codec.Named("category", codec.StringField),
	codec.Named("listMode", codec.Int32Field),
}

// Decode unpacks a NAVIGATOR_CATEGORY_LIST_MODE packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Category: values[0].String, ListMode: values[1].Int32}, nil
}
