// Package reportim contains messenger CALL_FOR_HELP decoding.
package reportim

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CALL_FOR_HELP_FROM_IM.
const Header uint16 = 2950

// Entry stores one private-chat evidence pair.
type Entry struct {
	// PlayerID identifies the author of the selected private message.
	PlayerID int32
	// Message stores the selected private-chat text.
	Message string
}

// Payload contains one messenger report.
type Payload struct {
	Message          string
	TopicID          int32
	ReportedPlayerID int32
	Entries          []Entry
}

// Decode reads fixed report fields and evidence pairs.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Uint16Field}, packet.Payload)
	if err != nil || values[3].Uint16 > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{Message: values[0].String, TopicID: values[1].Int32, ReportedPlayerID: values[2].Int32, Entries: make([]Entry, int(values[3].Uint16))}
	for i := range result.Entries {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field, codec.StringField}, rest)
		if err != nil {
			return Payload{}, err
		}
		result.Entries[i] = Entry{PlayerID: values[0].Int32, Message: values[1].String}
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return result, nil
}
