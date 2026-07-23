package stuffdata

import (
	"bytes"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendStringArrayEncodesNitroFormat verifies group furniture string-array wire data.
func TestAppendStringArrayEncodesNitroFormat(t *testing.T) {
	actual, err := AppendStringArray(nil, []string{"0", "4", "badge", "AABBCC", "112233"})
	if err != nil {
		t.Fatalf("append string array: %v", err)
	}
	expected, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.StringField}, codec.Int32(2), codec.Int32(5), codec.String("0"), codec.String("4"), codec.String("badge"), codec.String("AABBCC"), codec.String("112233"))
	if err != nil {
		t.Fatalf("build expected payload: %v", err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("unexpected payload: %v", actual)
	}
}

// TestLegacyEncodesNitroFormat verifies format-zero state object data.
func TestLegacyEncodesNitroFormat(t *testing.T) {
	payload, err := Legacy("11").Append(nil)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(0), codec.String("11"))
	if err != nil || !bytes.Equal(payload, expected) {
		t.Fatalf("payload=%v expected=%v err=%v", payload, expected, err)
	}
}
