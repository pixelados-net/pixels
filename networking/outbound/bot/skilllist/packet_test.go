package skilllist

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode writes a variable bot skill list in Nitro order.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, []Skill{{ID: 1, Data: "a"}, {ID: 9, Data: "b"}})
	values, rest, decodeErr := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != -7 || values[1].Int32 != 2 {
		t.Fatalf("values=%#v encode=%v decode=%v", values, err, decodeErr)
	}
	first, rest, decodeErr := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, rest)
	second, rest, secondErr := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, rest)
	if decodeErr != nil || secondErr != nil || len(rest) != 0 || first[0].Int32 != 1 || second[0].Int32 != 9 {
		t.Fatalf("first=%#v second=%#v rest=%d errors=%v/%v", first, second, len(rest), decodeErr, secondErr)
	}
}
