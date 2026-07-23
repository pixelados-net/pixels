package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// ActiveQuest loads one player's single active quest.
func (repository *Repository) ActiveQuest(ctx context.Context, playerID int64) (progressionrecord.PlayerQuestState, bool, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, `select s.player_id,coalesce(s.active_quest_id,0),s.accepted_at,coalesce(p.progress,0),p.completed_at from player_quest_state s left join player_quest_progress p on p.player_id=s.player_id and p.quest_id=s.active_quest_id where s.player_id=$1 and s.active_quest_id is not null`, playerID)
	state, err := scanQuestState(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return progressionrecord.PlayerQuestState{}, false, nil
		}
		return progressionrecord.PlayerQuestState{}, false, err
	}
	return state, true, nil
}

// DailyQuestRejected reports whether one UTC-day offer was discarded.
func (repository *Repository) DailyQuestRejected(ctx context.Context, playerID int64, day time.Time) (bool, error) {
	var rejected bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from player_daily_quest_rejections where player_id=$1 and offered_on=$2)`, playerID, day.UTC().Format("2006-01-02")).Scan(&rejected)
	return rejected, err
}

// RejectDailyQuest durably discards one UTC-day offer idempotently.
func (repository *Repository) RejectDailyQuest(ctx context.Context, playerID int64, day time.Time) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into player_daily_quest_rejections(player_id,offered_on) values($1,$2) on conflict do nothing`, playerID, day.UTC().Format("2006-01-02"))
	return err
}

// QuestProgress lists one player's durable quest history.
func (repository *Repository) QuestProgress(ctx context.Context, playerID int64) ([]progressionrecord.PlayerQuestState, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select p.player_id,p.quest_id,s.accepted_at,p.progress,p.completed_at from player_quest_progress p left join player_quest_state s on s.player_id=p.player_id and s.active_quest_id=p.quest_id where p.player_id=$1 order by p.quest_id`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]progressionrecord.PlayerQuestState, 0)
	for rows.Next() {
		value, scanErr := scanQuestState(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// ActivateQuest replaces one player's active quest and returns the previous identifier.
func (repository *Repository) ActivateQuest(ctx context.Context, playerID int64, questID int64) (int64, error) {
	executor := repository.executorFor(ctx)
	if _, err := executor.Exec(ctx, `insert into player_quest_state(player_id) values($1) on conflict do nothing`, playerID); err != nil {
		return 0, err
	}
	var previous int64
	if err := executor.QueryRow(ctx, `select coalesce(active_quest_id,0) from player_quest_state where player_id=$1 for update`, playerID).Scan(&previous); err != nil {
		return 0, err
	}
	result, err := executor.Exec(ctx, `insert into player_quest_progress(player_id,quest_id) values($1,$2) on conflict do nothing`, playerID, questID)
	if err != nil {
		return 0, err
	}
	if result.RowsAffected() == 0 {
		var completed bool
		if err = executor.QueryRow(ctx, `select completed_at is not null from player_quest_progress where player_id=$1 and quest_id=$2 for update`, playerID, questID).Scan(&completed); err != nil {
			return 0, err
		}
		if completed {
			return 0, progressionrecord.ErrConflict
		}
	}
	_, err = executor.Exec(ctx, `update player_quest_state set active_quest_id=$2,accepted_at=now() where player_id=$1`, playerID, questID)
	return previous, err
}

// IncrementQuest locks and increments active quest progress up to its goal.
func (repository *Repository) IncrementQuest(ctx context.Context, playerID int64, questID int64, amount int64, goal int64) (progressionrecord.PlayerQuestState, error) {
	executor := repository.executorFor(ctx)
	var active int64
	if err := executor.QueryRow(ctx, `select coalesce(active_quest_id,0) from player_quest_state where player_id=$1 for update`, playerID).Scan(&active); err != nil {
		return progressionrecord.PlayerQuestState{}, err
	}
	if active != questID {
		return progressionrecord.PlayerQuestState{}, progressionrecord.ErrConflict
	}
	row := executor.QueryRow(ctx, `update player_quest_progress set progress=least($4,progress+$3),updated_at=now() where player_id=$1 and quest_id=$2 and completed_at is null returning player_id,quest_id,(select accepted_at from player_quest_state where player_id=$1),progress,completed_at`, playerID, questID, amount, goal)
	return scanQuestState(row)
}

// CompleteQuest marks one active quest complete exactly once.
func (repository *Repository) CompleteQuest(ctx context.Context, playerID int64, questID int64) (bool, error) {
	executor := repository.executorFor(ctx)
	result, err := executor.Exec(ctx, `update player_quest_progress set completed_at=now(),updated_at=now() where player_id=$1 and quest_id=$2 and completed_at is null`, playerID, questID)
	if err != nil || result.RowsAffected() == 0 {
		return false, err
	}
	_, err = executor.Exec(ctx, `update player_quest_state set active_quest_id=null,accepted_at=null where player_id=$1 and active_quest_id=$2`, playerID, questID)
	return err == nil, err
}

// CancelQuest clears one active quest and returns its identifier.
func (repository *Repository) CancelQuest(ctx context.Context, playerID int64) (int64, error) {
	var questID int64
	err := repository.executorFor(ctx).QueryRow(ctx, `with current as (select active_quest_id from player_quest_state where player_id=$1 and active_quest_id is not null for update), cleared as (update player_quest_state set active_quest_id=null,accepted_at=null where player_id=$1 and active_quest_id is not null returning 1) select active_quest_id from current`, playerID).Scan(&questID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return questID, nil
}

// scanQuestState scans one quest progress projection.
func scanQuestState(row interface{ Scan(...any) error }) (progressionrecord.PlayerQuestState, error) {
	var value progressionrecord.PlayerQuestState
	var accepted, completed pgtype.Timestamptz
	err := row.Scan(&value.PlayerID, &value.ActiveQuestID, &accepted, &value.Progress, &completed)
	if accepted.Valid {
		value.AcceptedAt = accepted.Time
	}
	if completed.Valid {
		value.CompletedAt = &completed.Time
	}
	return value, err
}
