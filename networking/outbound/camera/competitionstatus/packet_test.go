package competitionstatus

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the unavailable compatibility result.
func TestEncode(t *testing.T) {
	packet, err := Encode(false, "")
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].Boolean || values[1].String != "" {
		t.Fatalf("values=%+v err=%v decode=%v", values, err, decodeErr)
	}
}
