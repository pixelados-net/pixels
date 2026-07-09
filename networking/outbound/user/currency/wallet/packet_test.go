package wallet

import (
	"errors"
	"math"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesWalletEntries verifies collection field order.
func TestEncodeWritesWalletEntries(t *testing.T) {
	packet, err := Encode([]Entry{{Type: 0, Amount: 10}, {Type: 5, Amount: 20}})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	if values[0].Int32 != 2 {
		t.Fatalf("unexpected count %d", values[0].Int32)
	}
	first, rest, err := codec.DecodePayload(nil, EntryDefinition, rest)
	if err != nil {
		t.Fatalf("decode first: %v", err)
	}
	second, rest, err := codec.DecodePayload(nil, EntryDefinition, rest)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode second: %v rest=%d", err, len(rest))
	}
	if first[0].Int32 != 0 || first[1].Int32 != 10 || second[0].Int32 != 5 || second[1].Int32 != 20 {
		t.Fatalf("unexpected entries first=%#v second=%#v", first, second)
	}
}

// TestEncodeRejectsOutOfRangeAmount verifies protocol overflow protection.
func TestEncodeRejectsOutOfRangeAmount(t *testing.T) {
	_, err := Encode([]Entry{{Type: 5, Amount: math.MaxInt32 + 1}})
	if !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("expected range error, got %v", err)
	}
}
