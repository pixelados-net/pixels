package mute

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsMuteFields verifies mute decoding.
func TestDecodeReadsMuteFields(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Int32(9), codec.Int32(10))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 || payload.RoomID != 9 || payload.Minutes != 10 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}
