package enter

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies ROOM_ENTER decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(42), codec.String("secret"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.FlatID != 42 || payload.Password != "secret" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeAcceptsEmptyPassword verifies Nitro empty password payloads.
func TestDecodeAcceptsEmptyPassword(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(1), codec.String(""))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.FlatID != 1 || payload.Password != "" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeRejectsInvalidHeader verifies header validation.
func TestDecodeRejectsInvalidHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
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
