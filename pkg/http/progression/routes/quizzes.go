package routes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
)

// QuizRequest aliases the quiz definition payload.
type QuizRequest = progressionrequest.Quiz

// QuizQuestionRequest aliases the quiz question payload.
type QuizQuestionRequest = progressionrequest.QuizQuestion

// PollRequest aliases the live word-quiz payload.
type PollRequest = progressionrequest.Poll

// registerQuizzes mounts quiz definition and result administration.
func registerQuizzes(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/quizzes", dependencies.listQuizzes)
	app.Post(basePath+"/quizzes", dependencies.createQuiz)
	app.Delete(basePath+"/quizzes/:code", dependencies.disableQuiz)
	app.Post(basePath+"/quizzes/:code/questions", dependencies.createQuizQuestion)
	app.Patch(basePath+"/quizzes/:code/questions/:id", dependencies.updateQuizQuestion)
	app.Delete(basePath+"/quizzes/:code/questions/:id", dependencies.deleteQuizQuestion)
	app.Get(basePath+"/players/:playerId/quizzes/:code", dependencies.playerQuizResult)
	app.Post(basePath+"/polls", dependencies.startPoll)
}

// startPoll launches one ephemeral room word quiz through the audited admin surface.
func (dependencies Dependencies) startPoll(ctx *fiber.Ctx) error {
	var request PollRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	request.Question = strings.TrimSpace(request.Question)
	if request.RoomID <= 0 || request.Question == "" || request.DurationSeconds < 5 || request.DurationSeconds > 600 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid room poll")
	}
	if err := validateAudit(request.Audit); err != nil {
		return err
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, progressionpolicy.ManageQuests); err != nil {
		return err
	}
	id, err := dependencies.Polls.Start(ctx.Context(), request.RoomID, request.ActorPlayerID, request.Question, time.Duration(request.DurationSeconds)*time.Second)
	if errors.Is(err, progressionpoll.ErrForbidden) {
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}
	if errors.Is(err, progressionpoll.ErrUnavailable) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if err != nil {
		return err
	}
	if err = dependencies.Admin.WithinTransaction(ctx.Context(), func(txCtx context.Context) error {
		return dependencies.Admin.InsertAudit(txCtx, request.ActorPlayerID, "poll.start", entityID("poll", id), request.Reason)
	}); err != nil {
		dependencies.Polls.Cancel(id)
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"pollId": id})
}

// listQuizzes returns the current immutable quiz catalog.
func (dependencies Dependencies) listQuizzes(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageQuests); err != nil {
		return err
	}
	generation := dependencies.Catalog.Current()
	if generation == nil {
		return ctx.JSON([]progressionrecord.Quiz{})
	}
	return ctx.JSON(generation.Catalog.Quizzes)
}

// createQuiz inserts one quiz without hot-reloading the catalog.
func (dependencies Dependencies) createQuiz(ctx *fiber.Ctx) error {
	var request QuizRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := progressionrecord.Quiz{Code: strings.ToUpper(strings.TrimSpace(request.Code)), Kind: strings.ToLower(strings.TrimSpace(request.Kind)), Enabled: true}
	if value.Code == "" || value.Kind != "safety" && value.Kind != "poll" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid quiz")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "quiz.create", entityID("quiz", value.Code), func(txCtx context.Context) error { return dependencies.Admin.CreateQuiz(txCtx, value) })
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// disableQuiz soft-disables one quiz definition.
func (dependencies Dependencies) disableQuiz(ctx *fiber.Ctx) error {
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	var request AuditRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request, "quiz.disable", entityID("quiz", code), func(txCtx context.Context) error {
		changed, disableErr := dependencies.Admin.DisableQuiz(txCtx, code)
		if disableErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return disableErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// createQuizQuestion inserts one quiz question.
func (dependencies Dependencies) createQuizQuestion(ctx *fiber.Ctx) error {
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	var request QuizQuestionRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := progressionrecord.QuizQuestion{QuizCode: code, QuestionRef: request.QuestionRef, CorrectAnswerID: request.CorrectAnswerID}
	if code == "" || value.QuestionRef <= 0 || value.CorrectAnswerID <= 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid quiz question")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "quiz.question.create", entityID("quiz-question", code), func(txCtx context.Context) error {
		var createErr error
		value, createErr = dependencies.Admin.CreateQuizQuestion(txCtx, value)
		return createErr
	})
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// updateQuizQuestion replaces one quiz question.
func (dependencies Dependencies) updateQuizQuestion(ctx *fiber.Ctx) error {
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request QuizQuestionRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	value := progressionrecord.QuizQuestion{ID: id, QuizCode: code, QuestionRef: request.QuestionRef, CorrectAnswerID: request.CorrectAnswerID}
	err = dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "quiz.question.update", entityID("quiz-question", id), func(txCtx context.Context) error {
		changed, updateErr := dependencies.Admin.UpdateQuizQuestion(txCtx, value)
		if updateErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return updateErr
	})
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// deleteQuizQuestion removes one quiz question.
func (dependencies Dependencies) deleteQuizQuestion(ctx *fiber.Ctx) error {
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageQuests, request, "quiz.question.delete", entityID("quiz-question", id), func(txCtx context.Context) error {
		changed, deleteErr := dependencies.Admin.DeleteQuizQuestion(txCtx, code, id)
		if deleteErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return deleteErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// playerQuizResult returns one durable player quiz result.
func (dependencies Dependencies) playerQuizResult(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.OverridePlayers); err != nil {
		return err
	}
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	value, found, err := dependencies.Admin.QuizResult(ctx.Context(), playerID, strings.ToUpper(strings.TrimSpace(ctx.Params("code"))))
	if err != nil {
		return err
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "quiz result not found")
	}
	return ctx.JSON(value)
}
