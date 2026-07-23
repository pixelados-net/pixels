package list

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRightsList verifies variable-length rights encoding.
func TestEncodeWritesRightsList(t *testing.T) {
	packet, err := Encode(9, []Right{{PlayerID: 7, Username: "Alice"}})
	if err != nil {
		t.Fatal(err)
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil || values[0].Int32 != 9 || values[1].Int32 != 1 {
		t.Fatalf("unexpected metadata %#v err=%v", values, err)
	}
	values, rest, err = codec.DecodePayload(values[:0], RightDefinition, rest)
	if err != nil || len(rest) != 0 || values[0].Int32 != 7 || values[1].String != "Alice" {
		t.Fatalf("unexpected right %#v err=%v", values, err)
	}
}
