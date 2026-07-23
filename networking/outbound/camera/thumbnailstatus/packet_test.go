package thumbnailstatus

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies both status flags.
func TestEncode(t *testing.T) {
	packet, err := Encode(true, false)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || !values[0].Boolean || values[1].Boolean {
		t.Fatalf("values=%+v err=%v decode=%v", values, err, decodeErr)
	}
}
