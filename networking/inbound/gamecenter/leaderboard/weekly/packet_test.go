package weekly

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the renderer wire shape and header validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(8), codec.Int32(9), codec.Int32(10), codec.Int32(11), codec.Int32(12))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(packet)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != (Payload{GameTypeID: 7, Year: 8, Week: 9, Offset: 10, Limit: 11, Unknown: 12}) {
		t.Fatalf("unexpected payload: %+v", decoded)
	}
	packet.Header++
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected header validation")
	}
}
