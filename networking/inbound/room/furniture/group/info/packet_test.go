package info

import (
	"errors"

	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeReadsSupportedEmulatorShapes verifies Comet, Nitro, and Arcturus compatibility.
func TestDecodeReadsSupportedEmulatorShapes(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		groupID int64
	}{
		{name: "comet item only", payload: []byte{0, 0, 0, 7}},
		{name: "nitro undefined group compatibility", payload: []byte{0, 0, 0, 7, 0, 0}},
		{name: "nitro and arcturus group", payload: []byte{0, 0, 0, 7, 0, 0, 0, 8}, groupID: 8},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := Decode(codec.Packet{Header: Header, Payload: test.payload})
			if err != nil || value.ObjectID != 7 || value.GroupID != test.groupID {
				t.Fatalf("value=%#v err=%v", value, err)
			}
		})
	}
}

// TestDecodeRejectsMalformedShapes verifies compatibility does not accept arbitrary trailing bytes.
func TestDecodeRejectsMalformedShapes(t *testing.T) {
	tests := []codec.Packet{
		{Header: Header, Payload: []byte{0, 0, 0}},
		{Header: Header, Payload: []byte{0, 0, 0, 7, 1, 0}},
		{Header: Header, Payload: []byte{0, 0, 0, 7, 0, 0, 0, 8, 1}},
	}
	for _, packet := range tests {
		if _, err := Decode(packet); err == nil {
			t.Fatalf("expected malformed payload rejection for %v", packet.Payload)
		}
	}
}

// TestDecodeRejectsUnexpectedHeader verifies packet ownership.
func TestDecodeRejectsUnexpectedHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1, Payload: []byte{0, 0, 0, 7}})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
