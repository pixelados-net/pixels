package sign

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesHeaderAndValue verifies exact inbound decoding.
func TestDecodeValidatesHeaderAndValue(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(3))
	if err != nil {
		t.Fatal(err)
	}
	value, err := Decode(packet)
	if err != nil || value != 3 {
		t.Fatalf("value=%d err=%v", value, err)
	}
	packet.Header++
	if _, err = Decode(packet); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
