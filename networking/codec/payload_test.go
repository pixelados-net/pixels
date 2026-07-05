package codec

import (
	"errors"
	"testing"
)

// TestAppendPayloadAndDecodePayload verifies schema-ordered payload round trips.
func TestAppendPayloadAndDecodePayload(t *testing.T) {
	definition := Definition{BooleanField, Int32Field, Uint16Field, Uint32Field, StringField}
	payload, err := AppendPayload(nil, definition, Bool(true), Int32(-7), Uint16(9), Uint32(11), String("nitro"))
	if err != nil {
		t.Fatalf("append payload: %v", err)
	}

	values, rest, err := DecodePayload(nil, definition, payload)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if len(rest) != 0 {
		t.Fatalf("expected no remaining payload, got %d", len(rest))
	}

	if !values[0].Boolean || values[1].Int32 != -7 || values[4].String != "nitro" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

// TestAppendPayloadRejectsMismatchedValues verifies definition arity is enforced.
func TestAppendPayloadRejectsMismatchedValues(t *testing.T) {
	_, err := AppendPayload(nil, Definition{Uint16Field})
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("expected invalid field error, got %v", err)
	}
}

// TestDecodePayloadRejectsTruncatedString verifies short strings fail decoding.
func TestDecodePayloadRejectsTruncatedString(t *testing.T) {
	_, _, err := DecodePayload(nil, Definition{StringField}, []byte{0, 5, 'h'})
	if !errors.Is(err, ErrTruncatedPayload) {
		t.Fatalf("expected truncated payload error, got %v", err)
	}
}
