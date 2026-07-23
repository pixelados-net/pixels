// Package save contains the SAVE_WARDROBE_OUTFIT inbound packet.
package save

import "github.com/niflaot/pixels/networking/codec"

// Header identifies SAVE_WARDROBE_OUTFIT.
const Header uint16 = 800

// Definition describes SAVE_WARDROBE_OUTFIT fields.
var Definition = codec.Definition{codec.Named("slotId", codec.Int32Field), codec.Named("figure", codec.StringField), codec.Named("gender", codec.StringField)}

// Payload contains decoded wardrobe fields.
type Payload struct {
	// SlotID identifies the wardrobe slot.
	SlotID int32
	// Figure stores the outfit figure.
	Figure string
	// Gender stores the outfit gender code.
	Gender string
}

// Decode decodes SAVE_WARDROBE_OUTFIT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{SlotID: values[0].Int32, Figure: values[1].String, Gender: values[2].String}, nil
}
