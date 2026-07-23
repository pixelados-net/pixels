package render

import (
	"bytes"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the length-prefixed raw PNG wire.
func TestDecode(t *testing.T) {
	prefix, _ := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(4))
	packet := codec.Packet{Header: Header, Payload: append(prefix, 1, 2, 3, 4)}
	payload, err := Decode(packet)
	if err != nil || !bytes.Equal(payload.PNG, []byte{1, 2, 3, 4}) {
		t.Fatalf("payload=%v err=%v", payload.PNG, err)
	}
}

// TestDecodeRejectsInvalidLengths verifies hostile length rejection.
func TestDecodeRejectsInvalidLengths(t *testing.T) {
	for _, size := range []int32{0, -1, MaxBytes + 1, 4} {
		prefix, _ := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(size))
		if _, err := Decode(codec.Packet{Header: Header, Payload: append(prefix, 1)}); err == nil {
			t.Fatalf("accepted size %d", size)
		}
	}
}
