package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// Find returns one non-deleted pet with ordered appearance parts.
func (repository *Repository) Find(ctx context.Context, petID int64) (petrecord.Pet, bool, error) {
	query := `select ` + petColumns + petFrom + `where p.id=$1 and p.deleted_at is null` + petGroup
	pet, err := scanPet(repository.executorFor(ctx).QueryRow(ctx, query, petID))
	if errors.Is(err, pgx.ErrNoRows) {
		return petrecord.Pet{}, false, nil
	}
	return pet, err == nil, err
}

// ListAdmin returns one bounded protected pet page without an offset scan.
func (repository *Repository) ListAdmin(ctx context.Context, filter petrecord.AdminFilter) ([]petrecord.Pet, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	query := `select ` + petColumns + petFrom + `where ($1::bigint is null or p.owner_player_id=$1) and ($2='' or p.name ilike '%'||$2||'%') and ($3::integer is null or p.type_id=$3) and ($4::bigint is null or p.room_id=$4) and ($5='' or p.state=$5) and ($6 or p.deleted_at is null) and p.id>$7` + petGroup + `order by p.id limit $8`
	rows, err := repository.executorFor(ctx).Query(ctx, query, filter.OwnerPlayerID, filter.Name, filter.TypeID, filter.RoomID, filter.State, filter.IncludeDeleted, filter.Cursor, filter.Limit)
	if err != nil {
		return nil, err
	}
	return scanPets(rows)
}

// FindByOperation returns the pet produced by one completed idempotent operation.
func (repository *Repository) FindByOperation(ctx context.Context, operationKey string) (petrecord.Pet, bool, error) {
	var petID int64
	err := repository.executorFor(ctx).QueryRow(ctx, `select pet_id from pet_operations where idempotency_key=$1 and state='completed'`, operationKey).Scan(&petID)
	if errors.Is(err, pgx.ErrNoRows) {
		return petrecord.Pet{}, false, nil
	}
	if err != nil {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, petID)
}

// Inventory lists one owner's inventory pets in stable identifier order.
func (repository *Repository) Inventory(ctx context.Context, ownerID int64) ([]petrecord.Pet, error) {
	query := `select ` + petColumns + petFrom + `where p.owner_player_id=$1 and p.room_id is null and p.state='inventory' and p.deleted_at is null` + petGroup + `order by p.id`
	rows, err := repository.executorFor(ctx).Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	return scanPets(rows)
}

// Room lists one room's placed pets in stable identifier order.
func (repository *Repository) Room(ctx context.Context, roomID int64) ([]petrecord.Pet, error) {
	query := `select ` + petColumns + petFrom + `where p.room_id=$1 and p.state='room' and p.deleted_at is null` + petGroup + `order by p.id`
	rows, err := repository.executorFor(ctx).Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	return scanPets(rows)
}

// CountInventory counts inventory pets for one owner.
func (repository *Repository) CountInventory(ctx context.Context, ownerID int64) (int, error) {
	var count int
	err := repository.executorFor(ctx).QueryRow(ctx, `select count(*) from pets where owner_player_id=$1 and room_id is null and state='inventory' and deleted_at is null`, ownerID).Scan(&count)
	return count, err
}

// Grant creates or returns one idempotently granted pet.
func (repository *Repository) Grant(ctx context.Context, params petrecord.GrantParams) (petrecord.Pet, bool, error) {
	var pet petrecord.Pet
	created := false
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var existingID int64
		err := repository.executorFor(txCtx).QueryRow(txCtx, `select pet_id from pet_operations where idempotency_key=$1 and state='completed'`, params.OperationKey).Scan(&existingID)
		if err == nil {
			var found bool
			pet, found, err = repository.Find(txCtx, existingID)
			if !found && err == nil {
				return petrecord.ErrConflict
			}
			return err
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		err = repository.executorFor(txCtx).QueryRow(txCtx, `insert into pets(owner_player_id,name,type_id,breed_id,palette_id,color,rarity,level,experience,energy,happiness,state) select $1,$2,$3,$4,$5,$6,b.rarity,1,0,100,100,'inventory' from pet_breeds b where b.type_id=$3 and b.breed_id=$4 and b.palette_id=$5 and b.enabled returning id`, params.OwnerPlayerID, params.Name, params.TypeID, params.BreedID, params.PaletteID, params.Color).Scan(&existingID)
		if err != nil {
			return err
		}
		if err = repository.insertAppearanceParts(txCtx, existingID, params.Parts); err != nil {
			return err
		}
		if _, err = repository.executorFor(txCtx).Exec(txCtx, `insert into pet_operations(idempotency_key,pet_id,kind,state,result) values($1,$2,$3,'completed','{}')`, params.OperationKey, existingID, operationKind(params.OperationKey)); err != nil {
			return err
		}
		var found bool
		pet, found, err = repository.Find(txCtx, existingID)
		created = found
		return err
	})
	return pet, created, err
}

// insertAppearanceParts persists one complete ordered renderer genotype in one query.
func (repository *Repository) insertAppearanceParts(ctx context.Context, petID int64, parts []petrecord.AppearancePart) error {
	if len(parts) == 0 {
		return nil
	}
	layers, components, palettes := make([]int32, len(parts)), make([]int32, len(parts)), make([]int32, len(parts))
	for index, part := range parts {
		layers[index], components[index], palettes[index] = part.LayerID, part.PartID, part.PaletteID
	}
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into pet_appearance_parts(pet_id,ordinal,layer_id,part_id,palette_id) select $1,part.ordinal-1,part.layer_id,part.part_id,part.palette_id from unnest($2::integer[],$3::integer[],$4::integer[]) with ordinality as part(layer_id,part_id,palette_id,ordinal)`, petID, layers, components, palettes)
	return err
}

// operationKind classifies durable grants for support and metrics.
func operationKind(operationKey string) string {
	switch {
	case strings.HasPrefix(operationKey, "breeding:"):
		return "breeding"
	case strings.HasPrefix(operationKey, "package:"):
		return "package"
	default:
		return "grant"
	}
}

// Place compare-and-swaps an owned inventory pet into a room.
func (repository *Repository) Place(ctx context.Context, petID int64, ownerID int64, roomID int64, x int, y int, z float64, rotation int16, version int64) (petrecord.Pet, bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pets set room_id=$3,x=$4,y=$5,z=$6,rotation=$7,state='room',posture='std',updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id is null and state='inventory' and version=$8 and deleted_at is null`, petID, ownerID, roomID, x, y, z, rotation, version)
	if err != nil || command.RowsAffected() == 0 {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, petID)
}

// Pickup compare-and-swaps a placed pet back to inventory.
func (repository *Repository) Pickup(ctx context.Context, petID int64, roomID int64, ownerID int64, version int64) (petrecord.Pet, bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pets set room_id=null,x=null,y=null,z=null,rotation=null,state='inventory',posture='std',updated_at=now(),version=version+1 where id=$1 and room_id=$2 and owner_player_id=$3 and state='room' and version=$4 and deleted_at is null`, petID, roomID, ownerID, version)
	if err != nil || command.RowsAffected() == 0 {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, petID)
}

// SavePosition persists a placed pet position by version.
func (repository *Repository) SavePosition(ctx context.Context, petID int64, roomID int64, x int, y int, z float64, rotation int16, version int64) (int64, bool, error) {
	var next int64
	err := repository.executorFor(ctx).QueryRow(ctx, `update pets set x=$3,y=$4,z=$5,rotation=$6,updated_at=now(),version=version+1 where id=$1 and room_id=$2 and state='room' and version=$7 and deleted_at is null returning version`, petID, roomID, x, y, z, rotation, version).Scan(&next)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return next, err == nil, err
}

// SoftDelete removes one owned pet from active reads.
func (repository *Repository) SoftDelete(ctx context.Context, petID int64, ownerID int64) (bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pets set deleted_at=now(),updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id is null and deleted_at is null`, petID, ownerID)
	return err == nil && command.RowsAffected() > 0, err
}

// UpdateAdmin replaces protected mutable fields with appearance validation.
func (repository *Repository) UpdateAdmin(ctx context.Context, petID int64, patch petrecord.AdminPatch) (petrecord.Pet, bool, error) {
	query := `update pets p set name=coalesce($2,p.name),breed_id=coalesce($3,p.breed_id),palette_id=coalesce($4,p.palette_id),color=coalesce($5,p.color),public_ride=coalesce($6,p.public_ride),public_breed=coalesce($7,p.public_breed),updated_at=now(),version=p.version+1 where p.id=$1 and p.version=$8 and p.deleted_at is null and exists(select 1 from pet_breeds b where b.type_id=p.type_id and b.breed_id=coalesce($3,p.breed_id) and b.palette_id=coalesce($4,p.palette_id) and b.enabled) returning p.id`
	var updatedID int64
	err := repository.executorFor(ctx).QueryRow(ctx, query, petID, patch.Name, patch.BreedID, patch.PaletteID, patch.Color, patch.PublicRide, patch.PublicBreed, patch.Version).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return petrecord.Pet{}, false, nil
	}
	if err != nil {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, updatedID)
}

// TransferOwner moves one inventory pet to another owner optimistically.
func (repository *Repository) TransferOwner(ctx context.Context, petID int64, ownerID int64, version int64) (petrecord.Pet, bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pets set owner_player_id=$2,updated_at=now(),version=version+1 where id=$1 and version=$3 and room_id is null and state='inventory' and deleted_at is null`, petID, ownerID, version)
	if err != nil || command.RowsAffected() == 0 {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, petID)
}

// DeleteAdmin soft-deletes one pet in any current location optimistically.
func (repository *Repository) DeleteAdmin(ctx context.Context, petID int64, version int64) (bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pets set deleted_at=now(),updated_at=now(),version=version+1 where id=$1 and version=$2 and deleted_at is null`, petID, version)
	return err == nil && command.RowsAffected() > 0, err
}

// AppendAudit records one protected pet mutation.
func (repository *Repository) AppendAudit(ctx context.Context, petID int64, actorID int64, action string, detail string) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into pet_audit_entries(pet_id,actor_player_id,action,detail) values($1,$2,$3,jsonb_build_object('reason',$4))`, petID, actorID, action, detail)
	return err
}

// AppendGlobalAudit records one protected reference mutation.
func (repository *Repository) AppendGlobalAudit(ctx context.Context, actorID int64, action string, detail string) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into pet_audit_entries(pet_id,actor_player_id,action,detail) values(null,$1,$2,jsonb_build_object('reason',$3))`, actorID, action, detail)
	return err
}
