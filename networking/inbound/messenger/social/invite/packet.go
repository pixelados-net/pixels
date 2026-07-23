// Package invite contains SEND_ROOM_INVITE.
package invite

import "github.com/niflaot/pixels/networking/codec"

// Header identifies SEND_ROOM_INVITE.
const Header uint16 = 1276

// Payload contains invite recipients and message.
type Payload struct {
	// PlayerIDs identifies invitation recipients.
	PlayerIDs []int64
	// Message stores the invitation text.
	Message string
}

// Decode unpacks SEND_ROOM_INVITE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil {
		return Payload{}, err
	}
	count := int(values[0].Int32)
	if count < 0 || count > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{PlayerIDs: make([]int64, 0, count)}
	for range count {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return Payload{}, err
		}
		result.PlayerIDs = append(result.PlayerIDs, int64(values[0].Int32))
	}
	values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.StringField}, rest)
	if err != nil {
		return Payload{}, err
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	result.Message = values[0].String
	return result, nil
}
