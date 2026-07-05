package toolbar

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies CLIENT_TOOLBAR_TOGGLE payload unpacking.
func TestDecode(t *testing.T) {
	packet := codec.Packet{Header: Header}
	decoded, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}

	_ = decoded
}

// TestDecodeRejectsInvalidHeader verifies packet header validation.
func TestDecodeRejectsInvalidHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
	if err == nil {
		t.Fatal("expected decode error")
	}
}

// TestDecodeRejectsInvalidPayload verifies exact payload validation.
func TestDecodeRejectsInvalidPayload(t *testing.T) {
	packet := codec.Packet{Header: Header, Payload: []byte{1}}
	_, err := Decode(packet)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

// mustPacket returns packet or fails the test.
func mustPacket(t *testing.T, packet codec.Packet, err error) codec.Packet {
	t.Helper()
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	return packet
}
