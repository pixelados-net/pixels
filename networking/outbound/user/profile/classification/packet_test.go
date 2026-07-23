package classification

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies classification list bounds.
func TestEncode(t *testing.T) {
	packet, err := Encode([]int32{7}, []string{"demo"}, []string{"peer"})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode classification: %#v, %v", packet, err)
	}
	if _, err = Encode([]int32{7}, nil, nil); !errors.Is(err, codec.ErrInvalidField) {
		t.Fatalf("expected invalid parallel lists, got %v", err)
	}
}
