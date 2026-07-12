package invite

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesSenderAndMessage verifies room-invite encoding.
func TestEncodeWritesSenderAndMessage(t *testing.T) {
	packet, err := Encode(7, "join")
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField})
	if err != nil || decodeErr != nil || values[0].Int32 != 7 || values[1].String != "join" {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
