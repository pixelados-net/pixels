package list

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// TestEncode verifies fragment metadata and pet count.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, 1, []petdata.Pet{{ID: 7, Name: "Pixel", Figure: petdata.Figure{Color: "FFFFFF"}, Level: 1}})
	values, rest, decodeErr := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 2 || values[1].Int32 != 1 || values[2].Int32 != 1 || len(rest) == 0 {
		t.Fatalf("values=%#v rest=%d errors=%v/%v", values, len(rest), err, decodeErr)
	}
}
