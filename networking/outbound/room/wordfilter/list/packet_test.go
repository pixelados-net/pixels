package list

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesFilterWords verifies variable filter list encoding.
func TestEncodeWritesFilterWords(t *testing.T) {
	packet, _ := Encode([]string{"spam", "scam"})
	values, rest, err := codec.DecodePacket(packet, CountDefinition)
	if err != nil || values[0].Int32 != 2 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
	values, rest, _ = codec.DecodePayload(nil, WordDefinition, rest)
	if values[0].String != "spam" {
		t.Fatalf("unexpected first word %#v", values)
	}
	values, rest, _ = codec.DecodePayload(nil, WordDefinition, rest)
	if values[0].String != "scam" || len(rest) != 0 {
		t.Fatalf("unexpected second word %#v", values)
	}
}
