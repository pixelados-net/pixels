// Package entry encodes COMPETITION_ENTRY_SUBMIT responses.
package entry

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMPETITION_ENTRY_SUBMIT.
const Header uint16 = 1177

// PrerequisitesNotMet identifies an unavailable competition submission.
const PrerequisitesNotMet int32 = 3

// Encode creates one competition submission result.
func Encode(goalID int32, goalCode string, result int32, required []string, missing []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field}, codec.Int32(goalID), codec.String(goalCode), codec.Int32(result), codec.Int32(int32(len(required))))
	for _, value := range required {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(value))
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(missing))))
	}
	for _, value := range missing {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(value))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
