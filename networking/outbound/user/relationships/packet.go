// Package relationships encodes MESSENGER_RELATIONSHIPS responses.
package relationships

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_RELATIONSHIPS.
const Header uint16 = 2016

// Entry stores one non-empty public relationship category.
type Entry struct {
	// Type identifies Nitro's relationship category.
	Type int32
	// Count stores assigned friend count.
	Count int32
	// FriendID identifies one representative friend.
	FriendID int64
	// FriendName stores the representative username.
	FriendName string
	// FriendLook stores the representative figure.
	FriendLook string
}

// Encode creates a public relationship summary packet.
func Encode(playerID int64, entries []Entry) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(playerID)), codec.Int32(int32(len(entries))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, entry := range entries {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField}, codec.Int32(entry.Type), codec.Int32(entry.Count), codec.Int32(int32(entry.FriendID)), codec.String(entry.FriendName), codec.String(entry.FriendLook))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
