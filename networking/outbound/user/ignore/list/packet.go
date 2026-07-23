// Package list encodes USER_IGNORED lists.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_IGNORED.
const Header uint16 = 126

// Encode creates an ignored-username list packet.
func Encode(usernames []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(usernames))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, username := range usernames {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(username))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
