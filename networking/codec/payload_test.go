package codec

import (
	"errors"
	"testing"
)

// TestAppendPayloadAndDecodePayload verifies schema-ordered payload round trips.
func TestAppendPayloadAndDecodePayload(t *testing.T) {
	definition := Definition{Named("ok", BooleanField), Int32Field, Uint16Field, Uint32Field, StringField, ByteField, DoubleField}
	payload, err := AppendPayload(nil, definition, Bool(true), Int32(-7), Uint16(9), Uint32(11), String("nitro"), Byte(200), Float64(0.125))
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

	if !values[0].Boolean || values[1].Int32 != -7 || values[4].String != "nitro" || values[5].Byte != 200 || values[6].Double != 0.125 {
		t.Fatalf("unexpected values: %#v", values)
	}

	if definition[0].Name != "ok" {
		t.Fatalf("expected named field, got %q", definition[0].Name)
	}
}

// TestAppendPayloadRejectsMismatchedValues verifies definition arity is enforced.
func TestAppendPayloadRejectsMismatchedValues(t *testing.T) {
	_, err := AppendPayload(nil, Definition{Uint16Field})
	if !errors.Is(err, ErrInvalidField) {
		t.Fatalf("expected invalid field error, got %v", err)
	}
}

// TestAppendPayloadSkipsMissingOptionalValues verifies optional values may be absent.
func TestAppendPayloadSkipsMissingOptionalValues(t *testing.T) {
	payload, err := AppendPayload(nil, Definition{Optional(Uint16Field)})
	if err != nil {
		t.Fatalf("append optional payload: %v", err)
	}

	if len(payload) != 0 {
		t.Fatalf("expected empty optional payload, got %d bytes", len(payload))
	}
}

// TestDecodePayloadRejectsTruncatedString verifies short strings fail decoding.
func TestDecodePayloadRejectsTruncatedString(t *testing.T) {
	_, _, err := DecodePayload(nil, Definition{StringField}, []byte{0, 5, 'h'})
	if !errors.Is(err, ErrTruncatedPayload) {
		t.Fatalf("expected truncated payload error, got %v", err)
	}
}

// TestDecodePayloadRejectsTruncatedByte verifies a missing byte field is reported.
func TestDecodePayloadRejectsTruncatedByte(t *testing.T) {
	_, _, err := DecodePayload(nil, Definition{ByteField}, nil)
	if !errors.Is(err, ErrTruncatedPayload) {
		t.Fatalf("expected truncated payload error, got %v", err)
	}
}

// TestDecodePayloadSkipsTruncatedOptional verifies absent optional fields are tolerated.
func TestDecodePayloadSkipsTruncatedOptional(t *testing.T) {
	values, rest, err := DecodePayload(nil, Definition{Optional(Int32Field)}, nil)
	if err != nil {
		t.Fatalf("decode optional payload: %v", err)
	}

	if len(values) != 0 {
		t.Fatalf("expected no optional values, got %d", len(values))
	}

	if len(rest) != 0 {
		t.Fatalf("expected no remaining payload, got %d", len(rest))
	}
}
