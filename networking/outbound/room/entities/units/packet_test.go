package units

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies UNIT packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode([]Unit{{
		UserID: 7, Name: "demo", Motto: "hi", Figure: "hd-180-1",
		RoomIndex: 3, X: 1, Y: 2, Z: "0", Direction: 4,
		Gender: "M", GroupID: -1, GroupStatus: -1, Moderator: true,
	}})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil {
		t.Fatalf("decode count: %v", err)
	}
	if values[0].Int32 != 1 || len(rest) == 0 {
		t.Fatalf("unexpected count=%#v rest=%d", values, len(rest))
	}
}

// TestEncodeRentableBot verifies Nitro's type-four owner and skill tail.
func TestEncodeRentableBot(t *testing.T) {
	packet, err := Encode([]Unit{{
		Type: RentableBotType, UserID: -7, Name: "Frank", Motto: "Tea?", Figure: "hd-180-1",
		RoomIndex: 3, X: 1, Y: 2, Z: "0", Direction: 4, Gender: "M",
		OwnerID: 1, OwnerName: "demo", Skills: []uint16{0, 2, 6},
	}})
	if err != nil {
		t.Fatalf("encode bot: %v", err)
	}
	count, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	base, rest, baseErr := codec.DecodePayload(nil, baseDefinition(), rest)
	bot, rest, botErr := codec.DecodePayload(nil, botDefinition(), rest)
	if err != nil || baseErr != nil || botErr != nil || count[0].Int32 != 1 || base[0].Int32 != -7 || base[9].Int32 != RentableBotType || bot[1].Int32 != 1 || bot[3].Int32 != 3 {
		t.Fatalf("count=%#v base=%#v bot=%#v errors=%v/%v/%v", count, base, bot, err, baseErr, botErr)
	}
	for index, expected := range []uint16{0, 2, 6} {
		values, next, skillErr := codec.DecodePayload(nil, codec.Definition{codec.Uint16Field}, rest)
		if skillErr != nil || values[0].Uint16 != expected {
			t.Fatalf("skill %d values=%#v err=%v", index, values, skillErr)
		}
		rest = next
	}
	if len(rest) != 0 {
		t.Fatalf("unexpected trailing payload %d", len(rest))
	}
}

// TestEncodePet verifies Nitro's type-two pet tail without avatar fields.
func TestEncodePet(t *testing.T) {
	packet, err := Encode([]Unit{{
		Type: PetType, UserID: 42, Name: "Pixel", Figure: "0 0 FFFFFF 0",
		RoomIndex: 9, X: 4, Y: 5, Z: "1.2", Direction: 2, PetSpecies: 0,
		OwnerID: 7, OwnerName: "demo", PetRarity: 3, HasSaddle: true,
		IsRiding: false, CanBreed: true, HasBreedingPermission: true, PetLevel: 8, Posture: "std",
	}})
	if err != nil {
		t.Fatalf("encode pet: %v", err)
	}
	count, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	base, rest, baseErr := codec.DecodePayload(nil, baseDefinition(), rest)
	pet, rest, petErr := codec.DecodePayload(nil, petDefinition(), rest)
	if err != nil || baseErr != nil || petErr != nil || count[0].Int32 != 1 || base[9].Int32 != PetType {
		t.Fatalf("count=%#v base=%#v pet=%#v errors=%v/%v/%v", count, base, pet, err, baseErr, petErr)
	}
	if pet[1].Int32 != 7 || pet[2].String != "demo" || pet[3].Int32 != 3 || !pet[4].Boolean || pet[10].Int32 != 8 || pet[11].String != "std" {
		t.Fatalf("unexpected pet tail %#v", pet)
	}
	if len(rest) != 0 {
		t.Fatalf("unexpected trailing payload %d", len(rest))
	}
}

// BenchmarkUnitPetEncode measures the protocol type-two encoding path.
func BenchmarkUnitPetEncode(b *testing.B) {
	unit := Unit{
		Type: PetType, UserID: 42, Name: "Pixel", Figure: "0 0 FFFFFF 0", RoomIndex: 9,
		X: 4, Y: 5, Z: "1.2", Direction: 2, PetSpecies: 0, OwnerID: 7, OwnerName: "demo",
		PetRarity: 3, HasSaddle: true, CanBreed: true, HasBreedingPermission: true, PetLevel: 8, Posture: "std",
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		_, _ = Encode([]Unit{unit})
	}
}
