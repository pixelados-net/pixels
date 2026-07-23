package alert

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies GENERIC_ALERT packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode("Hotel message")
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	if packet.Header != Header {
		t.Fatalf("expected header %d, got %d", Header, packet.Header)
	}

	values, err := decodeValues(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}

	if len(values) != 1 || values[0].String != "Hotel message" {
		t.Fatalf("unexpected values %#v", values)
	}
}

// decodeValues returns decoded packet values or an error.
func decodeValues(packet codec.Packet) ([]codec.Value, error) {
	return codec.DecodePacketExact(packet, Definition)
}
