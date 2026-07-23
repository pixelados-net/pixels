package save

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsCompleteSettings verifies variable tags and fixed settings order.
func TestDecodeReadsCompleteSettings(t *testing.T) {
	payload, _ := codec.AppendPayload(nil, PrefixDefinition, codec.Int32(7), codec.String("Room"), codec.String("Description"), codec.Int32(2), codec.String("1234"), codec.Int32(25), codec.Int32(1), codec.Int32(1))
	payload, _ = codec.AppendPayload(payload, TagDefinition, codec.String("social"))
	payload, _ = codec.AppendPayload(payload, SuffixDefinition, codec.Int32(2), codec.Bool(true), codec.Bool(false), codec.Bool(true), codec.Bool(false), codec.Int32(1), codec.Int32(0), codec.Int32(1), codec.Int32(2), codec.Int32(0), codec.Int32(1), codec.Int32(2), codec.Int32(1), codec.Int32(50), codec.Int32(2))
	decoded, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.RoomID != 7 || len(decoded.Tags) != 1 || decoded.Tags[0] != "social" || !decoded.AllowPets || decoded.ChatDistance != 50 {
		t.Fatalf("unexpected payload %#v", decoded)
	}
}

// TestDecodeRejectsUnboundedTagCount verifies allocation guard behavior.
func TestDecodeRejectsUnboundedTagCount(t *testing.T) {
	payload, _ := codec.AppendPayload(nil, PrefixDefinition, codec.Int32(7), codec.String("Room"), codec.String(""), codec.Int32(0), codec.String(""), codec.Int32(25), codec.Int32(1), codec.Int32(MaxTags+1))
	if _, err := Decode(codec.Packet{Header: Header, Payload: payload}); err == nil {
		t.Fatal("expected invalid tag count")
	}
}
