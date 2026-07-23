package levelup

import (
	"errors"
	"testing"

	talentdata "github.com/niflaot/pixels/networking/outbound/progression/talent/data"
)

// TestEncodeRejectsMixedRewards preserves Nitro's mutually exclusive parser branches.
func TestEncodeRejectsMixedRewards(t *testing.T) {
	_, err := Encode("citizenship", 2, []int32{1}, []talentdata.Product{{Name: "chair", Value: 42}})
	if !errors.Is(err, ErrMixedRewards) {
		t.Fatalf("expected mixed rewards error, got %v", err)
	}
	if packet, err := Encode("citizenship", 2, []int32{1}, nil); err != nil || packet.Header != Header {
		t.Fatalf("encode perk reward: %#v %v", packet, err)
	}
}
