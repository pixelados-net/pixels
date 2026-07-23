// Package lovelock coordinates two-player friend furniture confirmation.
package lovelock

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	netconn "github.com/niflaot/pixels/networking/connection"
	outconfirmed "github.com/niflaot/pixels/networking/outbound/furniture/lovelock/confirmed"
	outfinished "github.com/niflaot/pixels/networking/outbound/furniture/lovelock/finished"
	outstart "github.com/niflaot/pixels/networking/outbound/furniture/lovelock/start"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Pending stores one durable lovelock handshake.
type Pending struct {
	// ItemID identifies the lovelock.
	ItemID int64
	// FirstPlayerID identifies the player who started the handshake.
	FirstPlayerID int64
	// SecondPlayerID identifies the player awaiting confirmation.
	SecondPlayerID *int64
	// SealedAt stores permanent completion.
	SealedAt *time.Time
}

// Store persists concurrency-safe lovelock state.
type Store interface {
	// Start inserts the first participant when unsealed and idle.
	Start(context.Context, int64, int64) (Pending, bool, error)
	// Invite sets a distinct second participant.
	Invite(context.Context, int64, int64) (Pending, bool, error)
	// Finish seals a handshake confirmed by its second participant.
	Finish(context.Context, int64, int64) (bool, error)
	// Cancel removes an unsealed handshake for its second participant.
	Cancel(context.Context, int64, int64) (bool, error)
}

// Repository implements lovelock persistence.
type Repository struct{ executor postgres.Executor }

// NewRepository creates lovelock persistence.
func NewRepository(executor postgres.Executor) *Repository { return &Repository{executor: executor} }

// Start inserts the first participant when unsealed and idle.
func (repository *Repository) Start(ctx context.Context, itemID int64, playerID int64) (Pending, bool, error) {
	return repository.query(ctx, `insert into furniture_lovelocks(item_id,player_one_id) values($1,$2) on conflict(item_id) do nothing returning item_id,player_one_id,player_two_id,sealed_at`, itemID, playerID)
}

// Invite sets a distinct second participant.
func (repository *Repository) Invite(ctx context.Context, itemID int64, playerID int64) (Pending, bool, error) {
	return repository.query(ctx, `update furniture_lovelocks set player_two_id=$2,updated_at=now() where item_id=$1 and sealed_at is null and player_two_id is null and player_one_id<>$2 returning item_id,player_one_id,player_two_id,sealed_at`, itemID, playerID)
}

// Finish seals a handshake confirmed by its second participant.
func (repository *Repository) Finish(ctx context.Context, itemID int64, playerID int64) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.executor).Exec(ctx, `update furniture_lovelocks set sealed_at=now(),updated_at=now() where item_id=$1 and player_two_id=$2 and sealed_at is null`, itemID, playerID)
	return err == nil && result.RowsAffected() == 1, err
}

// Cancel removes an unsealed handshake for its second participant.
func (repository *Repository) Cancel(ctx context.Context, itemID int64, playerID int64) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.executor).Exec(ctx, `delete from furniture_lovelocks where item_id=$1 and player_two_id=$2 and sealed_at is null`, itemID, playerID)
	return err == nil && result.RowsAffected() == 1, err
}

// query maps one row-returning guarded mutation.
func (repository *Repository) query(ctx context.Context, query string, arguments ...any) (Pending, bool, error) {
	var value Pending
	err := postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, query, arguments...).Scan(&value.ItemID, &value.FirstPlayerID, &value.SecondPlayerID, &value.SealedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Pending{}, false, nil
	}
	return value, err == nil, err
}

// Service coordinates room interaction and explicit confirmation.
type Service struct {
	store       Store
	connections *netconn.Registry
}

// New creates lovelock behavior.
func New(store Store, connections *netconn.Registry) *Service {
	return &Service{store: store, connections: connections}
}

// UseFurniture starts or joins an unsealed lovelock handshake.
func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	if request.Item.Definition.InteractionType != "lovelock" && request.Item.Definition.InteractionType != "hween_lovelock" {
		return false, nil
	}
	_, started, err := service.store.Start(ctx, request.Item.ID, request.PlayerID)
	if err != nil {
		return true, err
	}
	if !started {
		_, started, err = service.store.Invite(ctx, request.Item.ID, request.PlayerID)
		if err != nil {
			return true, err
		}
	}
	if !started {
		return true, nil
	}
	packet, err := outstart.Encode(int32(request.Item.ID), true)
	if err != nil {
		return true, err
	}
	return true, request.Target.Send(ctx, packet)
}

// Confirm seals or cancels the handshake for its invited second player.
func (service *Service) Confirm(ctx context.Context, request essential.Request, confirmed bool) error {
	var changed bool
	var err error
	if confirmed {
		changed, err = service.store.Finish(ctx, request.Item.ID, request.PlayerID)
	} else {
		changed, err = service.store.Cancel(ctx, request.Item.ID, request.PlayerID)
	}
	if err != nil || !changed {
		return err
	}
	if !confirmed {
		packet, encodeErr := outstart.Encode(int32(request.Item.ID), false)
		if encodeErr != nil {
			return encodeErr
		}
		return broadcast.RoomPacket(ctx, service.connections, request.Room, packet, 0)
	}
	partial, err := outconfirmed.Encode(int32(request.Item.ID))
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, service.connections, request.Room, partial, 0); err != nil {
		return err
	}
	finished, err := outfinished.Encode(int32(request.Item.ID))
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, request.Room, finished, 0)
}
