package gotoflat

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies direct room admission wire validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(130))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 130 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
