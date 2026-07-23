package info

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the complete variable-length PET_INFO wire.
func TestEncode(t *testing.T) {
	packet, err := Encode(Info{ID: 7, Name: "Pixel", Level: 2, MaximumLevel: 20, OwnerID: 1, OwnerName: "demo", SkillThresholds: []int32{1, 3}, PubliclyBreedable: true})
	values, rest, decodeErr := codec.DecodePayload(nil, baseDefinition(), packet.Payload)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 7 || values[17].Int32 != 2 {
		t.Fatalf("values=%#v errors=%v/%v", values, err, decodeErr)
	}
	for range 2 {
		_, rest, decodeErr = codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, rest)
		if decodeErr != nil {
			t.Fatalf("decode threshold: %v", decodeErr)
		}
	}
	_, rest, decodeErr = codec.DecodePayload(nil, tailDefinition(), rest)
	if decodeErr != nil || len(rest) != 0 {
		t.Fatalf("tail rest=%d err=%v", len(rest), decodeErr)
	}
}
