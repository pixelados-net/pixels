package request

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_CATALOG_INDEX decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("NORMAL"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Mode != "NORMAL" {
		t.Fatalf("unexpected payload %#v error %v", payload, err)
	}
}

// TestDecodeRejectsHeader verifies packet identity validation.
func TestDecodeRejectsHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
