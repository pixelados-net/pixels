package request

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsRoomID verifies ROOM_SETTINGS decoding.
func TestDecodeReadsRoomID(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(42))
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 42 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
