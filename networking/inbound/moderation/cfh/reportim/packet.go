// Package reportim contains messenger CALL_FOR_HELP decoding.
package reportim

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CALL_FOR_HELP_FROM_IM.
const Header uint16 = 2950

// Entry stores one private-chat evidence pair.
type Entry struct {
	Pattern string
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
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || values[3].Int32 < 0 || values[3].Int32 > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{Message: values[0].String, TopicID: values[1].Int32, ReportedPlayerID: values[2].Int32, Entries: make([]Entry, values[3].Int32)}
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
