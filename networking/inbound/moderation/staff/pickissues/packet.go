// Package pickissues contains the moderation issue claim packet.
package pickissues

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PICK_ISSUES.
const Header uint16 = 15

// Payload contains an atomic issue claim request.
type Payload struct {
	IssueIDs     []int32
	RetryEnabled bool
	RetryCount   int32
	Message      string
}

// Decode decodes a length-prefixed issue claim.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil || values[0].Int32 < 0 || values[0].Int32 > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{IssueIDs: make([]int32, values[0].Int32)}
	for index := range result.IssueIDs {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return Payload{}, err
		}
		result.IssueIDs[index] = values[0].Int32
	}
	values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.BooleanField, codec.Int32Field, codec.StringField}, rest)
	if err != nil || len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	result.RetryEnabled, result.RetryCount, result.Message = values[0].Boolean, values[1].Int32, values[2].String
	return result, nil
}
