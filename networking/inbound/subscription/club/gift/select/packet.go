// Package selectgift contains the CATALOG_SELECT_VIP_GIFT inbound packet.
package selectgift

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CATALOG_SELECT_VIP_GIFT.
	Header uint16 = 2276
)

// Payload contains a selected club gift.
type Payload struct {
	// GiftName identifies the catalog product code.
	GiftName string
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.StringField}

// Decode decodes a CATALOG_SELECT_VIP_GIFT packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{GiftName: v[0].String}, nil
}
