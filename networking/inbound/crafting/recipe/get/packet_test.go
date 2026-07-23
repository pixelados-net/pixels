package get

import (
	"bytes"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeDocumentsCrossedStringHeader verifies header 633 carries a recipe name.
func TestDecodeDocumentsCrossedStringHeader(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.StringField}, codec.String("demo"))
	if !bytes.Equal(packet.Payload, []byte{0, 4, 'd', 'e', 'm', 'o'}) {
		t.Fatalf("unexpected golden payload %v", packet.Payload)
	}
	payload, err := Decode(packet)
	if err != nil || payload.RecipeName != "demo" {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}
