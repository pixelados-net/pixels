package thumbnail

import (
	"bytes"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies bounded thumbnail decoding.
func TestDecode(t *testing.T) {
	prefix, _ := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(3))
	payload, err := Decode(codec.Packet{Header: Header, Payload: append(prefix, 7, 8, 9)})
	if err != nil || !bytes.Equal(payload.PNG, []byte{7, 8, 9}) {
		t.Fatalf("payload=%v err=%v", payload.PNG, err)
	}
}

// TestDecodeRejectsInvalidLength verifies exact payload enforcement.
func TestDecodeRejectsInvalidLength(t *testing.T) {
	prefix, _ := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(MaxBytes+1))
	if _, err := Decode(codec.Packet{Header: Header, Payload: prefix}); err == nil {
		t.Fatal("expected invalid length")
	}
}
