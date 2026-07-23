package status

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeReturnsNeutralBuildersState verifies the intentional compatibility stub.
func TestEncodeReturnsNeutralBuildersState(t *testing.T) {
	packet, err := Encode()
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, rest, decodeErr := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || decodeErr != nil || packet.Header != Header || len(rest) != 0 {
		t.Fatalf("packet=%#v values=%#v error=%v decode=%v", packet, values, err, decodeErr)
	}
	for _, value := range values {
		if value.Int32 != 0 {
			t.Fatalf("expected neutral values, got %#v", values)
		}
	}
}
