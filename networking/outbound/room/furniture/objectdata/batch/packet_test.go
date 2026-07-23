package batch

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

// TestEncode verifies count-prefixed typed object updates.
func TestEncode(t *testing.T) {
	packet, err := Encode([]int64{3}, []*stuffdata.Data{stuffdata.IntArray([]int32{9})})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode batch: %v %#v", err, packet)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || len(rest) != 0 || values[0].Int32 != 1 || values[1].Int32 != 3 {
		t.Fatalf("unexpected batch %#v rest=%d %v", values, len(rest), err)
	}
	if _, err = Encode([]int64{1}, nil); !errors.Is(err, codec.ErrInvalidField) {
		t.Fatalf("expected invalid field, got %v", err)
	}
}
