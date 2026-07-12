package friends

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	friendcard "github.com/niflaot/pixels/networking/outbound/messenger/friend/card"
)

// TestEncodeWritesFragmentMetadata verifies friend-list framing.
func TestEncodeWritesFragmentMetadata(t *testing.T) {
	packet, err := Encode(1, 0, []friendcard.Card{{PlayerID: 7, Username: "demo"}})
	values, _, decodeErr := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || packet.Header != Header || values[2].Int32 != 1 {
		t.Fatalf("unexpected packet=%#v values=%#v err=%v decode=%v", packet, values, err, decodeErr)
	}
}
