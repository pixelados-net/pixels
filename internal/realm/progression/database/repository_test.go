package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// TestProgressionAggregatesAgainstPostgres verifies every durable progression aggregate in one rolled-back transaction.
func TestProgressionAggregatesAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("PIXELS_PROGRESSION_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PIXELS_PROGRESSION_TEST_DATABASE_URL is not configured")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)
	rollback := errors.New("rollback progression integration")
	err = repository.WithinTransaction(context.Background(), func(ctx context.Context) error {
		var playerID int64
		if scanErr := repository.executorFor(ctx).QueryRow(ctx, `insert into players(username) values($1) returning id`, "progression-test-"+uuid.NewString()).Scan(&playerID); scanErr != nil {
			return scanErr
		}
		if testErr := verifyCatalog(ctx, repository); testErr != nil {
			return testErr
		}
		if testErr := verifyAchievement(ctx, repository, playerID); testErr != nil {
			return testErr
		}
		if testErr := verifyQuest(ctx, repository, playerID); testErr != nil {
			return testErr
		}
		if testErr := verifyAuxiliary(ctx, repository, playerID); testErr != nil {
			return testErr
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("unexpected integration result: %v", err)
	}
}

// verifyCatalog checks the complete seeded catalog can be grouped and decoded.
func verifyCatalog(ctx context.Context, repository *Repository) error {
	catalog, err := repository.Catalog(ctx)
	if err != nil {
		return err
	}
	if len(catalog.Achievements) != 51 || len(catalog.Talents) != 6 || len(catalog.Quests) != 8 || len(catalog.Quizzes) != 1 || len(catalog.Promos) != 1 {
		return errors.New("unexpected seeded progression catalog")
	}
	return nil
}

// verifyAchievement checks locking, daily idempotence, score, and reset behavior.
func verifyAchievement(ctx context.Context, repository *Repository, playerID int64) error {
	catalog, err := repository.Catalog(ctx)
	if err != nil {
		return err
	}
	definition := catalog.Achievements[0]
	mutation, err := repository.MutateProgress(ctx, playerID, definition, definition.Levels[0].ProgressNeeded, false)
	if err != nil || mutation.After.Level != 1 || len(mutation.Crossed) != 1 {
		return errors.New("achievement level mutation failed")
	}
	daily, err := repository.MutateProgress(ctx, playerID, definition, 1, true)
	if err != nil {
		return err
	}
	repeated, err := repository.MutateProgress(ctx, playerID, definition, 1, true)
	if err != nil || repeated.After.Progress != daily.After.Progress {
		return errors.New("daily achievement mutation was not idempotent")
	}
	if score, scoreErr := repository.AddScore(ctx, playerID, 10); scoreErr != nil || score != 10 {
		return errors.New("achievement score mutation failed")
	}
	changed, err := repository.ResetProgress(ctx, playerID, definition.ID)
	if err != nil || !changed {
		return errors.New("achievement reset failed")
	}
	return nil
}

// verifyQuest checks activation, completion, history, daily rejection, and replay protection.
func verifyQuest(ctx context.Context, repository *Repository, playerID int64) error {
	const questID int64 = 940001
	if _, err := repository.ActivateQuest(ctx, playerID, questID); err != nil {
		return err
	}
	state, err := repository.IncrementQuest(ctx, playerID, questID, 1, 1)
	if err != nil || state.Progress != 1 {
		return errors.New("quest increment failed")
	}
	completed, err := repository.CompleteQuest(ctx, playerID, questID)
	if err != nil || !completed {
		return errors.New("quest completion failed")
	}
	if _, err = repository.ActivateQuest(ctx, playerID, questID); !errors.Is(err, progressionrecord.ErrConflict) {
		return errors.New("completed quest could be replayed")
	}
	if err = repository.RejectDailyQuest(ctx, playerID, state.AcceptedAt); err != nil {
		return err
	}
	rejected, err := repository.DailyQuestRejected(ctx, playerID, state.AcceptedAt)
	if err != nil || !rejected {
		return errors.New("daily quest rejection failed")
	}
	return nil
}

// verifyAuxiliary checks talent, quiz, and promotional claim persistence.
func verifyAuxiliary(ctx context.Context, repository *Repository, playerID int64) error {
	changed, err := repository.SetTalent(ctx, playerID, "citizenship", 1)
	if err != nil || !changed {
		return errors.New("talent advance failed")
	}
	if err = repository.ForceTalent(ctx, playerID, "citizenship", 0); err != nil {
		return err
	}
	if _, err = repository.SaveQuizResult(ctx, playerID, "SafetyQuiz", true, nil); err != nil {
		return err
	}
	if passed, passErr := repository.QuizPassed(ctx, playerID, "SafetyQuiz"); passErr != nil || !passed {
		return errors.New("quiz result failed")
	}
	catalog, err := repository.Catalog(ctx)
	if err != nil {
		return err
	}
	claimed, err := repository.ClaimPromo(ctx, playerID, catalog.Promos[0], false)
	if err != nil || !claimed {
		return errors.New("promotion claim failed")
	}
	return nil
}
