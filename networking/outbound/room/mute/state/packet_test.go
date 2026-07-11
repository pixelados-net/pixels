package state

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesMuteState verifies room mute-all encoding.
func TestEncodeWritesMuteState(t *testing.T) {
	packet, _ := Encode(true)
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || !values[0].Boolean {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
