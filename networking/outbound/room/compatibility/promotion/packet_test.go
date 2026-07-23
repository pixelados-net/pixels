package promotion

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies the neutral golden-only packet.
func TestEncode(t *testing.T) {
	p, e := Encode()
	if e != nil || p.Header != Header {
		t.Fatalf("packet=%#v err=%v", p, e)
	}
	if _, e = codec.DecodePacketExact(p, Definition); e != nil {
		t.Fatal(e)
	}
}
