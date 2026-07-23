package database

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
)

// pollChoice stores one JSON-backed poll option.
type pollChoice struct {
	// Value stores the submitted value.
	Value string `json:"value"`
	// Text stores the visible label.
	Text string `json:"text"`
	// Type stores Nitro's choice type.
	Type int32 `json:"type"`
}

// Polls returns every enabled poll with ordered questions.
func (repository *Repository) Polls(ctx context.Context) ([]progressionpoll.Definition, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id from polls where enabled order by id`)
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
	definitions := make([]progressionpoll.Definition, 0, len(ids))
	for _, id := range ids {
		definition, found, loadErr := repository.Poll(ctx, id)
		if loadErr != nil {
			return nil, loadErr
		}
		if found {
			definitions = append(definitions, definition)
		}
	}
	return definitions, nil
}

// Poll returns one enabled poll with ordered questions.
func (repository *Repository) Poll(ctx context.Context, pollID int32) (progressionpoll.Definition, bool, error) {
	return repository.poll(ctx, `poll.id=$1`, pollID)
}

// PollForRoom returns one enabled poll assigned to a room.
func (repository *Repository) PollForRoom(ctx context.Context, roomID int64) (progressionpoll.Definition, bool, error) {
	return repository.poll(ctx, `poll.room_id=$1`, roomID)
}

// poll loads one definition and its ordered questions.
func (repository *Repository) poll(ctx context.Context, predicate string, value any) (progressionpoll.Definition, bool, error) {
	return repository.pollWhere(ctx, predicate, value, true)
}

// pollWhere loads one definition with optional enabled filtering.
func (repository *Repository) pollWhere(ctx context.Context, predicate string, value any, enabledOnly bool) (progressionpoll.Definition, bool, error) {
	var definition progressionpoll.Definition
	err := repository.executorFor(ctx).QueryRow(ctx, `select poll.id,poll.title,poll.headline,poll.summary,poll.start_message,poll.thanks_message,coalesce(poll.room_id,0),coalesce(poll.reward_badge,''),poll.version,poll.enabled from polls poll where `+predicate+` and (not $2 or poll.enabled)`, value, enabledOnly).Scan(&definition.ID, &definition.Title, &definition.Headline, &definition.Summary, &definition.StartMessage, &definition.ThanksMessage, &definition.RoomID, &definition.RewardBadge, &definition.Version, &definition.Enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return progressionpoll.Definition{}, false, nil
	}
	if err != nil {
		return progressionpoll.Definition{}, false, err
	}
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,sort_order,kind,text_ref,category,answer_type,options from poll_questions where poll_id=$1 order by sort_order,id`, definition.ID)
	if err != nil {
		return progressionpoll.Definition{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var question outcontents.Question
		var raw []byte
		if err = rows.Scan(&question.ID, &question.SortOrder, &question.Type, &question.Text, &question.Category, &question.AnswerType, &raw); err != nil {
			return progressionpoll.Definition{}, false, err
		}
		var choices []pollChoice
		if err = json.Unmarshal(raw, &choices); err != nil {
			return progressionpoll.Definition{}, false, err
		}
		for _, choice := range choices {
			question.Choices = append(question.Choices, outcontents.Choice{Value: choice.Value, Text: choice.Text, Type: choice.Type})
		}
		definition.Questions = append(definition.Questions, question)
	}
	return definition, true, rows.Err()
}

// Completed reports whether all questions were answered or the poll was rejected.
func (repository *Repository) Completed(ctx context.Context, playerID int64, pollID int32) (bool, error) {
	var completed bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from poll_answers where poll_id=$1 and player_id=$2 and rejected) or (select count(*) from poll_answers where poll_id=$1 and player_id=$2 and not rejected) >= (select count(*) from poll_questions where poll_id=$1)`, pollID, playerID).Scan(&completed)
	return completed, err
}

// SaveAnswer inserts one answer and reports first-time completion.
func (repository *Repository) SaveAnswer(ctx context.Context, playerID int64, pollID int32, questionID int32, values []string) (bool, string, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return false, "", err
	}
	result, err := repository.executorFor(ctx).Exec(ctx, `insert into poll_answers(poll_id,question_id,player_id,answer) select $1,$2,$3,$4 from poll_questions where id=$2 and poll_id=$1 on conflict do nothing`, pollID, questionID, playerID, raw)
	if err != nil {
		return false, "", err
	}
	if result.RowsAffected() == 0 {
		return false, "", progressionpoll.ErrInvalidAnswer
	}
	completed, err := repository.Completed(ctx, playerID, pollID)
	if err != nil || !completed {
		return false, "", err
	}
	var badge string
	err = repository.executorFor(ctx).QueryRow(ctx, `select coalesce(reward_badge,'') from polls where id=$1 and enabled`, pollID).Scan(&badge)
	return err == nil, badge, err
}

// RejectPoll records one durable rejection using the first question as its foreign-key anchor.
func (repository *Repository) RejectPoll(ctx context.Context, playerID int64, pollID int32) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into poll_answers(poll_id,question_id,player_id,answer,rejected) select $1,min(id),$2,'[]',true from poll_questions where poll_id=$1 on conflict do nothing`, pollID, playerID)
	return err
}

var pollStoreAssertion progressionpoll.Store = (*Repository)(nil)
