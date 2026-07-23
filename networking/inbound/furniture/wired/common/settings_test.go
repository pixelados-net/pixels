package common

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeSettings verifies both trigger and action save shapes.
func TestDecodeSettings(t *testing.T) {
	payload, err := encodeSettings(true, 9)
	if err != nil {
		t.Fatalf("encode settings: %v", err)
	}
	settings, err := DecodeSettings(payload, true)
	if err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if settings.ItemID != 9 || settings.DelayPulses != 3 || settings.SelectionMode != 2 || len(settings.IntParams) != 2 || len(settings.ItemIDs) != 2 {
		t.Fatalf("unexpected settings %#v", settings)
	}
}

// TestDecodeSettingsRejectsCounts verifies allocation bounds.
func TestDecodeSettingsRejectsCounts(t *testing.T) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(MaxIntParams+1))
	if err != nil {
		t.Fatalf("encode count: %v", err)
	}
	if _, err = DecodeSettings(payload, false); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload, got %v", err)
	}
}

// encodeSettings creates one test save payload.
func encodeSettings(withDelay bool, itemID int32) ([]byte, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field},
		codec.Int32(itemID), codec.Int32(2), codec.Int32(7), codec.Int32(8), codec.String("hello"), codec.Int32(2), codec.Int32(11), codec.Int32(12))
	if err != nil {
		return nil, err
	}
	if withDelay {
		return codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(3), codec.Int32(2))
	}
	return codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(2))
}
