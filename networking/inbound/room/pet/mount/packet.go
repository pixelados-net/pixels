// Package mount decodes PET_RIDE and PET_MOUNT requests.
package mount

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_RIDE and PET_MOUNT.
const Header uint16 = 1036

// Definition describes PET_MOUNT fields.
var Definition = codec.Definition{codec.Int32Field, codec.BooleanField}

// Payload contains one mount or dismount request.
type Payload struct {
	// PetID identifies the room pet.
	PetID int64
	// Mount reports whether to mount rather than dismount.
	Mount bool
}

// Decode decodes PET_MOUNT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PetID: int64(values[0].Int32), Mount: values[1].Boolean}, nil
}
