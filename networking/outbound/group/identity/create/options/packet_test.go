package options

import (
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesCreatorOptions verifies list encoding.
func TestEncodeWritesCreatorOptions(t *testing.T) {
	packet, err := Encode(10, []grouprecord.EligibleRoom{{ID: 130, Name: "GROUPS QA Creator"}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
	values, remaining, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.BooleanField,
	}, packet.Payload)
	if err != nil || len(remaining) != 0 || values[0].Int32 != 10 || values[1].Int32 != 1 || values[2].Int32 != 130 || values[3].String != "GROUPS QA Creator" || values[4].Boolean {
		t.Fatalf("values=%#v remaining=%d err=%v", values, len(remaining), err)
	}
}
