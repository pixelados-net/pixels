package globalid

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies renderer-confirmed global identifier decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("room:130"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.GlobalID != "room:130" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}

// TestDecodeRejectsOversized verifies bounded link identifiers.
func TestDecodeRejectsOversized(t *testing.T) {
	value := make([]byte, 129)
	packet, _ := codec.NewPacket(Header, Definition, codec.String(string(value)))
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected oversized identifier rejection")
	}
}
