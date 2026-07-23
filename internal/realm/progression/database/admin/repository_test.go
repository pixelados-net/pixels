package admin

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// TestProgressionAdministrationAgainstPostgres verifies CRUD, optimistic locking, and audit persistence.
func TestProgressionAdministrationAgainstPostgres(t *testing.T) {
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
	rollback := errors.New("rollback progression admin integration")
	err = repository.WithinTransaction(context.Background(), func(ctx context.Context) error {
		definition, createErr := repository.CreateAchievement(ctx, progressionrecord.AchievementDefinition{Name: "Test" + uuid.NewString(), Category: "test", TriggerKey: "test.trigger", Visible: true, Enabled: true})
		if createErr != nil {
			return createErr
		}
		if levelErr := repository.UpsertAchievementLevel(ctx, progressionrecord.AchievementLevel{DefinitionID: definition.ID, Level: 1, ProgressNeeded: 1}); levelErr != nil {
			return levelErr
		}
		updated, updateErr := repository.UpdateAchievement(ctx, definition, definition.Version)
		if updateErr != nil || updated.Version != definition.Version+1 {
			return errors.New("achievement optimistic update failed")
		}
		if _, updateErr = repository.UpdateAchievement(ctx, definition, definition.Version); !errors.Is(updateErr, progressionrecord.ErrConflict) {
			return errors.New("stale achievement version was accepted")
		}
		campaign := progressionrecord.QuestCampaign{Code: "test-" + uuid.NewString(), Enabled: true}
		if createErr = repository.CreateCampaign(ctx, campaign); createErr != nil {
			return createErr
		}
		quest, createErr := repository.CreateQuest(ctx, progressionrecord.QuestDefinition{CampaignCode: campaign.Code, SeriesNumber: 1, Name: "Test", LocalizationCode: "test", TriggerKey: "test.trigger", GoalAmount: 1, RewardKind: "currency", Enabled: true})
		if createErr != nil {
			return createErr
		}
		quest.Name = "Updated"
		quest, updateErr = repository.UpdateQuest(ctx, quest, quest.Version)
		if updateErr != nil || quest.Version != 2 {
			return errors.New("quest optimistic update failed")
		}
		if _, updateErr = repository.UpdateQuest(ctx, quest, 1); !errors.Is(updateErr, progressionrecord.ErrConflict) {
			return errors.New("stale quest version was accepted")
		}
		if createErr = verifyAdminAuxiliary(ctx, repository); createErr != nil {
			return createErr
		}
		if createErr = repository.InsertAudit(ctx, 1, "integration.test", "progression", "verify atomic audit"); createErr != nil {
			return createErr
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("unexpected integration result: %v", err)
	}
}

// verifyAdminAuxiliary checks talent, quiz, and promotion administration.
func verifyAdminAuxiliary(ctx context.Context, repository *Repository) error {
	if err := repository.UpsertTalentLevel(ctx, progressionrecord.TalentLevel{Track: "test", Level: 1, Requirements: []progressionrecord.TalentRequirement{{DefinitionID: 930001, RequiredLevel: 1}}}); err != nil {
		return err
	}
	quizCode := "test-" + uuid.NewString()
	if err := repository.CreateQuiz(ctx, progressionrecord.Quiz{Code: quizCode, Kind: "safety", Enabled: true}); err != nil {
		return err
	}
	question, err := repository.CreateQuizQuestion(ctx, progressionrecord.QuizQuestion{QuizCode: quizCode, QuestionRef: 1, CorrectAnswerID: 2})
	if err != nil {
		return err
	}
	question.CorrectAnswerID = 3
	if changed, updateErr := repository.UpdateQuizQuestion(ctx, question); updateErr != nil || !changed {
		return errors.New("quiz question update failed")
	}
	promo := progressionrecord.PromoBadge{Code: "test-" + uuid.NewString(), BadgeCode: "TEST", Enabled: true}
	if err = repository.CreatePromo(ctx, promo); err != nil {
		return err
	}
	promo.BadgeCode = "TEST2"
	if changed, updateErr := repository.UpdatePromo(ctx, promo); updateErr != nil || !changed {
		return errors.New("promotion update failed")
	}
	return nil
}
