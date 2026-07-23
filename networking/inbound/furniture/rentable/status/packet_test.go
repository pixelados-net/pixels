package status

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the header-only status request.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{})
	if err != nil {
		t.Fatal(err)
	}
	if err = Decode(packet); err != nil {
		t.Fatal(err)
	}
	packet.Header++
	if err = Decode(packet); err == nil {
		t.Fatal("expected header error")
	}
}
