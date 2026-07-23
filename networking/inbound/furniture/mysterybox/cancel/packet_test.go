package cancel

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the mystery-box owner identifier.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7))
	if err != nil {
		t.Fatal(err)
	}
	ownerID, err := Decode(packet)
	if err != nil || ownerID != 7 {
		t.Fatalf("owner=%d err=%v", ownerID, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header error")
	}
}
