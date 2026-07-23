package database

import (
	"encoding/json"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

const petColumns = `p.id,p.owner_player_id,o.username,p.name,p.type_id,p.breed_id,p.palette_id,p.color,p.rarity,p.level,p.experience,p.energy,p.happiness,p.respect,p.stats_at,p.room_id,p.x,p.y,p.z,p.rotation,p.posture,p.has_saddle,p.can_breed,p.public_ride,p.public_breed,p.grow_at,p.die_at,p.state,p.created_at,p.updated_at,p.deleted_at,p.version,coalesce(jsonb_agg(jsonb_build_object('layer_id',a.layer_id,'part_id',a.part_id,'palette_id',a.palette_id) order by a.ordinal) filter (where a.pet_id is not null),'[]'::jsonb)`

const petFrom = ` from pets p join players o on o.id=p.owner_player_id left join pet_appearance_parts a on a.pet_id=p.id `

const petGroup = ` group by p.id,o.username `

// rowScanner scans one PostgreSQL row.
type rowScanner interface {
	// Scan copies columns into destinations.
	Scan(...any) error
}

// rowsScanner scans a PostgreSQL result set.
type rowsScanner interface {
	// Next advances the result set.
	Next() bool
	// Scan copies current columns into destinations.
	Scan(...any) error
	// Err reports iteration failure.
	Err() error
	// Close releases result resources.
	Close()
}

// partJSON stores one database appearance part.
type partJSON struct {
	// LayerID identifies the renderer layer.
	LayerID int32 `json:"layer_id"`
	// PartID identifies the custom part.
	PartID int32 `json:"part_id"`
	// PaletteID identifies the custom palette.
	PaletteID int32 `json:"palette_id"`
}

// scanPet maps one aggregate row.
func scanPet(row rowScanner) (petrecord.Pet, error) {
	pet := petrecord.Pet{}
	var encoded []byte
	err := row.Scan(&pet.ID, &pet.OwnerPlayerID, &pet.OwnerName, &pet.Name, &pet.TypeID, &pet.BreedID, &pet.PaletteID,
		&pet.Color, &pet.Rarity, &pet.Level, &pet.Experience, &pet.Energy, &pet.Happiness, &pet.Respect, &pet.StatsAt,
		&pet.RoomID, &pet.X, &pet.Y, &pet.Z, &pet.Rotation, &pet.Posture, &pet.HasSaddle, &pet.CanBreed, &pet.PublicRide, &pet.PublicBreed,
		&pet.GrowAt, &pet.DieAt, &pet.State, &pet.CreatedAt, &pet.UpdatedAt, &pet.DeletedAt, &pet.Version, &encoded)
	if err != nil {
		return petrecord.Pet{}, err
	}
	var parts []partJSON
	if err = json.Unmarshal(encoded, &parts); err != nil {
		return petrecord.Pet{}, err
	}
	pet.Parts = make([]petrecord.AppearancePart, len(parts))
	for index, part := range parts {
		pet.Parts[index] = petrecord.AppearancePart{LayerID: part.LayerID, PartID: part.PartID, PaletteID: part.PaletteID}
	}
	return pet, nil
}

// scanPets maps every aggregate row.
func scanPets(rows rowsScanner) ([]petrecord.Pet, error) {
	defer rows.Close()
	items := make([]petrecord.Pet, 0)
	for rows.Next() {
		item, err := scanPet(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
