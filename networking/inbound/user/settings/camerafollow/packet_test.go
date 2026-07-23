package camerafollow

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the privacy toggle wire.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Bool(true))
	payload, err := Decode(packet)
	if err != nil || !payload.CameraFollowBlocked {
		t.Fatalf("payload=%+v err=%v", payload, err)
	}
}
