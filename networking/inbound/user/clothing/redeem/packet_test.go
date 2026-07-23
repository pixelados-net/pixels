package redeem

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies clothing redemption wire validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(12))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 12 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
