// Package quiz implements security quiz evaluation.
package quiz

import (
	"context"
	"strings"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// Service owns quiz catalogs and durable evaluation.
type Service struct {
	// catalog owns immutable quiz definitions.
	catalog *progressionengine.Catalog
	// store persists quiz attempts.
	store progressionrecord.Store
	// progression grants the safety achievement through its normal trigger.
	progression Progressor
}

// Progressor grants a passed safety quiz through the normal achievement engine.
type Progressor interface {
	// ProgressNow applies one trigger synchronously.
	ProgressNow(context.Context, int64, string, int64, bool) error
}

// New creates a quiz service.
func New(catalog *progressionengine.Catalog, store progressionrecord.Store, progression Progressor) *Service {
	return &Service{catalog: catalog, store: store, progression: progression}
}

// Questions returns ordered client question identifiers.
func (service *Service) Questions(code string) ([]int32, error) {
	quiz, found := service.find(code)
	if !found {
		return nil, progressionrecord.ErrNotFound
	}
	ids := make([]int32, len(quiz.Questions))
	for index, question := range quiz.Questions {
		ids[index] = question.QuestionRef
	}
	return ids, nil
}

// Submit evaluates ordered answers and returns failed question identifiers.
func (service *Service) Submit(ctx context.Context, playerID int64, code string, answers []int32) ([]int32, error) {
	quiz, found := service.find(code)
	if !found || len(answers) != len(quiz.Questions) {
		return nil, progressionrecord.ErrInvalid
	}
	failed := make([]int32, 0)
	for index, question := range quiz.Questions {
		if answers[index] != question.CorrectAnswerID {
			failed = append(failed, question.QuestionRef)
		}
	}
	passed := len(failed) == 0
	alreadyPassed, err := service.store.QuizPassed(ctx, playerID, quiz.Code)
	if err != nil {
		return nil, err
	}
	if passed && !alreadyPassed {
		if err := service.progression.ProgressNow(ctx, playerID, "quiz.safety.passed", 1, true); err != nil {
			return nil, err
		}
	}
	if _, err = service.store.SaveQuizResult(ctx, playerID, quiz.Code, passed, failed); err != nil {
		return nil, err
	}
	return failed, nil
}

// find resolves one enabled quiz case-insensitively.
func (service *Service) find(code string) (progressionrecord.Quiz, bool) {
	generation := service.catalog.Current()
	if generation == nil {
		return progressionrecord.Quiz{}, false
	}
	quiz := generation.QuizByCode[strings.ToUpper(strings.TrimSpace(code))]
	if quiz == nil {
		return progressionrecord.Quiz{}, false
	}
	return *quiz, true
}
