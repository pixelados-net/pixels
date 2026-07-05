package init

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies HANDSHAKE_INIT_DIFFIE packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode("value", "value")
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

	if len(values) != 2 {
		t.Fatalf("expected %d fields, got %d", 2, len(values))
	}
}

// TestDefinitionNames verifies declarative field names.
func TestDefinitionNames(t *testing.T) {
	if Definition[0].Name == "" {
		t.Fatal("expected named definition field")
	}
}

// decodeValues returns decoded packet values or an error.
func decodeValues(packet codec.Packet) ([]codec.Value, error) {
	return codec.DecodePacketExact(packet, Definition)
}
