package status

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies UNIT_STATUS packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode([]Unit{{
		RoomIndex: 3, X: 1, Y: 2, Z: "0", HeadDirection: 4, BodyDirection: 4,
		Actions: []Action{{Key: "mv", Value: "2,2,0"}},
	}})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil {
		t.Fatalf("decode count: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 1 || len(rest) == 0 {
		t.Fatalf("unexpected packet=%#v count=%#v rest=%d", packet, values, len(rest))
	}
}

// TestActionString verifies Nitro action formatting.
func TestActionString(t *testing.T) {
	value := actionString([]Action{{Key: "mv", Value: "2,2,0"}, {Key: "sit"}})
	if value != "/mv 2,2,0/sit/" {
		t.Fatalf("unexpected action string %q", value)
	}
}
