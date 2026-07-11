package request

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsRoomID verifies filter request decoding.
func TestDecodeReadsRoomID(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(9))
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 9 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
