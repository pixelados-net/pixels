package update

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	friendcard "github.com/niflaot/pixels/networking/outbound/messenger/friend/card"
)

// TestEncodeWritesMixedUpdates verifies native update discriminators.
func TestEncodeWritesMixedUpdates(t *testing.T) {
	packet, err := Encode([]Entry{{Type: Removed, PlayerID: 7}, {Type: Added, Card: friendcard.Card{PlayerID: 8, Username: "alice"}}})
	values, rest, decodeErr := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[1].Int32 != 2 || values[2].Int32 != -1 || values[3].Int32 != 7 || len(rest) == 0 {
		t.Fatalf("unexpected values=%#v rest=%d err=%v decode=%v", values, len(rest), err, decodeErr)
	}
}
