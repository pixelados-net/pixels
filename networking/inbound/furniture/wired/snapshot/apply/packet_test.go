package apply

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Nitro's single-ID snapshot shape.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 7 {
		t.Fatalf("unexpected decode %#v %v", payload, err)
	}
	packet.Payload = append(packet.Payload, 0)
	if _, err = Decode(packet); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected extra payload rejection, got %v", err)
	}
}
