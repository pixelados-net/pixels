// Package objectdata decodes SET_OBJECT_DATA requests.
package objectdata

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies SET_OBJECT_DATA.
const Header uint16 = 3608

// Payload contains one furniture id and custom key/value data.
type Payload struct {
	// ObjectID identifies the floor furniture instance.
	ObjectID int32
	// Data stores detached custom values.
	Data map[string]string
}

// PrefixDefinition describes the object id and flattened string count.
var PrefixDefinition = codec.Definition{codec.Named("objectId", codec.Int32Field), codec.Named("valueCount", codec.Int32Field)}

// Decode returns the renderer's flattened even-sized key/value map.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, rest, err := codec.DecodePacket(packet, PrefixDefinition)
	if err != nil {
		return Payload{}, err
	}
	count := v[1].Int32
	if count < 0 || count%2 != 0 || count > 200 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	result := Payload{ObjectID: v[0].Int32, Data: make(map[string]string, count/2)}
	for index := int32(0); index < count; index += 2 {
		pair, next, decodeErr := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.StringField}, rest)
		if decodeErr != nil {
			return Payload{}, decodeErr
		}
		result.Data[pair[0].String] = pair[1].String
		rest = next
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return result, nil
}
