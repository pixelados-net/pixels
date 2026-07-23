package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// ListServeItems returns every serving mapping.
func (repository *Repository) ListServeItems(ctx context.Context) ([]botrecord.ServeItem, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,keyword,definition_id from bot_serve_items order by keyword`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]botrecord.ServeItem, 0)
	for rows.Next() {
		item := botrecord.ServeItem{}
		if err = rows.Scan(&item.ID, &item.Keyword, &item.DefinitionID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// CreateServeItem inserts a serving mapping.
func (repository *Repository) CreateServeItem(ctx context.Context, keyword string, definitionID int64) (botrecord.ServeItem, error) {
	item := botrecord.ServeItem{}
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into bot_serve_items(keyword,definition_id) values($1,$2) returning id,keyword,definition_id`, keyword, definitionID).Scan(&item.ID, &item.Keyword, &item.DefinitionID)
	if duplicate(err) {
		return item, botrecord.ErrServeKeywordExists
	}
	return item, err
}

// UpdateServeItem changes a serving mapping.
func (repository *Repository) UpdateServeItem(ctx context.Context, id int64, keyword string, definitionID int64) (botrecord.ServeItem, bool, error) {
	item := botrecord.ServeItem{}
	err := repository.executorFor(ctx).QueryRow(ctx, `update bot_serve_items set keyword=$2,definition_id=$3 where id=$1 returning id,keyword,definition_id`, id, keyword, definitionID).Scan(&item.ID, &item.Keyword, &item.DefinitionID)
	if duplicate(err) {
		return item, false, botrecord.ErrServeKeywordExists
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return item, false, nil
	}
	return item, err == nil, err
}

// duplicate reports a PostgreSQL unique constraint violation.
func duplicate(err error) bool {
	value := &pgconn.PgError{}
	return errors.As(err, &value) && value.Code == "23505"
}

// DeleteServeItem removes a serving mapping.
func (repository *Repository) DeleteServeItem(ctx context.Context, id int64) (bool, error) {
	command, err := repository.executorFor(ctx).Exec(ctx, `delete from bot_serve_items where id=$1`, id)
	return err == nil && command.RowsAffected() > 0, err
}

// RecordVisit appends one room entry.
func (repository *Repository) RecordVisit(ctx context.Context, roomID int64, playerID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into room_visits(room_id,player_id) values($1,$2)`, roomID, playerID)
	return err
}

// VisitsSince returns recent visits excluding the requesting owner.
func (repository *Repository) VisitsSince(ctx context.Context, roomID int64, excludedPlayerID int64, since time.Time, limit int) ([]botrecord.Visit, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select v.room_id,v.player_id,p.username,v.entered_at from room_visits v join players p on p.id=v.player_id where v.room_id=$1 and v.player_id<>$2 and v.entered_at>$3 order by v.entered_at desc limit $4`, roomID, excludedPlayerID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	visits := make([]botrecord.Visit, 0)
	for rows.Next() {
		visit := botrecord.Visit{}
		if err = rows.Scan(&visit.RoomID, &visit.PlayerID, &visit.PlayerName, &visit.EnteredAt); err != nil {
			return nil, err
		}
		visits = append(visits, visit)
	}
	return visits, rows.Err()
}

// storeAssertion verifies repository completeness.
var storeAssertion botrecord.Store = (*Repository)(nil)
