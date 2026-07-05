package search

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies NAVIGATOR_SEARCH decoding.
func TestDecode(t *testing.T) {
	packet := mustPacket(t, codec.String("hotel_view"), codec.String("tag:demo"))
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Code != "hotel_view" || payload.Data != "tag:demo" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDefinitionNames verifies declarative field names.
func TestDefinitionNames(t *testing.T) {
	if Definition[0].Name != "code" || Definition[1].Name != "data" {
		t.Fatalf("unexpected definition %#v", Definition)
	}
}

// mustPacket creates a test packet.
func mustPacket(t *testing.T, values ...codec.Value) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(Header, Definition, values...)
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	return packet
}
