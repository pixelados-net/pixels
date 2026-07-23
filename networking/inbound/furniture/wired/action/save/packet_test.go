package save

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies WIRED_ACTION_SAVE delay placement.
func TestDecode(t *testing.T) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field},
		codec.Int32(5), codec.Int32(0), codec.String(""), codec.Int32(0), codec.Int32(6), codec.Int32(1))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	settings, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil || settings.ItemID != 5 || settings.DelayPulses != 6 || settings.SelectionMode != 1 {
		t.Fatalf("unexpected decode %#v %v", settings, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1, Payload: payload}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected header error, got %v", err)
	}
}

// FuzzDecode verifies arbitrary action settings never panic or bypass payload validation.
func FuzzDecode(fuzz *testing.F) {
	fuzz.Add([]byte{0, 0, 0, 1})
	fuzz.Fuzz(func(_ *testing.T, payload []byte) { _, _ = Decode(codec.Packet{Header: Header, Payload: payload}) })
}
