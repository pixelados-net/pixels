package search

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsTerm verifies HABBO_SEARCH decoding.
func TestDecodeReadsTerm(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.StringField}, codec.String("ali"))
	term, err := Decode(packet)
	if err != nil || term != "ali" {
		t.Fatalf("unexpected term=%q err=%v", term, err)
	}
}
