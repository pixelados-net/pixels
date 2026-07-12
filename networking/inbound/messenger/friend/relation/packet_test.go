package relation

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsRelationship verifies SET_RELATIONSHIP_STATUS decoding.
func TestDecodeReadsRelationship(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(2))
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 || payload.Relation != 2 {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
}
