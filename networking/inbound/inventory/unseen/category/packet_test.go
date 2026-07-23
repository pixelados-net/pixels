package category

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsCategory verifies the exact Nitro shape.
func TestDecodeReadsCategory(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(1))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Category != 1 {
		t.Fatalf("payload=%#v error=%v", payload, err)
	}
}

// TestDecodeRejectsInvalidPackets verifies header and trailing payload validation.
func TestDecodeRejectsInvalidPackets(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header, Payload: []byte{0, 0, 0, 1, 0}}); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload, got %v", err)
	}
}
