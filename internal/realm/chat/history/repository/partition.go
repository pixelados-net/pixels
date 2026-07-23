package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// EnsurePartitions creates daily partitions through one date.
func (repository *Repository) EnsurePartitions(ctx context.Context, from time.Time, through time.Time) error {
	day := utcDay(from)
	last := utcDay(through)
	for !day.After(last) {
		next := day.AddDate(0, 0, 1)
		name := "chat_messages_" + day.Format("2006_01_02")
		statement := fmt.Sprintf(
			"create table if not exists %s partition of chat_messages for values from ('%s') to ('%s')",
			pgx.Identifier{name}.Sanitize(), day.Format("2006-01-02T15:04:05Z07:00"), next.Format("2006-01-02T15:04:05Z07:00"),
		)
		if _, err := repository.pool.Exec(ctx, statement); err != nil {
			return fmt.Errorf("ensure chat history partition %s: %w", name, err)
		}
		day = next
	}

	return nil
}

// DropBefore drops daily partitions older than a cutoff.
func (repository *Repository) DropBefore(ctx context.Context, cutoff time.Time) error {
	rows, err := repository.pool.Query(ctx, `select child.relname from pg_inherits join pg_class parent on pg_inherits.inhparent=parent.oid join pg_class child on pg_inherits.inhrelid=child.oid where parent.relname='chat_messages'`)
	if err != nil {
		return fmt.Errorf("list chat history partitions: %w", err)
	}
	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		names = append(names, name)
	}
	rows.Close()
	for _, name := range names {
		day, parseErr := time.Parse("2006_01_02", strings.TrimPrefix(name, "chat_messages_"))
		if parseErr != nil || !day.Before(utcDay(cutoff)) {
			continue
		}
		if _, err = repository.pool.Exec(ctx, "drop table "+pgx.Identifier{name}.Sanitize()); err != nil {
			return fmt.Errorf("drop chat history partition %s: %w", name, err)
		}
	}

	return nil
}

// utcDay returns one UTC date boundary.
func utcDay(value time.Time) time.Time {
	value = value.UTC()

	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
