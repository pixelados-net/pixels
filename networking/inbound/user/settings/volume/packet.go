// Package volume contains the USER_SETTINGS_VOLUME inbound packet.
package volume

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_SETTINGS_VOLUME.
const Header uint16 = 1367

// Definition describes USER_SETTINGS_VOLUME fields.
var Definition = codec.Definition{codec.Named("system", codec.Int32Field), codec.Named("furniture", codec.Int32Field), codec.Named("trax", codec.Int32Field)}

// Payload contains decoded volume settings.
type Payload struct {
	// System stores UI/system volume.
	System int32
	// Furniture stores furniture sound volume.
	Furniture int32
	// Trax stores music volume.
	Trax int32
}

// Decode decodes USER_SETTINGS_VOLUME.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{System: values[0].Int32, Furniture: values[1].Int32, Trax: values[2].Int32}, nil
}
