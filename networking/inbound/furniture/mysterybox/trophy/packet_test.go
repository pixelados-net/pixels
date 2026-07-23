package trophy

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the mystery-trophy inscription payload.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.String("hola"))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 7 || payload.Text != "hola" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header error")
	}
}
