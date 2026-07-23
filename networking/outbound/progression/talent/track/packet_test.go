package track

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	talentdata "github.com/niflaot/pixels/networking/outbound/progression/talent/data"
)

// TestEncodePreservesNestedTrackSections verifies levels, tasks, perks, and products.
func TestEncodePreservesNestedTrackSections(t *testing.T) {
	levels := []talentdata.Level{{ID: 2, State: 1, Tasks: []talentdata.Task{{ID: 7, Index: 3, BadgeCode: "ACH_Test3", State: 2, Progress: 20, RequiredProgress: 50}}, Perks: []string{"TRADE"}, Products: []talentdata.Product{{Name: "chair", Value: 42}}}}
	packet, err := Encode("citizenship", levels)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode track: %#v %v", packet, err)
	}
	definition := codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field}
	values, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode track: %v rest=%d", err, len(rest))
	}
	if values[0].String != "citizenship" || values[4].Int32 != 1 || values[7].String != "ACH_Test3" || values[12].String != "TRADE" || values[14].String != "chair" {
		t.Fatalf("unexpected nested values %#v", values)
	}
}
