package list

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode writes every USER_BOTS inventory field.
func TestEncode(t *testing.T) {
	packet, err := Encode([]Bot{{ID: 7, Name: "Frank", Motto: "Tea?", Gender: "m", Figure: "hd-180-1"}})
	values, rest, decodeErr := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	bot, rest, botErr := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField}, rest)
	if err != nil || decodeErr != nil || botErr != nil || len(rest) != 0 || values[0].Int32 != 1 || bot[0].Int32 != 7 || bot[1].String != "Frank" || bot[4].String != "hd-180-1" {
		t.Fatalf("count=%#v bot=%#v rest=%d errors=%v/%v/%v", values, bot, len(rest), err, decodeErr, botErr)
	}
}
