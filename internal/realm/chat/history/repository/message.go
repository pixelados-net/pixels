package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
)

// InsertBatch writes entries through PostgreSQL COPY.
func (repository *Repository) InsertBatch(ctx context.Context, entries []historymodel.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	rows := make([][]any, len(entries))
	for index := range entries {
		entry := entries[index]
		rows[index] = []any{entry.RoomID, entry.PlayerID, entry.TargetPlayerID, entry.Kind, entry.Message, entry.Censored, entry.CreatedAt}
	}
	_, err := repository.pool.CopyFrom(ctx, pgx.Identifier{"chat_messages"}, []string{"room_id", "player_id", "target_player_id", "kind", "message", "censored", "created_at"}, pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("copy chat history batch: %w", err)
	}

	return nil
}

// History returns one keyset history page.
func (repository *Repository) History(ctx context.Context, query historymodel.Query) ([]historymodel.Entry, error) {
	query = query.Normalize()
	clauses := make([]string, 0, 3)
	arguments := make([]any, 0, 4)
	if query.RoomID != nil {
		arguments = append(arguments, *query.RoomID)
		clauses = append(clauses, fmt.Sprintf("room_id=$%d", len(arguments)))
	}
	if query.PlayerID != nil {
		arguments = append(arguments, *query.PlayerID)
		clauses = append(clauses, fmt.Sprintf("player_id=$%d", len(arguments)))
	}
	if query.Before != nil {
		arguments = append(arguments, *query.Before)
		clauses = append(clauses, fmt.Sprintf("id<$%d", len(arguments)))
	}
	statement := `select id,room_id,player_id,target_player_id,kind,message,censored,created_at from chat_messages`
	if len(clauses) > 0 {
		statement += " where " + strings.Join(clauses, " and ")
	}
	arguments = append(arguments, query.Limit)
	statement += fmt.Sprintf(" order by id desc limit $%d", len(arguments))
	rows, err := repository.pool.Query(ctx, statement, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query chat history: %w", err)
	}
	defer rows.Close()
	items := make([]historymodel.Entry, 0, query.Limit)
	for rows.Next() {
		var entry historymodel.Entry
		if err = rows.Scan(&entry.ID, &entry.RoomID, &entry.PlayerID, &entry.TargetPlayerID, &entry.Kind, &entry.Message, &entry.Censored, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat history: %w", err)
		}
		items = append(items, entry)
	}

	return items, rows.Err()
}
