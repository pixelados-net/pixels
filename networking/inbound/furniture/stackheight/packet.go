// Package stackheight decodes ITEM_STACK_HELPER requests.
package stackheight

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ITEM_STACK_HELPER.
const Header uint16 = 3839

// AutoHeight is the renderer sentinel for automatic stacking.
const AutoHeight int32 = -100

// Payload contains one stack-height override in centimeters.
type Payload struct {
	// ItemID identifies the helper furniture instance.
	ItemID int32
	// Height stores centimeters or AutoHeight.
	Height int32
}

// Definition describes a stack-height override.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("height", codec.Int32Field)}

// Decode returns one stack-height override.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Height: values[1].Int32}, nil
}
