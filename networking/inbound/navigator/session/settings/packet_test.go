package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Navigator settings wire order and exactness.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(1), codec.Int32(2), codec.Int32(425), codec.Int32(592), codec.Bool(true), codec.Int32(1))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.WindowWidth != 425 || !payload.LeftPanelHidden || payload.ResultsMode != 1 {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
