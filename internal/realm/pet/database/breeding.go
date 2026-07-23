package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// SaveBreedingSession creates or adds one owner confirmation to a compatible session.
func (repository *Repository) SaveBreedingSession(ctx context.Context, value petrecord.BreedingSession, actorID int64) (petrecord.BreedingSession, bool, error) {
	saved := petrecord.BreedingSession{}
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, lockErr := repository.executorFor(txCtx).Exec(txCtx, `select pg_advisory_xact_lock(hashtextextended('pet-breeding-parent:'||$1::text,0)),pg_advisory_xact_lock(hashtextextended('pet-breeding-parent:'||$2::text,0)),pg_advisory_xact_lock(hashtextextended('pet-breeding-nest:'||$3::text,0))`, value.ParentOneID, value.ParentTwoID, value.NestItemID); lockErr != nil {
			return lockErr
		}
		if _, deleteErr := repository.executorFor(txCtx).Exec(txCtx, `delete from pet_breeding_sessions where nest_item_id=$1 and (state in ('completed','cancelled') or expires_at<=now())`, value.NestItemID); deleteErr != nil {
			return deleteErr
		}
		query := `insert into pet_breeding_sessions(nest_item_id,room_id,generation_token,parent_one_id,parent_two_id,owner_one_confirmed,owner_two_confirmed,state,expires_at)
select $1,$2,$3,$4,$5,p1.owner_player_id=$6,p2.owner_player_id=$6,'requested',$7 from pets p1,pets p2 where p1.id=$4 and p2.id=$5
and not exists(select 1 from pet_breeding_sessions s where s.nest_item_id<>$1 and s.state in ('requested','confirmed') and s.expires_at>now() and ($4 in (s.parent_one_id,s.parent_two_id) or $5 in (s.parent_one_id,s.parent_two_id)))
on conflict(nest_item_id) do update set owner_one_confirmed=pet_breeding_sessions.owner_one_confirmed or (select owner_player_id=$6 from pets where id=$4),owner_two_confirmed=pet_breeding_sessions.owner_two_confirmed or (select owner_player_id=$6 from pets where id=$5),state=case when (pet_breeding_sessions.owner_one_confirmed or (select owner_player_id=$6 from pets where id=$4)) and (pet_breeding_sessions.owner_two_confirmed or (select owner_player_id=$6 from pets where id=$5)) then 'confirmed' else 'requested' end,expires_at=$7,version=pet_breeding_sessions.version+1
where pet_breeding_sessions.room_id=$2 and pet_breeding_sessions.parent_one_id=$4 and pet_breeding_sessions.parent_two_id=$5 and pet_breeding_sessions.state in ('requested','confirmed') and pet_breeding_sessions.expires_at>now()
returning nest_item_id,room_id,generation_token,parent_one_id,parent_two_id,owner_one_confirmed,owner_two_confirmed,state,expires_at,version`
		return repository.executorFor(txCtx).QueryRow(txCtx, query, value.NestItemID, value.RoomID, value.GenerationToken, value.ParentOneID, value.ParentTwoID, actorID, value.ExpiresAt).Scan(
			&saved.NestItemID, &saved.RoomID, &saved.GenerationToken, &saved.ParentOneID, &saved.ParentTwoID, &saved.OwnerOneConfirmed, &saved.OwnerTwoConfirmed, &saved.State, &saved.ExpiresAt, &saved.Version)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return petrecord.BreedingSession{}, false, nil
	}
	return saved, err == nil, err
}

// FindBreedingSession returns one non-expired active or recently completed session.
func (repository *Repository) FindBreedingSession(ctx context.Context, nestItemID int64) (petrecord.BreedingSession, bool, error) {
	value := petrecord.BreedingSession{}
	err := repository.executorFor(ctx).QueryRow(ctx, `select nest_item_id,room_id,generation_token,parent_one_id,parent_two_id,owner_one_confirmed,owner_two_confirmed,state,expires_at,version from pet_breeding_sessions where nest_item_id=$1`, nestItemID).Scan(
		&value.NestItemID, &value.RoomID, &value.GenerationToken, &value.ParentOneID, &value.ParentTwoID, &value.OwnerOneConfirmed, &value.OwnerTwoConfirmed, &value.State, &value.ExpiresAt, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return petrecord.BreedingSession{}, false, nil
	}
	return value, err == nil, err
}

// SetBreedingSessionState compare-and-swaps one durable workflow state.
func (repository *Repository) SetBreedingSessionState(ctx context.Context, nestItemID int64, from string, to string, version int64) (bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `update pet_breeding_sessions set state=$3,version=version+1 where nest_item_id=$1 and state=$2 and version=$4`, nestItemID, from, to, version)
	return err == nil && command.RowsAffected() > 0, err
}

// CancelBreedingRoom releases every active reservation in one closing room.
func (repository *Repository) CancelBreedingRoom(ctx context.Context, roomID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `update pet_breeding_sessions set state='cancelled',version=version+1 where room_id=$1 and state in ('requested','confirmed')`, roomID)
	return err
}

// CancelBreedingPet releases every active reservation involving one pet.
func (repository *Repository) CancelBreedingPet(ctx context.Context, petID int64, roomID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `update pet_breeding_sessions set state='cancelled',version=version+1 where room_id=$2 and state in ('requested','confirmed') and $1 in (parent_one_id,parent_two_id)`, petID, roomID)
	return err
}
