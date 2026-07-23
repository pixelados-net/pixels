package request

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_CATALOG_PAGE decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(4), codec.Int32(-1), codec.String("NORMAL"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PageID != 4 || payload.OfferID != -1 || payload.Mode != "NORMAL" {
		t.Fatalf("unexpected payload %#v error %v", payload, err)
	}
}

// TestDecodeRejectsTrailingPayload verifies exact decoding.
func TestDecodeRejectsTrailingPayload(t *testing.T) {
	packet := codec.Packet{Header: Header, Payload: []byte{0, 0, 0, 1}}
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected truncated payload error")
	}
}
