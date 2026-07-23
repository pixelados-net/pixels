package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// Respect applies one daily idempotent respect and experience grant.
func (repository *Repository) Respect(ctx context.Context, petID int64, actorID int64, experience int32, dailyLimit int) (petrecord.Pet, bool, error) {
	pet := petrecord.Pet{}
	applied := false
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, err := repository.executorFor(txCtx).Exec(txCtx, `select pg_advisory_xact_lock(hashtextextended('pet-respect:'||$1::text||':'||current_date::text,0))`, actorID); err != nil {
			return err
		}
		if dailyLimit > 0 {
			var current int
			if err := repository.executorFor(txCtx).QueryRow(txCtx, `select count(*) from pet_respects where actor_player_id=$1 and respected_on=current_date`, actorID).Scan(&current); err != nil {
				return err
			}
			if current >= dailyLimit {
				return nil
			}
		}
		command, err := repository.executorFor(txCtx).Exec(txCtx, `insert into pet_respects(pet_id,actor_player_id,respected_on) values($1,$2,current_date) on conflict do nothing`, petID, actorID)
		if err != nil || command.RowsAffected() == 0 {
			return err
		}
		var found bool
		pet, found, err = repository.updateStats(txCtx, petID, 0, 0, experience, 0, false)
		if err == nil && found {
			_, err = repository.executorFor(txCtx).Exec(txCtx, `update pets set respect=respect+1 where id=$1`, petID)
			if err == nil {
				pet.Respect++
			}
		}
		applied = found
		return err
	})
	return pet, applied, err
}

// UpdateLifecycle replaces absolute monsterplant deadlines.
func (repository *Repository) UpdateLifecycle(ctx context.Context, petID int64, ownerID int64, growAt *time.Time, dieAt *time.Time, version int64) (petrecord.Pet, bool, error) {
	return repository.updateAndFind(ctx, petID, `update pets set grow_at=$3,die_at=$4,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and version=$5 and state='room' and deleted_at is null`, ownerID, growAt, dieAt, version)
}

// ConsumePlant atomically soft-deletes one eligible monsterplant state.
func (repository *Repository) ConsumePlant(ctx context.Context, petID int64, ownerID int64, roomID int64, state string, version int64) (bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pets set state=$4,deleted_at=now(),updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id=$3 and type_id=16 and state='room' and version=$5 and deleted_at is null`, petID, ownerID, roomID, state, version)
	return err == nil && command.RowsAffected() > 0, err
}

// UpdateStats applies bounded stat deltas and recomputes level.
func (repository *Repository) UpdateStats(ctx context.Context, petID int64, energyDelta int32, happinessDelta int32, experienceDelta int32, version int64) (petrecord.Pet, bool, error) {
	return repository.updateStats(ctx, petID, energyDelta, happinessDelta, experienceDelta, version, true)
}

// updateStats applies one stat mutation with optional version checking.
func (repository *Repository) updateStats(ctx context.Context, petID int64, energyDelta int32, happinessDelta int32, experienceDelta int32, version int64, checkVersion bool) (petrecord.Pet, bool, error) {
	query := `update pets set energy=greatest(0,least((level*100),energy-(greatest(0,floor(extract(epoch from now()-stats_at)/$7))::integer*$8)+$2)),happiness=greatest(0,least(100,happiness-(greatest(0,floor(extract(epoch from now()-stats_at)/$7))::integer*$9)+$3)),experience=greatest(0,experience+$4),level=least(s.max_level,1+(select count(*) from unnest(array[100,200,400,600,900,1300,1800,2400,3200,4300,5700,7600,10100,13300,17500,23000,30200,39600,51900]) t(v) where v <= greatest(0,pets.experience+$4))),stats_at=now(),updated_at=now(),version=pets.version+1 from pet_species s where pets.id=$1 and s.type_id=pets.type_id and pets.deleted_at is null and (not $6 or pets.version=$5) returning pets.id`
	var updatedID int64
	err := repository.executorFor(ctx).QueryRow(ctx, query, petID, energyDelta, happinessDelta, experienceDelta, version, checkVersion, repository.config.StatDecayInterval.Seconds(), repository.config.EnergyDecay, repository.config.HappinessDecay).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return petrecord.Pet{}, false, nil
	}
	if err != nil {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, updatedID)
}

// UpdateFlags replaces public riding and breeding flags.
func (repository *Repository) UpdateFlags(ctx context.Context, petID int64, ownerID int64, publicRide bool, publicBreed bool, version int64) (petrecord.Pet, bool, error) {
	return repository.updateAndFind(ctx, petID, `update pets set public_ride=$3,public_breed=$4,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and version=$5 and deleted_at is null`, ownerID, publicRide, publicBreed, version)
}

// SetSaddle replaces equipment state.
func (repository *Repository) SetSaddle(ctx context.Context, petID int64, ownerID int64, equipped bool, version int64) (petrecord.Pet, bool, error) {
	return repository.updateAndFind(ctx, petID, `update pets set has_saddle=$3,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and version=$4 and deleted_at is null`, ownerID, equipped, version)
}

// SetBreedingEligibility replaces one pet's consumable breeding eligibility.
func (repository *Repository) SetBreedingEligibility(ctx context.Context, petID int64, eligible bool, version int64) (petrecord.Pet, bool, error) {
	return repository.updateAndFind(ctx, petID, `update pets set can_breed=$2,updated_at=now(),version=version+1 where id=$1 and version=$3 and state='room' and deleted_at is null`, eligible, version)
}

// updateAndFind executes a constrained update and reloads its aggregate.
func (repository *Repository) updateAndFind(ctx context.Context, petID int64, query string, args ...any) (petrecord.Pet, bool, error) {
	values := append([]any{petID}, args...)
	command, err := repository.executorFor(ctx).Exec(ctx, query, values...)
	if err != nil || command.RowsAffected() == 0 {
		return petrecord.Pet{}, false, err
	}
	return repository.Find(ctx, petID)
}
