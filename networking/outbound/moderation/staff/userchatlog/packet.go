// Package userchatlog contains moderator user chat history.
package userchatlog

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/moderation/chatrecord"
)

// Header identifies MODTOOL_USER_CHATLOG.
const Header uint16 = 3377

// Encode creates one user's grouped room chatlogs.
func Encode(playerID int32, username string, records []chatrecord.Record) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(playerID), codec.String(username), codec.Int32(int32(len(records))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, record := range records {
		payload, err = chatrecord.Append(payload, record)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
