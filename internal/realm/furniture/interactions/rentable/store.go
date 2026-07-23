package rentable

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Store persists guarded rentable furniture transitions.
type Store interface {
	// FindRoomSpace finds the active room's rentable-space furniture.
	FindRoomSpace(context.Context, int64) (State, bool, error)
	// FindItem finds one rentable furniture instance.
	FindItem(context.Context, int64) (State, bool, error)
	// Rent starts or extends a rental guarded by current state.
	Rent(context.Context, int64, int64, int32, int64) (State, bool, error)
	// Cancel clears a rental owned by one renter.
	Cancel(context.Context, int64, int64) (bool, error)
	// Buyout transfers permanent ownership to the current renter.
	Buyout(context.Context, int64, int64) (bool, error)
	// WithinTransaction runs currency and furniture changes atomically.
	WithinTransaction(context.Context, func(context.Context) error) error
}

// Repository implements rentable persistence with PostgreSQL guards.
type Repository struct{ pool *postgres.Pool }

// NewRepository creates rentable persistence.
func NewRepository(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// FindRoomSpace finds the active room's rentable-space furniture.
func (repository *Repository) FindRoomSpace(ctx context.Context, roomID int64) (State, bool, error) {
	return repository.query(ctx, `select i.id,i.owner_player_id,i.rental_owner_player_id,i.rental_expires_at,coalesce(i.rental_price_credits,0) from furniture_items i join furniture_definitions d on d.id=i.definition_id where i.room_id=$1 and i.deleted_at is null and d.interaction_type='rentable_space' order by i.id limit 1`, roomID)
}

// FindItem finds one rentable furniture instance.
func (repository *Repository) FindItem(ctx context.Context, itemID int64) (State, bool, error) {
	return repository.query(ctx, `select i.id,i.owner_player_id,i.rental_owner_player_id,i.rental_expires_at,coalesce(i.rental_price_credits,0) from furniture_items i join furniture_definitions d on d.id=i.definition_id where i.id=$1 and i.deleted_at is null and d.interaction_type='rentable_space'`, itemID)
}

// Rent starts or extends a rental guarded by current state.
func (repository *Repository) Rent(ctx context.Context, itemID int64, renterID int64, price int32, durationSeconds int64) (State, bool, error) {
	return repository.query(ctx, `update furniture_items set rental_owner_player_id=$2,rental_expires_at=greatest(coalesce(rental_expires_at,now()),now())+make_interval(secs=>$4),rental_price_credits=$3,updated_at=now(),version=version+1 where id=$1 and deleted_at is null and (rental_owner_player_id is null or rental_expires_at<=now() or rental_owner_player_id=$2) returning id,owner_player_id,rental_owner_player_id,rental_expires_at,coalesce(rental_price_credits,0)`, itemID, renterID, price, durationSeconds)
}

// Cancel clears a rental owned by one renter.
func (repository *Repository) Cancel(ctx context.Context, itemID int64, renterID int64) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `update furniture_items set rental_owner_player_id=null,rental_expires_at=null,updated_at=now(),version=version+1 where id=$1 and rental_owner_player_id=$2 and deleted_at is null`, itemID, renterID)
	return err == nil && result.RowsAffected() == 1, err
}

// Buyout transfers permanent ownership to the current renter.
func (repository *Repository) Buyout(ctx context.Context, itemID int64, renterID int64) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `update furniture_items set owner_player_id=$2,rental_owner_player_id=null,rental_expires_at=null,updated_at=now(),version=version+1 where id=$1 and rental_owner_player_id=$2 and rental_expires_at>now() and deleted_at is null`, itemID, renterID)
	return err == nil && result.RowsAffected() == 1, err
}

// WithinTransaction runs currency and furniture changes atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return postgres.WithinScope(ctx, repository.pool, work)
}

// query reads one rentable state and maps a missing row.
func (repository *Repository) query(ctx context.Context, query string, arguments ...any) (State, bool, error) {
	var state State
	err := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, query, arguments...).Scan(&state.ItemID, &state.OwnerPlayerID, &state.RenterPlayerID, &state.ExpiresAt, &state.PriceCredits)
	if errors.Is(err, pgx.ErrNoRows) {
		return State{}, false, nil
	}
	return state, err == nil, err
}
