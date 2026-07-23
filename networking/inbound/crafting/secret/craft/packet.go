// Package craft contains the CRAFT_SECRET inbound packet.
package craft

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CRAFT_SECRET.
	Header uint16 = 1251
	// MaxItems bounds allocation from an untrusted combination count.
	MaxItems int32 = 100
)

// Payload stores one free-form exact combination request.
type Payload struct {
	// AltarItemID identifies the placed altar instance.
	AltarItemID int64
	// ItemIDs stores the submitted inventory furniture instances.
	ItemIDs []int64
}

// Decode reads one bounded free-form combination request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	count := values[1].Int32
	if values[0].Int32 <= 0 || count <= 0 || count > MaxItems {
		return Payload{}, codec.ErrInvalidField
	}
	payload := Payload{AltarItemID: int64(values[0].Int32), ItemIDs: make([]int64, 0, count)}
	for range count {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil || values[0].Int32 <= 0 {
			if err == nil {
				err = codec.ErrInvalidField
			}
			return Payload{}, err
		}
		payload.ItemIDs = append(payload.ItemIDs, int64(values[0].Int32))
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return payload, nil
}
