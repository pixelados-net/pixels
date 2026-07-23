package notification

import (
	"errors"
	"math"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesNotificationFields verifies notification field order.
func TestEncodeWritesNotificationFields(t *testing.T) {
	packet, err := Encode(20, -5, 5)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if values[0].Int32 != 20 || values[1].Int32 != -5 || values[2].Int32 != 5 {
		t.Fatalf("unexpected values %#v", values)
	}
}

// TestEncodeRejectsOutOfRangeValues verifies protocol overflow protection.
func TestEncodeRejectsOutOfRangeValues(t *testing.T) {
	_, err := Encode(math.MaxInt32+1, 1, 5)
	if !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("expected range error, got %v", err)
	}
}
