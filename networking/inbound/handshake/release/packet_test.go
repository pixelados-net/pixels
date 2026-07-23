package release

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies RELEASE_VERSION payload unpacking.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("value"), codec.String("value"), codec.Int32(7), codec.Int32(7))
	packet = mustPacket(t, packet, err)
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
	packet, err := codec.NewPacket(Header, Definition, codec.String("value"), codec.String("value"), codec.Int32(7), codec.Int32(7))
	packet = mustPacket(t, packet, err)
	packet.Payload = append(packet.Payload, 1, 2, 3, 4, 5)
	_, err = Decode(packet)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

// TestDefinitionNames verifies declarative field names.
func TestDefinitionNames(t *testing.T) {
	if Definition[0].Name == "" {
		t.Fatal("expected named definition field")
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
