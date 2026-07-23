package storageurl

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the final URL wire.
func TestEncode(t *testing.T) {
	packet, err := Encode("https://storage/photos/1.png")
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].String != "https://storage/photos/1.png" {
		t.Fatalf("values=%+v err=%v decode=%v", values, err, decodeErr)
	}
}
