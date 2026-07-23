package id

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies USER_IGNORE_ID decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(42))
	value, err := Decode(packet)
	if err != nil || value != 42 {
		t.Fatalf("decode value=%d err=%v", value, err)
	}
}
