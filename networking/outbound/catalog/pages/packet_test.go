package pages

import (
	"strings"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeVerifiesSyntheticRoot verifies the recursive CATALOG_INDEX prefix.
func TestEncodeVerifiesSyntheticRoot(t *testing.T) {
	packet, err := Encode([]Node{{Visible: true, IconImage: 2, PageID: 4, Name: "chairs", Localization: "Chairs"}}, "NORMAL")
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, _, err := codec.DecodePayload(nil, codec.Definition{
		codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field,
	}, packet.Payload)
	if err != nil || packet.Header != Header || values[2].Int32 != -1 || values[3].String != "root" || values[6].Int32 != 1 {
		t.Fatalf("unexpected values %#v error %v", values, err)
	}
}

// TestEncodeRejectsOversizedNestedNode verifies recursive encoding errors.
func TestEncodeRejectsOversizedNestedNode(t *testing.T) {
	node := Node{Visible: true, PageID: 1, Name: "root", Children: []Node{{Name: strings.Repeat("x", 1<<16)}}}
	if _, err := Encode([]Node{node}, "NORMAL"); err == nil {
		t.Fatal("expected oversized nested node error")
	}
	if _, err := Encode(nil, strings.Repeat("x", 1<<16)); err == nil {
		t.Fatal("expected oversized mode error")
	}
}
