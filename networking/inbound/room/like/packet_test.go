package like

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies ROOM_LIKE decoding and validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(1))
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Rating != 1 {
		t.Fatalf("decode packet: payload=%+v err=%v", payload, err)
	}
	if _, err := Decode(codec.Packet{Header: Header + 1}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("unexpected header error = %v", err)
	}
}

// BenchmarkDecode measures ROOM_LIKE decoding.
func BenchmarkDecode(b *testing.B) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(1))
	b.ReportAllocs()
	for range b.N {
		_, _ = Decode(packet)
	}
}
