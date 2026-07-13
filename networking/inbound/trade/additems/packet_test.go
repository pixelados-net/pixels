package additems

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies a bounded offered-item collection.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(2), codec.Int32(4), codec.Int32(5))
	values, err := Decode(packet)
	if err != nil || len(values) != 2 || values[0] != 4 || values[1] != 5 {
		t.Fatalf("values=%v err=%v", values, err)
	}
}
