package search

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeSeparatesFriendsAndOthers verifies native search collections.
func TestEncodeSeparatesFriendsAndOthers(t *testing.T) {
	packet, err := Encode([]Result{{PlayerID: 1, Username: "demo"}}, []Result{{PlayerID: 2, Username: "alice"}})
	values, rest, decodeErr := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != 1 || len(rest) == 0 {
		t.Fatalf("unexpected values=%#v rest=%d err=%v decode=%v", values, len(rest), err, decodeErr)
	}
}
