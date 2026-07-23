package status

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsRequestedProduct verifies Nitro's captured subscription request.
func TestDecodeReadsRequestedProduct(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("habbo_club"))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ProductName != "habbo_club" {
		t.Fatalf("payload=%#v error=%v", payload, err)
	}
}

// TestDecodeRejectsUnexpectedStatusPayload verifies strict header and payload validation.
func TestDecodeRejectsUnexpectedStatusPayload(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header}); err == nil {
		t.Fatal("expected missing product failure")
	}
}
