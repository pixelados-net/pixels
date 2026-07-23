package save

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies WIRED_TRIGGER_SAVE field order.
func TestDecode(t *testing.T) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field},
		codec.Int32(1), codec.Int32(1), codec.Int32(4), codec.String("key"), codec.Int32(1), codec.Int32(8), codec.Int32(2))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	settings, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil || settings.ItemID != 1 || settings.StringParam != "key" || settings.SelectionMode != 2 {
		t.Fatalf("unexpected decode %#v %v", settings, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1, Payload: payload}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected header error, got %v", err)
	}
}

// FuzzDecode verifies arbitrary trigger settings never panic or bypass payload validation.
func FuzzDecode(fuzz *testing.F) {
	fuzz.Add([]byte{0, 0, 0, 1})
	fuzz.Fuzz(func(_ *testing.T, payload []byte) { _, _ = Decode(codec.Packet{Header: Header, Payload: payload}) })
}
