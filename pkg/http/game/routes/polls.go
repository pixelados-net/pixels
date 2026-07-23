package routes

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
)

// listPolls returns every durable poll definition.
func (dependencies Dependencies) listPolls(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, false); err != nil {
		return err
	}
	polls, err := dependencies.Progression.AllPolls(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(polls)
}

// createPoll creates one nested poll definition.
func (dependencies Dependencies) createPoll(ctx *fiber.Ctx) error {
	var request PollRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	definition, err := pollDefinition(request, 0)
	if err != nil {
		return err
	}
	err = dependencies.mutate(ctx, request.AuditRequest, false, "games.poll.create", "poll", func(txCtx context.Context) error {
		definition, err = dependencies.Progression.CreatePoll(txCtx, definition)
		return err
	})
	if err != nil {
		return err
	}
	if err = dependencies.Polls.ReloadDatabase(ctx.Context()); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(definition)
}

// updatePoll replaces one nested poll definition.
func (dependencies Dependencies) updatePoll(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "id")
	if err != nil {
		return err
	}
	var request PollRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	definition, err := pollDefinition(request, int32(id))
	if err != nil {
		return err
	}
	updated := false
	err = dependencies.mutate(ctx, request.AuditRequest, false, "games.poll.update", "poll:"+strconv.FormatInt(id, 10), func(txCtx context.Context) error {
		definition, updated, err = dependencies.Progression.UpdatePoll(txCtx, definition)
		return err
	})
	if err != nil {
		return err
	}
	if !updated {
		return fiber.NewError(fiber.StatusConflict, "poll version conflict")
	}
	if err = dependencies.Polls.ReloadDatabase(ctx.Context()); err != nil {
		return err
	}
	return ctx.JSON(definition)
}

// deletePoll disables one poll.
func (dependencies Dependencies) deletePoll(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "id")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, request, false, "games.poll.disable", "poll:"+strconv.FormatInt(id, 10), func(txCtx context.Context) error {
		_, err = dependencies.Progression.DisablePoll(txCtx, int32(id))
		return err
	})
	if err != nil {
		return err
	}
	if err = dependencies.Polls.ReloadDatabase(ctx.Context()); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// assignPoll assigns or clears one poll room.
func (dependencies Dependencies) assignPoll(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "id")
	if err != nil {
		return err
	}
	var request PollRoomRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	updated := false
	var version int64
	err = dependencies.mutate(ctx, request.AuditRequest, false, "games.poll.assign", "poll:"+strconv.FormatInt(id, 10), func(txCtx context.Context) error {
		version, updated, err = dependencies.Progression.AssignPoll(txCtx, int32(id), request.RoomID, request.Version)
		return err
	})
	if err != nil {
		return err
	}
	if !updated {
		return fiber.NewError(fiber.StatusConflict, "poll version conflict")
	}
	if err = dependencies.Polls.ReloadDatabase(ctx.Context()); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"version": version, "roomId": request.RoomID})
}

// listRoomScores returns paginated durable match history.
func (dependencies Dependencies) listRoomScores(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, true); err != nil {
		return err
	}
	roomID, err := parseID(ctx, "roomId")
	if err != nil {
		return err
	}
	before, _ := strconv.ParseInt(ctx.Query("beforeId"), 10, 64)
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	page, err := dependencies.Scores.List(ctx.Context(), roomID, before, limit)
	if err != nil {
		return err
	}
	return ctx.JSON(page)
}

// reload atomically refreshes Game Center and poll caches.
func (dependencies Dependencies) reload(ctx *fiber.Ctx) error {
	var request AuditRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	if err := dependencies.mutate(ctx, request, true, "games.reload", "games", func(context.Context) error { return nil }); err != nil {
		return err
	}
	if err := dependencies.Lobby.Reload(ctx.Context()); err != nil {
		return err
	}
	if err := dependencies.Polls.ReloadDatabase(ctx.Context()); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"reloaded": true})
}

// pollDefinition validates and maps one poll request.
func pollDefinition(request PollRequest, id int32) (progressionpoll.Definition, error) {
	if strings.TrimSpace(request.Title) == "" || len(request.Questions) == 0 {
		return progressionpoll.Definition{}, fiber.NewError(fiber.StatusUnprocessableEntity, "poll title and questions are required")
	}
	seen := make(map[int32]bool, len(request.Questions))
	for _, question := range request.Questions {
		if question.SortOrder < 0 || question.Type < 0 || question.Type > 2 || strings.TrimSpace(question.Text) == "" || seen[question.SortOrder] {
			return progressionpoll.Definition{}, fiber.NewError(fiber.StatusUnprocessableEntity, "invalid poll question")
		}
		if (question.Type == 1 || question.Type == 2) && !validChoices(question.Choices) {
			return progressionpoll.Definition{}, fiber.NewError(fiber.StatusUnprocessableEntity, "invalid poll choices")
		}
		seen[question.SortOrder] = true
	}
	return progressionpoll.Definition{ID: id, Title: strings.TrimSpace(request.Title), Headline: strings.TrimSpace(request.Headline), Summary: strings.TrimSpace(request.Summary), StartMessage: strings.TrimSpace(request.StartMessage), ThanksMessage: strings.TrimSpace(request.ThanksMessage), RoomID: request.RoomID, RewardBadge: strings.TrimSpace(request.RewardBadge), Questions: request.Questions, Version: request.Version, Enabled: request.Enabled}, nil
}

// validChoices reports whether selectable answers are non-empty and unique.
func validChoices(choices []outcontents.Choice) bool {
	if len(choices) == 0 || len(choices) > 64 {
		return false
	}
	seen := make(map[string]struct{}, len(choices))
	for _, choice := range choices {
		value := strings.TrimSpace(choice.Value)
		if value == "" || strings.TrimSpace(choice.Text) == "" {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
