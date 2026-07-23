package list

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates empty badge inventory requests.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("wrong header error=%v", err)
	}
	if err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("payload error=%v", err)
	}
}
