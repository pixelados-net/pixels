package progress

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the expired empty progress snapshot wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || len(packet.Payload) != 35 {
		t.Fatalf("packet=%+v err=%v", packet, err)
	}
	if !values[0].Boolean || values[7].String != "" || values[9].Int32 != 0 {
		t.Fatalf("unexpected neutral values: %+v", values)
	}
}

// TestRewardLimitDefinition verifies the documented repeating reward field.
func TestRewardLimitDefinition(t *testing.T) {
	payload, err := codec.AppendPayload(nil, RewardLimitDefinition, codec.Int32(25))
	if err != nil || len(payload) != 4 {
		t.Fatalf("payload=%v err=%v", payload, err)
	}
}
