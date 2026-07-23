// Package closeissue contains close-with-default-action requests.
package closeissue

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CLOSE_ISSUE_DEFAULT_ACTION.
const Header uint16 = 2717

// Payload contains action, issue ids, and sanction id.
type Payload struct {
	Action     int32
	IssueIDs   []int32
	SanctionID int32
}

// Decode reads the default-action request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || values[1].Int32 < 0 || values[1].Int32 > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{Action: values[0].Int32, IssueIDs: make([]int32, values[1].Int32)}
	for i := range result.IssueIDs {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return Payload{}, err
		}
		result.IssueIDs[i] = values[0].Int32
	}
	values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
	if err != nil || len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	result.SanctionID = values[0].Int32
	return result, nil
}
