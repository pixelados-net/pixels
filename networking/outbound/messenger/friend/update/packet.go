// Package update contains the MESSENGER_UPDATE outbound packet.
package update

import (
	"github.com/niflaot/pixels/networking/codec"
	friendcard "github.com/niflaot/pixels/networking/outbound/messenger/friend/card"
)

// Header identifies MESSENGER_UPDATE.
const Header uint16 = 2800

// Type identifies one Nitro friend-list update operation.
type Type int32

const (
	// Removed removes a friend card by player id.
	Removed Type = -1
	// Changed replaces an existing friend card.
	Changed Type = 0
	// Added inserts a new friend card.
	Added Type = 1
)

// Entry contains one friend-list update.
type Entry struct {
	// Type stores the update operation.
	Type Type
	// PlayerID identifies a removed friend.
	PlayerID int64
	// Card stores an added or changed friend.
	Card friendcard.Card
}

// Encode creates MESSENGER_UPDATE without deferred category changes.
func Encode(entries []Entry) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(0), codec.Int32(int32(len(entries))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, entry := range entries {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(entry.Type)))
		if err != nil {
			return codec.Packet{}, err
		}
		if entry.Type == Removed {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(entry.PlayerID)))
		} else {
			payload, err = friendcard.Append(payload, entry.Card)
		}
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
