package database

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
)

// AllPolls returns every poll with nested questions.
func (repository *Repository) AllPolls(ctx context.Context) ([]progressionpoll.Definition, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id from polls order by id`)
	if err != nil {
		return nil, err
	}
	var ids []int32
	for rows.Next() {
		var id int32
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	result := make([]progressionpoll.Definition, 0, len(ids))
	for _, id := range ids {
		definition, found, loadErr := repository.pollWhere(ctx, `poll.id=$1`, id, false)
		if loadErr != nil {
			return nil, loadErr
		}
		if found {
			result = append(result, definition)
		}
	}
	return result, nil
}

// CreatePoll inserts one poll and its nested questions atomically.
func (repository *Repository) CreatePoll(ctx context.Context, definition progressionpoll.Definition) (progressionpoll.Definition, error) {
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		executor := repository.executorFor(txCtx)
		var roomID any
		if definition.RoomID > 0 {
			roomID = definition.RoomID
		}
		var badge any
		if definition.RewardBadge != "" {
			badge = definition.RewardBadge
		}
		if err := executor.QueryRow(txCtx, `insert into polls(title,headline,summary,start_message,thanks_message,room_id,reward_badge,enabled) values($1,$2,$3,$4,$5,$6,$7,$8) returning id,version`, definition.Title, definition.Headline, definition.Summary, definition.StartMessage, definition.ThanksMessage, roomID, badge, definition.Enabled).Scan(&definition.ID, &definition.Version); err != nil {
			return err
		}
		return repository.replaceQuestions(txCtx, definition)
	})
	return definition, err
}

// UpdatePoll replaces one poll and its questions with optimistic concurrency.
func (repository *Repository) UpdatePoll(ctx context.Context, definition progressionpoll.Definition) (progressionpoll.Definition, bool, error) {
	updated := false
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		executor := repository.executorFor(txCtx)
		var roomID any
		if definition.RoomID > 0 {
			roomID = definition.RoomID
		}
		var badge any
		if definition.RewardBadge != "" {
			badge = definition.RewardBadge
		}
		err := executor.QueryRow(txCtx, `update polls set title=$3,headline=$4,summary=$5,start_message=$6,thanks_message=$7,room_id=$8,reward_badge=$9,enabled=$10,version=version+1,updated_at=now() where id=$1 and version=$2 returning version`, definition.ID, definition.Version, definition.Title, definition.Headline, definition.Summary, definition.StartMessage, definition.ThanksMessage, roomID, badge, definition.Enabled).Scan(&definition.Version)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		updated = true
		if _, err = executor.Exec(txCtx, `delete from poll_questions where poll_id=$1`, definition.ID); err != nil {
			return err
		}
		return repository.replaceQuestions(txCtx, definition)
	})
	return definition, updated, err
}

// replaceQuestions inserts ordered poll questions.
func (repository *Repository) replaceQuestions(ctx context.Context, definition progressionpoll.Definition) error {
	executor := repository.executorFor(ctx)
	for _, question := range definition.Questions {
		raw, err := json.Marshal(question.Choices)
		if err != nil {
			return err
		}
		if _, err = executor.Exec(ctx, `insert into poll_questions(poll_id,sort_order,kind,text_ref,category,answer_type,options) values($1,$2,$3,$4,$5,$6,$7)`, definition.ID, question.SortOrder, question.Type, question.Text, question.Category, question.AnswerType, raw); err != nil {
			return err
		}
	}
	return nil
}

// DisablePoll idempotently disables one poll.
func (repository *Repository) DisablePoll(ctx context.Context, pollID int32) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update polls set enabled=false,room_id=null,version=version+1,updated_at=now() where id=$1 and enabled`, pollID)
	return err == nil && result.RowsAffected() > 0, err
}

// AssignPoll assigns or clears one poll room with optimistic concurrency.
func (repository *Repository) AssignPoll(ctx context.Context, pollID int32, roomID int64, version int64) (int64, bool, error) {
	var value any
	if roomID > 0 {
		value = roomID
	}
	err := repository.executorFor(ctx).QueryRow(ctx, `update polls set room_id=$3,version=version+1,updated_at=now() where id=$1 and version=$2 returning version`, pollID, version, value).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return version, err == nil, err
}
