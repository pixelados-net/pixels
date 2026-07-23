package configuration

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode writes Nitro's negative room bot id and requested skill.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 5, "name")
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].Int32 != -7 || values[1].Int32 != 5 || values[2].String != "name" {
		t.Fatalf("values=%#v encode=%v decode=%v", values, err, decodeErr)
	}
}
