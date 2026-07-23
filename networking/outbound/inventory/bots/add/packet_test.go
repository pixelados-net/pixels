package add

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/bots/list"
)

// TestEncode writes one inventory bot followed by the open flag.
func TestEncode(t *testing.T) {
	packet, err := Encode(outlist.Bot{ID: 7, Name: "Frank", Motto: "Tea?", Gender: "m", Figure: "hd-180-1"}, true)
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.BooleanField}
	values, decodeErr := codec.DecodePacketExact(packet, definition)
	if err != nil || decodeErr != nil || values[0].Int32 != 7 || !values[5].Boolean {
		t.Fatalf("values=%#v encode=%v decode=%v", values, err, decodeErr)
	}
}
