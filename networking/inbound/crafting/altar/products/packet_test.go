package products

import (
	"bytes"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeDocumentsCrossedIntegerHeader verifies header 1173 carries an altar id.
func TestDecodeDocumentsCrossedIntegerHeader(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(42))
	if !bytes.Equal(packet.Payload, []byte{0, 0, 0, 42}) {
		t.Fatalf("unexpected golden payload %v", packet.Payload)
	}
	payload, err := Decode(packet)
	if err != nil || payload.AltarItemID != 42 {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}
