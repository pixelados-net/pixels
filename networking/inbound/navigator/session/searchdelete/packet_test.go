package searchdelete

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies NAVIGATOR_DELETE_SAVED_SEARCH decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(3))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.SearchID != 3 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
