// Package equip decodes active player badge replacements.
package equip

import (
	"errors"

	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies USER_BADGES_CURRENT_UPDATE requests.
const Header uint16 = 644

// SlotCount is Nitro's fixed number of wearable badge positions.
const SlotCount = 5

// ErrInvalidSlot reports a duplicate or out-of-range badge position.
var ErrInvalidSlot = errors.New("invalid badge slot")

// Definition describes Nitro's five slot and badge-code pairs.
var Definition = codec.Definition{
	codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField,
	codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField,
	codec.Int32Field, codec.StringField,
}

// Decode returns badge codes ordered by their requested slot.
func Decode(packet codec.Packet) ([SlotCount]string, error) {
	var badges [SlotCount]string
	if packet.Header != Header {
		return badges, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return badges, err
	}
	var occupied [SlotCount]bool
	for index := 0; index < len(values); index += 2 {
		slot := values[index].Int32
		if slot < 1 || slot > SlotCount || occupied[slot-1] {
			return badges, ErrInvalidSlot
		}
		occupied[slot-1] = true
		badges[slot-1] = values[index+1].String
	}
	return badges, nil
}
