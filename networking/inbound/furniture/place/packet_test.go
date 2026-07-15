package place

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies PLACE_FLOOR_ITEM decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("12 3 4 2"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload != (Payload{ItemID: 12, X: 3, Y: 4, Rotation: 2}) {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestDecodeWall verifies the shared placement header accepts Nitro wall notation.
func TestDecodeWall(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("12 :w=3,2 l=1,1 r "))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 12 || payload.WallPosition != ":w=3,2 l=1,1 r" {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}

// TestDecodeRejectsUnexpectedHeader verifies header validation.
func TestDecodeRejectsUnexpectedHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}

// TestDecodeRejectsMalformedValues verifies placement string validation.
func TestDecodeRejectsMalformedValues(t *testing.T) {
	cases := []string{"", "12 3 4", "12 3 4 x", "12 3 4 2 5"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			packet, err := codec.NewPacket(Header, Definition, codec.String(raw))
			if err != nil {
				t.Fatalf("new packet: %v", err)
			}

			_, err = Decode(packet)
			if !errors.Is(err, ErrMalformedPlacement) {
				t.Fatalf("expected malformed placement, got %v", err)
			}
		})
	}
}
