package report

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Nitro's optional string and fixed report suffix.
func TestDecode(t *testing.T) {
	for _, extraData := range []string{"", "capture"} {
		definition := codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
		payload, _ := codec.AppendPayload(nil, definition, codec.String(extraData), codec.Int32(7), codec.Int32(8), codec.Int32(9), codec.Int32(10))
		decoded, err := Decode(codec.Packet{Header: Header, Payload: payload})
		if err != nil || decoded.RoomID != 7 || decoded.ItemID != 10 || decoded.ExtraDataID != extraData {
			t.Fatalf("decoded=%+v err=%v", decoded, err)
		}
	}
}

// TestDecodeRejectsDuplicatedStringLength guards the former synthetic shape.
func TestDecodeRejectsDuplicatedStringLength(t *testing.T) {
	payload, _ := codec.AppendPayload(nil, codec.Definition{codec.Uint16Field, codec.StringField}, codec.Uint16(1), codec.String("capture"))
	payload, _ = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(8), codec.Int32(9), codec.Int32(10))
	if _, err := Decode(codec.Packet{Header: Header, Payload: payload}); err == nil {
		t.Fatal("expected duplicated length to be rejected")
	}
}
