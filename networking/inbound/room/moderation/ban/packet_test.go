package ban

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsBanFields verifies ban decoding.
func TestDecodeReadsBanFields(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Int32(9), codec.String("RWUAM_BAN_USER_DAY"))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 || payload.RoomID != 9 || payload.Duration != "RWUAM_BAN_USER_DAY" {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}
