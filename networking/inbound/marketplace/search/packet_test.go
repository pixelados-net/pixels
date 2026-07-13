package search

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies Marketplace search filter decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(1), codec.Int32(99), codec.String("chair"), codec.Int32(2))
	value, err := Decode(packet)
	if err != nil || value.MinimumPrice != 1 || value.MaximumPrice != 99 || value.Query != "chair" || value.SortType != 2 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
