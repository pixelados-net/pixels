package current

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesNitroSettingsOrder verifies variable tags and integer booleans.
func TestEncodeUsesNitroSettingsOrder(t *testing.T) {
	packet, err := Encode(Params{RoomID: 7, Name: "Room", MaxUsers: 25, MaxUsersLimit: 100, Tags: []string{"social"}, AllowPets: true, ChatDistance: 50})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	values, rest, err := codec.DecodePacket(packet, PrefixDefinition)
	if err != nil || values[7].Int32 != 1 {
		t.Fatalf("prefix=%#v err=%v", values, err)
	}
	_, rest, _ = codec.DecodePayload(nil, TagDefinition, rest)
	values, rest, err = codec.DecodePayload(nil, SuffixDefinition, rest)
	if err != nil || len(rest) != 0 || values[1].Int32 != 1 || values[10].Int32 != 50 {
		t.Fatalf("suffix=%#v rest=%d err=%v", values, len(rest), err)
	}
}
