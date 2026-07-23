package codec

import "testing"

// BenchmarkAppendFrame measures header-only frame encoding.
func BenchmarkAppendFrame(b *testing.B) {
	packet := Packet{Header: 3928}
	dst := make([]byte, 0, FrameOverhead)

	b.ReportAllocs()
	for b.Loop() {
		dst = dst[:0]
		var err error
		dst, err = AppendFrame(dst, packet)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeFrames measures decoding concatenated header-only frames.
func BenchmarkDecodeFrames(b *testing.B) {
	encoded, err := AppendFrame(nil, Packet{Header: 3928})
	if err != nil {
		b.Fatal(err)
	}

	encoded, err = AppendFrame(encoded, Packet{Header: 2596})
	if err != nil {
		b.Fatal(err)
	}

	packets := make([]Packet, 0, 2)

	b.ReportAllocs()
	for b.Loop() {
		packets = packets[:0]
		var rest []byte
		packets, rest, err = DecodeFrames(packets, encoded)
		if err != nil {
			b.Fatal(err)
		}
		if len(rest) != 0 {
			b.Fatalf("unexpected rest: %d", len(rest))
		}
	}
}

// BenchmarkAppendPayload measures schema-ordered payload encoding.
func BenchmarkAppendPayload(b *testing.B) {
	definition := Definition{BooleanField, Int32Field, Uint16Field, Uint32Field, StringField}
	dst := make([]byte, 0, 32)

	b.ReportAllocs()
	for b.Loop() {
		dst = dst[:0]
		var err error
		dst, err = AppendPayload(dst, definition, Bool(true), Int32(-7), Uint16(9), Uint32(11), String("nitro"))
		if err != nil {
			b.Fatal(err)
		}
	}
}
