package relinquish

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsRoomID verifies relinquish decoding.
func TestDecodeReadsRoomID(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(9))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 9 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}
