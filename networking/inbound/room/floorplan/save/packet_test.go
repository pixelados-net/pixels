package save

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsNitroFloorplanFields verifies the confirmed seven-field wire order.
func TestDecodeReadsNitroFloorplanFields(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition,
		codec.String("00\r00"), codec.Int32(1), codec.Int32(0), codec.Int32(2),
		codec.Int32(-1), codec.Int32(1), codec.Int32(4),
	)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Heightmap != "00\r00" || payload.DoorX != 1 || payload.DoorDirection != 2 || payload.WallHeight != 4 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeRejectsUnexpectedHeader verifies inbound header validation.
func TestDecodeRejectsUnexpectedHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("expected header error, got %v", err)
	}
}

// TestDecodeRejectsTrailingPayload verifies strict packet consumption.
func TestDecodeRejectsTrailingPayload(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition,
		codec.String("0"), codec.Int32(0), codec.Int32(0), codec.Int32(2),
		codec.Int32(0), codec.Int32(0), codec.Int32(-1),
	)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	packet.Payload = append(packet.Payload, 1)
	if _, err = Decode(packet); err != codec.ErrUnexpectedPayload {
		t.Fatalf("expected trailing payload error, got %v", err)
	}
}
