// Package banlist contains the ROOM_BAN_LIST outbound packet.
package banlist

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_BAN_LIST.
	Header uint16 = 1869
)

// Ban contains one active room ban.
type Ban struct {
	// PlayerID identifies the banned player.
	PlayerID int32
	// Username stores the banned player username.
	Username string
}

// Definition describes ban-list metadata.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("count", codec.Int32Field)}

// BanDefinition describes one banned player.
var BanDefinition = codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("username", codec.StringField)}

// Encode creates a ROOM_BAN_LIST packet.
func Encode(roomID int32, bans []Ban) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(roomID), codec.Int32(int32(len(bans))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, ban := range bans {
		payload, err = codec.AppendPayload(payload, BanDefinition, codec.Int32(ban.PlayerID), codec.String(ban.Username))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
