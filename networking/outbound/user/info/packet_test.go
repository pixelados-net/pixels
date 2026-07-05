package info

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies USER_INFO packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(Params{
		UserID:        7,
		Username:      "demo",
		Figure:        "hd-180-1",
		Gender:        "M",
		Motto:         "Welcome",
		CanChangeName: true,
	})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}

	if packet.Header != Header {
		t.Fatalf("expected header %d, got %d", Header, packet.Header)
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}

	if values[0].Int32 != 7 || values[1].String != "demo" || values[13].Boolean {
		t.Fatalf("unexpected user info values: %#v", values)
	}
	if !values[12].Boolean {
		t.Fatal("expected can change name")
	}
}

// TestDefinitionNames verifies declarative field names.
func TestDefinitionNames(t *testing.T) {
	for index, field := range Definition {
		if field.Name == "" {
			t.Fatalf("expected field %d to be named", index)
		}
	}
}
