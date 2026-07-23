package roomalert

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsActionInsteadOfRoomID verifies Nitro's room-alert wire shape.
func TestDecodeReadsActionInsteadOfRoomID(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(0), codec.String("Causitas"), codec.String(""))
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Action != 0 || payload.Message != "Causitas" || payload.Topic != "" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	packet := codec.Packet{Header: Header + 1}
	if _, err := Decode(packet); err != codec.ErrUnexpectedHeader {
		t.Fatalf("err=%v", err)
	}
}
