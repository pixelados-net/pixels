package bubblealert

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies BUBBLE_ALERT encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode("furni_placement_error", "cant_stack", WithDisplayBubble())
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	values, err := codec.DecodePacketExact(packet, definition(2))
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if values[0].String != "furni_placement_error" || values[1].Int32 != 2 ||
		values[2].String != "message" || values[3].String != "cant_stack" ||
		values[4].String != "display" || values[5].String != "BUBBLE" {
		t.Fatalf("unexpected values %#v", values)
	}
}

// TestDefinitionMatchesSingleMessage verifies the exported base definition.
func TestDefinitionMatchesSingleMessage(t *testing.T) {
	packet, err := Encode("furni_placement_error", "cant_stack")
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}

	if values[1].Int32 != 1 {
		t.Fatalf("expected one pair, got %#v", values)
	}
}
