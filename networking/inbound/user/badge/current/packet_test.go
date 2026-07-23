package current

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates current badge target decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7))
	if err != nil {
		t.Fatal(err)
	}
	playerID, err := Decode(packet)
	if err != nil || playerID != 7 {
		t.Fatalf("player=%d err=%v", playerID, err)
	}
	packet.Header++
	if _, err = Decode(packet); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("wrong header error=%v", err)
	}
}
