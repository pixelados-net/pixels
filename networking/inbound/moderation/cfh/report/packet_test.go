package report

import (
	"encoding/base64"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeRejectsWrongHeader verifies strict headers.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{}); err != codec.ErrUnexpectedHeader {
		t.Fatal(err)
	}
}

// TestDecodeReadsCapturedNitroPacket verifies the exact packet that previously disconnected a reporter.
func TestDecodeReadsCapturedNitroPacket(t *testing.T) {
	payload, err := base64.StdEncoding.DecodeString("ABdIZWxwIGhlIGlzIHZpb2xhdGluZyBtZQAAAAEAAAADAAAAAAABAAAAAwAEV09PTw==")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil || result.RoomID != 0 || len(result.Entries) != 1 || result.Entries[0].PlayerID != 3 || result.Entries[0].Message != "WOOO" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

// TestDecodeReadsNitroChatEvidence verifies the real int32 and string pair shape.
func TestDecodeReadsNitroChatEvidence(t *testing.T) {
	definition := codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Uint16Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField}
	payload, err := codec.AppendPayload(nil, definition, codec.String("Help"), codec.Int32(1), codec.Int32(3), codec.Int32(7), codec.Uint16(2), codec.Int32(3), codec.String("first"), codec.Int32(3), codec.String("second"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil || result.RoomID != 7 || len(result.Entries) != 2 || result.Entries[0].PlayerID != 3 || result.Entries[1].Message != "second" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

// TestDecodeRejectsInvalidEvidenceCount verifies the bounded evidence allocation.
func TestDecodeRejectsInvalidEvidenceCount(t *testing.T) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Uint16Field}, codec.String("Help"), codec.Int32(1), codec.Int32(3), codec.Int32(7), codec.Uint16(101))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(codec.Packet{Header: Header, Payload: payload}); err != codec.ErrInvalidField {
		t.Fatalf("err=%v", err)
	}
}
