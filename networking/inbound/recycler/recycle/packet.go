// Package recycle contains the RECYCLER_ITEMS inbound packet.
package recycle

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies RECYCLER_ITEMS.
	Header uint16 = 2771
	// MaxItems bounds allocation from an untrusted recycler count.
	MaxItems int32 = 100
)

// Payload stores one recycler batch request.
type Payload struct {
	// ItemIDs stores submitted inventory furniture instances.
	ItemIDs []int64
}

// Decode reads one bounded recycler batch request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	count := values[0].Int32
	if count <= 0 || count > MaxItems {
		return Payload{}, codec.ErrInvalidField
	}
	payload := Payload{ItemIDs: make([]int64, 0, count)}
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
