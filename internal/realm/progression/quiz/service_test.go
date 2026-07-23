package quiz

import (
	"context"
	"testing"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// quizCatalogSource returns one fixed quiz catalog.
type quizCatalogSource struct{ value progressionrecord.Catalog }

// Catalog returns the fixed quiz catalog.
func (source quizCatalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return source.value, nil
}

// memoryQuizStore implements focused quiz persistence.
type memoryQuizStore struct {
	progressionrecord.Store
	passed bool
	failed []int32
}

// QuizPassed returns the current durable pass flag.
func (store *memoryQuizStore) QuizPassed(context.Context, int64, string) (bool, error) {
	return store.passed, nil
}

// SaveQuizResult records one attempt while preserving a prior pass.
func (store *memoryQuizStore) SaveQuizResult(_ context.Context, _ int64, _ string, passed bool, failed []int32) (bool, error) {
	store.passed = store.passed || passed
	store.failed = append(store.failed[:0], failed...)
	return true, nil
}

// quizProgressor records safety achievement triggers.
type quizProgressor struct{ calls int }

// ProgressNow records one synchronous progression request.
func (progressor *quizProgressor) ProgressNow(context.Context, int64, string, int64, bool) error {
	progressor.calls++
	return nil
}

// quizFixture builds one loaded safety quiz service.
func quizFixture(t testing.TB) (*Service, *memoryQuizStore, *quizProgressor) {
	t.Helper()
	questions := []progressionrecord.QuizQuestion{{QuestionRef: 10, CorrectAnswerID: 2}, {QuestionRef: 11, CorrectAnswerID: 3}}
	catalog := progressionengine.NewCatalog(quizCatalogSource{value: progressionrecord.Catalog{Quizzes: []progressionrecord.Quiz{{Code: "SafetyQuiz", Kind: "safety", Enabled: true, Questions: questions}}}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	store, progressor := &memoryQuizStore{}, &quizProgressor{}
	return New(catalog, store, progressor), store, progressor
}

// TestQuestionsAreOrdered verifies renderer localization identifiers are preserved.
func TestQuestionsAreOrdered(t *testing.T) {
	service, _, _ := quizFixture(t)
	values, err := service.Questions("safetyquiz")
	if err != nil || len(values) != 2 || values[0] != 10 || values[1] != 11 {
		t.Fatalf("questions=%v err=%v", values, err)
	}
}

// TestSubmitReportsFailuresWithoutAchievement verifies exact failed question identifiers.
func TestSubmitReportsFailuresWithoutAchievement(t *testing.T) {
	service, store, progressor := quizFixture(t)
	failed, err := service.Submit(context.Background(), 42, "SafetyQuiz", []int32{1, 3})
	if err != nil || len(failed) != 1 || failed[0] != 10 {
		t.Fatalf("failed=%v err=%v", failed, err)
	}
	if store.passed || progressor.calls != 0 {
		t.Fatalf("passed=%v progression=%d", store.passed, progressor.calls)
	}
}

// TestSubmitPassesOnlyOnce verifies retries never increment the safety trigger twice.
func TestSubmitPassesOnlyOnce(t *testing.T) {
	service, store, progressor := quizFixture(t)
	for range 2 {
		failed, err := service.Submit(context.Background(), 42, "SafetyQuiz", []int32{2, 3})
		if err != nil || len(failed) != 0 {
			t.Fatalf("failed=%v err=%v", failed, err)
		}
	}
	if !store.passed || progressor.calls != 1 {
		t.Fatalf("passed=%v progression=%d", store.passed, progressor.calls)
	}
}

// TestSubmitRejectsMismatchedAnswers verifies malformed submissions do not persist.
func TestSubmitRejectsMismatchedAnswers(t *testing.T) {
	service, store, progressor := quizFixture(t)
	if _, err := service.Submit(context.Background(), 42, "SafetyQuiz", []int32{2}); err != progressionrecord.ErrInvalid {
		t.Fatalf("error %v", err)
	}
	if store.passed || progressor.calls != 0 {
		t.Fatal("malformed quiz changed state")
	}
}
