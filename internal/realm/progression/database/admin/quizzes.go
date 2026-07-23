package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// CreateQuiz inserts one quiz definition.
func (repository *Repository) CreateQuiz(ctx context.Context, value progressionrecord.Quiz) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into quizzes(code,kind,enabled) values($1,$2,$3)`, value.Code, value.Kind, value.Enabled)
	return err
}

// DisableQuiz soft-disables one quiz definition.
func (repository *Repository) DisableQuiz(ctx context.Context, code string) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update quizzes set enabled=false where code=$1 and enabled`, code)
	return result.RowsAffected() > 0, err
}

// CreateQuizQuestion inserts one quiz question.
func (repository *Repository) CreateQuizQuestion(ctx context.Context, value progressionrecord.QuizQuestion) (progressionrecord.QuizQuestion, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into quiz_questions(quiz_code,question_ref,correct_answer_id) values($1,$2,$3) returning id`, value.QuizCode, value.QuestionRef, value.CorrectAnswerID).Scan(&value.ID)
	return value, err
}

// UpdateQuizQuestion replaces one quiz question.
func (repository *Repository) UpdateQuizQuestion(ctx context.Context, value progressionrecord.QuizQuestion) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update quiz_questions set question_ref=$3,correct_answer_id=$4 where id=$1 and quiz_code=$2`, value.ID, value.QuizCode, value.QuestionRef, value.CorrectAnswerID)
	return result.RowsAffected() > 0, err
}

// DeleteQuizQuestion removes one quiz question.
func (repository *Repository) DeleteQuizQuestion(ctx context.Context, code string, id int64) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `delete from quiz_questions where id=$1 and quiz_code=$2`, id, code)
	return result.RowsAffected() > 0, err
}

// QuizResult reads one player's durable quiz outcome.
func (repository *Repository) QuizResult(ctx context.Context, playerID int64, code string) (progressionrecord.QuizResult, bool, error) {
	var value progressionrecord.QuizResult
	var passedAt pgtype.Timestamptz
	err := repository.executorFor(ctx).QueryRow(ctx, `select player_id,quiz_code,passed,failed_question_refs,attempted_at,passed_at from player_quiz_results where player_id=$1 and quiz_code=$2`, playerID, code).Scan(&value.PlayerID, &value.QuizCode, &value.Passed, &value.FailedQuestionRefs, &value.AttemptedAt, &passedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return progressionrecord.QuizResult{}, false, nil
	}
	if err != nil {
		return progressionrecord.QuizResult{}, false, err
	}
	if passedAt.Valid {
		value.PassedAt = &passedAt.Time
	}
	return value, true, nil
}
