// Package inviteerror contains MESSENGER_INVITE_ERROR.
package inviteerror

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_INVITE_ERROR.
const Header uint16 = 462

// Encode creates one invite failure group.
func Encode(errorCode int32, playerIDs []int64) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(errorCode), codec.Int32(int32(len(playerIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, playerID := range playerIDs {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(playerID)))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
