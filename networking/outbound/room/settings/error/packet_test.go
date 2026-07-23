package error

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesSettingsError verifies protocol error fields.
func TestEncodeWritesSettingsError(t *testing.T) {
	packet, _ := Encode(8, CodeInvalidPassword, "")
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 8 || values[1].Int32 != CodeInvalidPassword {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
