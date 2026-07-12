// Package acceptresult contains MESSENGER_ACCEPT_FRIENDS.
package acceptresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_ACCEPT_FRIENDS.
const Header uint16 = 896

// Failure contains one request that could not be accepted.
type Failure struct {
	// PlayerID identifies the requester.
	PlayerID int64
	// ErrorCode stores Nitro's acceptance failure code.
	ErrorCode int32
}

// Encode creates MESSENGER_ACCEPT_FRIENDS.
func Encode(failures []Failure) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(failures))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, failure := range failures {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(failure.PlayerID)), codec.Int32(failure.ErrorCode))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
