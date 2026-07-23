package look

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies UNIT_LOOK decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(12), codec.Int32(7))
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.X != 12 || payload.Y != 7 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}

// TestDecodeRejectsInvalidPayload verifies exact payload validation.
func TestDecodeRejectsInvalidPayload(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header, Payload: []byte{1}})
	if err == nil {
		t.Fatal("expected payload error")
	}
}
