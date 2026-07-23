// Package requests contains the MESSENGER_REQUESTS outbound packet.
package requests

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_REQUESTS.
const Header uint16 = 280

// Request contains one incoming friend request.
type Request struct {
	// PlayerID identifies the requester and request.
	PlayerID int64
	// Username stores the requester name.
	Username string
	// Look stores the requester avatar figure.
	Look string
}

// Encode creates MESSENGER_REQUESTS with Nitro's embedded pending count.
func Encode(totalRequests int32, requests []Request) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(totalRequests), codec.Int32(int32(len(requests))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, request := range requests {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}, codec.Int32(int32(request.PlayerID)), codec.String(request.Username), codec.String(request.Look))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
