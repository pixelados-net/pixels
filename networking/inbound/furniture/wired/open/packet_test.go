package open

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies WIRED_OPEN decoding and header validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(42))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 42 {
		t.Fatalf("unexpected decode %#v %v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected header error, got %v", err)
	}
}
