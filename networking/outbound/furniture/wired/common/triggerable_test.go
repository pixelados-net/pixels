package common

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendTriggerable verifies the confirmed shared field order.
func TestAppendTriggerable(t *testing.T) {
	payload, err := AppendTriggerable(nil, true, 5, []int64{11}, 99, 7, "x", []int32{2}, 3)
	if err != nil {
		t.Fatalf("append triggerable: %v", err)
	}
	definition := codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, rest, err := codec.DecodePayload(nil, definition, payload)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode triggerable: %v rest=%d", err, len(rest))
	}
	if !values[0].Boolean || values[3].Int32 != 11 || values[6].String != "x" || values[9].Int32 != 3 {
		t.Fatalf("unexpected values %#v", values)
	}
}
