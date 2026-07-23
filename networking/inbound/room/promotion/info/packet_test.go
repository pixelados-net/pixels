package info

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies strict header-only decoding.
func TestDecode(t *testing.T) {
	if e := Decode(codec.Packet{Header: Header}); e != nil {
		t.Fatal(e)
	}
	if e := Decode(codec.Packet{Header: Header + 1}); e == nil {
		t.Fatal("expected header error")
	}
	if e := Decode(codec.Packet{Header: Header, Payload: []byte{0}}); e == nil {
		t.Fatal("expected trailing payload")
	}
}
