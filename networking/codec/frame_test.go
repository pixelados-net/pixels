package codec

import (
	"errors"
	"testing"
)

// TestAppendFrameAndDecodeFrames verifies frame round trips and concatenation.
func TestAppendFrameAndDecodeFrames(t *testing.T) {
	var encoded []byte
	var err error
	encoded, err = AppendFrame(encoded, Packet{Header: 3928})
	if err != nil {
		t.Fatalf("append first frame: %v", err)
	}

	encoded, err = AppendFrame(encoded, Packet{Header: 2596, Payload: []byte{1, 2}})
	if err != nil {
		t.Fatalf("append second frame: %v", err)
	}

	packets, rest, err := DecodeFrames(nil, encoded)
	if err != nil {
		t.Fatalf("decode frames: %v", err)
	}

	if len(rest) != 0 {
		t.Fatalf("expected no remaining bytes, got %d", len(rest))
	}

	if len(packets) != 2 {
		t.Fatalf("expected two packets, got %d", len(packets))
	}

	if packets[0].Header != 3928 || packets[1].Header != 2596 {
		t.Fatalf("unexpected headers: %d %d", packets[0].Header, packets[1].Header)
	}
}

// TestDecodeFramesReturnsPartialRest verifies incomplete frames are preserved.
func TestDecodeFramesReturnsPartialRest(t *testing.T) {
	encoded, err := AppendFrame(nil, Packet{Header: 3928})
	if err != nil {
		t.Fatalf("append frame: %v", err)
	}

	packets, rest, err := DecodeFrames(nil, encoded[:5])
	if err != nil {
		t.Fatalf("decode partial frame: %v", err)
	}

	if len(packets) != 0 {
		t.Fatalf("expected no packets, got %d", len(packets))
	}

	if len(rest) != 5 {
		t.Fatalf("expected five remaining bytes, got %d", len(rest))
	}
}

// TestDecodeFramesRejectsSmallFrame verifies invalid frame lengths fail.
func TestDecodeFramesRejectsSmallFrame(t *testing.T) {
	_, _, err := DecodeFrames(nil, []byte{0, 0, 0, 1, 0})
	if !errors.Is(err, ErrFrameTooSmall) {
		t.Fatalf("expected small frame error, got %v", err)
	}
}

// TestNewPacketAndDecodePacket verifies packet payload helpers.
func TestNewPacketAndDecodePacket(t *testing.T) {
	packet, err := NewPacket(295, Definition{Int32Field}, Int32(99))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	values, rest, err := DecodePacket(packet, Definition{Int32Field})
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}

	if len(rest) != 0 {
		t.Fatalf("expected no rest, got %d", len(rest))
	}

	if values[0].Int32 != 99 {
		t.Fatalf("expected request id 99, got %d", values[0].Int32)
	}
}

// TestDecodePacketExactRejectsRest verifies exact packet decoding rejects trailing payload.
func TestDecodePacketExactRejectsRest(t *testing.T) {
	packet := Packet{Header: 1, Payload: []byte{1}}

	_, err := DecodePacketExact(packet, Definition{})
	if !errors.Is(err, ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload error, got %v", err)
	}
}
