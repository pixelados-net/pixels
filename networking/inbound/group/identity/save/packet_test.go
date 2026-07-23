package save

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeReadsPayload verifies exact protocol decoding.
func TestDecodeReadsPayload(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}, codec.Int32(7), codec.String("value"), codec.String("value"))
	payload, err := Decode(packet)
	if err != nil || payload.GroupID == 0 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
