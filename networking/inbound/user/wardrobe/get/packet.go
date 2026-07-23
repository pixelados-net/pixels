// Package get contains the GET_WARDROBE inbound packet.
package get

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_WARDROBE.
const Header uint16 = 2742

// Definition describes GET_WARDROBE fields.
var Definition = codec.Definition{codec.Named("pageId", codec.Int32Field)}

// Payload contains decoded wardrobe request fields.
type Payload struct {
	// PageID identifies the requested wardrobe page.
	PageID int32
}

// Decode decodes GET_WARDROBE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PageID: values[0].Int32}, nil
}
