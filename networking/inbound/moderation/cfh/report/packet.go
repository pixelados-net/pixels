// Package report contains CALL_FOR_HELP decoding.
package report

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CALL_FOR_HELP.
const Header uint16 = 1691

// Entry stores one client-selected evidence pair.
type Entry struct {
	Pattern string
	Message string
}

// Payload contains one room call for help.
type Payload struct {
	Message          string
	TopicID          int32
	ReportedPlayerID int32
	RoomID           int32
	Entries          []Entry
}

// Decode reads fixed report fields and bounded evidence pairs.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || values[4].Int32 < 0 || values[4].Int32 > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{Message: values[0].String, TopicID: values[1].Int32, ReportedPlayerID: values[2].Int32, RoomID: values[3].Int32, Entries: make([]Entry, values[4].Int32)}
	for i := range result.Entries {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.StringField, codec.StringField}, rest)
		if err != nil {
			return Payload{}, err
		}
		result.Entries[i] = Entry{Pattern: values[0].String, Message: values[1].String}
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return result, nil
}
