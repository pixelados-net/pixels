package name

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies USER_IGNORE decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.StringField}, codec.String("demo"))
	value, err := Decode(packet)
	if err != nil || value != "demo" {
		t.Fatalf("decode value=%q err=%v", value, err)
	}
}
