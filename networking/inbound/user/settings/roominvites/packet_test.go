package roominvites

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsBlockedFlag verifies USER_SETTINGS_INVITES decoding.
func TestDecodeReadsBlockedFlag(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.BooleanField}, codec.Bool(true))
	blocked, err := Decode(packet)
	if err != nil || !blocked {
		t.Fatalf("unexpected blocked=%v err=%v", blocked, err)
	}
}
