package earnedprizes

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the empty earned-prize snapshot wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || len(packet.Payload) != 4 || values[0].Int32 != 0 {
		t.Fatalf("packet=%+v values=%+v err=%v", packet, values, err)
	}
}

// TestPrizeDefinition verifies the documented repeating prize record.
func TestPrizeDefinition(t *testing.T) {
	payload, err := codec.AppendPayload(nil, PrizeDefinition,
		codec.Int32(9), codec.String("goal"), codec.Int32(3),
		codec.String("reward"), codec.Bool(true), codec.String("Prize"))
	if err != nil {
		t.Fatal(err)
	}
	packet := codec.Packet{Header: Header, Payload: payload}
	values, err := codec.DecodePacketExact(packet, PrizeDefinition)
	if err != nil || values[0].Int32 != 9 || !values[4].Boolean {
		t.Fatalf("values=%+v err=%v", values, err)
	}
}
