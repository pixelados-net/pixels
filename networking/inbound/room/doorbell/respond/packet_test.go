package respond

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies ROOM_DOORBELL decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("Demo"), codec.Bool(true))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Username != "Demo" || !payload.Accepted {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeRejectsInvalidPacket verifies header and payload validation.
func TestDecodeRejectsInvalidPacket(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected unexpected header")
	}
	if _, err := Decode(codec.Packet{Header: Header}); err == nil {
		t.Fatal("expected malformed payload")
	}
}
