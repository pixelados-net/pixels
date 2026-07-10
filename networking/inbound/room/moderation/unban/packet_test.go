package unban

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsUnbanFields verifies unban decoding.
func TestDecodeReadsUnbanFields(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Int32(9))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 || payload.RoomID != 9 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}
