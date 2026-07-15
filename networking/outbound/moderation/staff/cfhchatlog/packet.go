// Package cfhchatlog contains frozen issue chat evidence.
package cfhchatlog

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/moderation/chatrecord"
)

// Header identifies CFH_CHATLOG.
const Header uint16 = 607

// Encode creates one issue chatlog packet.
func Encode(issueID, reporterID, reportedID int32, record chatrecord.Record) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(issueID), codec.Int32(reporterID), codec.Int32(reportedID), codec.Int32(issueID))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = chatrecord.Append(payload, record)
	return codec.Packet{Header: Header, Payload: payload}, err
}
