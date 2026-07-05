package roomcard

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppend verifies room card payload encoding.
func TestAppend(t *testing.T) {
	payload, err := Append(nil, Card{
		RoomID:       5,
		RoomName:     "Demo",
		OwnerID:      1,
		OwnerName:    "demo",
		MaxUserCount: 25,
		Tags:         []string{"chat"},
		ShowOwner:    true,
		AllowPets:    true,
	})
	if err != nil {
		t.Fatalf("append room card: %v", err)
	}

	values, rest, err := codec.DecodePayload(nil, baseDefinition, payload)
	if err != nil {
		t.Fatalf("decode base: %v", err)
	}
	if values[0].Int32 != 5 || values[12].Int32 != 1 {
		t.Fatalf("unexpected base values %#v", values)
	}

	_, rest, err = codec.DecodePayload(nil, tagDefinition, rest)
	if err != nil {
		t.Fatalf("decode tag: %v", err)
	}
	values, _, err = codec.DecodePayload(nil, bitmaskDefinition, rest)
	if err != nil {
		t.Fatalf("decode bitmask: %v", err)
	}
	if values[0].Int32 != BitShowOwner|BitAllowPets {
		t.Fatalf("unexpected bitmask %d", values[0].Int32)
	}
}

// TestBitmask verifies optional room card flags.
func TestBitmask(t *testing.T) {
	bitmask := Bitmask(Card{
		OfficialRoomPicRef: "thumb",
		Group:              &Group{},
		Ad:                 &Ad{},
		ShowOwner:          true,
		AllowPets:          true,
		DisplayAd:          true,
	})
	if bitmask != 63 {
		t.Fatalf("expected all bitmask flags, got %d", bitmask)
	}
}
