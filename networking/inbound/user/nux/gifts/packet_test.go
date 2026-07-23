package gifts

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies bounded retired NUX triples.
func TestDecode(t *testing.T) {
	payload, err := codec.AppendPayload(nil, CountDefinition, codec.Int32(3))
	if err != nil {
		t.Fatalf("encode count: %v", err)
	}
	for _, value := range []int32{1, 2, 3} {
		payload, err = codec.AppendPayload(payload, ValueDefinition, codec.Int32(value))
		if err != nil {
			t.Fatalf("encode value: %v", err)
		}
	}
	decoded, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil || len(decoded.Values) != 3 || decoded.Values[2] != 3 {
		t.Fatalf("decode gifts: %#v, %v", decoded, err)
	}
	invalid, _ := codec.AppendPayload(nil, CountDefinition, codec.Int32(2))
	if _, err = Decode(codec.Packet{Header: Header, Payload: invalid}); !errors.Is(err, codec.ErrInvalidField) {
		t.Fatalf("expected invalid count, got %v", err)
	}
}
