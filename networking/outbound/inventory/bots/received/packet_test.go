package received

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/bots/list"
)

// TestEncode writes the gift flag before one received bot.
func TestEncode(t *testing.T) {
	packet, err := Encode(outlist.Bot{ID: 7, Name: "Frank", Motto: "Tea?", Gender: "m", Figure: "hd-180-1"}, true)
	definition := codec.Definition{codec.BooleanField, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField}
	values, decodeErr := codec.DecodePacketExact(packet, definition)
	if err != nil || decodeErr != nil || !values[0].Boolean || values[1].Int32 != 7 {
		t.Fatalf("values=%#v encode=%v decode=%v", values, err, decodeErr)
	}
}
