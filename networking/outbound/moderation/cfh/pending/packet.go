// Package pending contains CFH_PENDING_CALLS projection.
package pending

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CFH_PENDING_CALLS.
const Header uint16 = 1121

// Call stores one unresolved report summary.
type Call struct {
	ID        string
	Timestamp string
	Message   string
}

// Encode creates the unresolved-call list.
func Encode(calls []Call) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(calls))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, call := range calls {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField, codec.StringField}, codec.String(call.ID), codec.String(call.Timestamp), codec.String(call.Message))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
