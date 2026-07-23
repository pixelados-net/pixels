package routes

import (
	"context"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
)

var colorPattern = regexp.MustCompile(`^[0-9a-fA-F]{6}$`)

// listCenter returns every external game registration.
func (dependencies Dependencies) listCenter(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, true); err != nil {
		return err
	}
	games, err := dependencies.Center.ListGames(ctx.Context(), false)
	if err != nil {
		return err
	}
	return ctx.JSON(games)
}

// createCenter creates one external game registration.
func (dependencies Dependencies) createCenter(ctx *fiber.Ctx) error {
	var request CenterRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	game, err := centerGame(request, 0)
	if err != nil {
		return err
	}
	err = dependencies.mutate(ctx, request.AuditRequest, true, "games.center.create", "game:center", func(txCtx context.Context) error {
		game, err = dependencies.Center.CreateGame(txCtx, game)
		return err
	})
	if err != nil {
		return err
	}
	if err = dependencies.Lobby.Reload(ctx.Context()); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(game)
}

// updateCenter replaces one game using optimistic concurrency.
func (dependencies Dependencies) updateCenter(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "id")
	if err != nil {
		return err
	}
	var request CenterRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	game, err := centerGame(request, int32(id))
	if err != nil {
		return err
	}
	updated := false
	err = dependencies.mutate(ctx, request.AuditRequest, true, "games.center.update", "game:center", func(txCtx context.Context) error {
		game, updated, err = dependencies.Center.UpdateGame(txCtx, game)
		return err
	})
	if err != nil {
		return err
	}
	if !updated {
		return fiber.NewError(fiber.StatusConflict, "game version conflict")
	}
	if err = dependencies.Lobby.Reload(ctx.Context()); err != nil {
		return err
	}
	return ctx.JSON(game)
}

// deleteCenter disables one game registration.
func (dependencies Dependencies) deleteCenter(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "id")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, request, true, "games.center.disable", "game:center", func(txCtx context.Context) error {
		_, err = dependencies.Center.DisableGame(txCtx, int32(id))
		return err
	})
	if err != nil {
		return err
	}
	if err = dependencies.Lobby.Reload(ctx.Context()); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// putScore stores one external weekly best score.
func (dependencies Dependencies) putScore(ctx *fiber.Ctx) error {
	gameID, err := parseID(ctx, "id")
	if err != nil {
		return err
	}
	playerID, err := parseID(ctx, "playerId")
	if err != nil {
		return err
	}
	var request ScoreRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if request.Year < 2000 || request.Week < 1 || request.Week > 53 || request.Score < 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid weekly score")
	}
	err = dependencies.mutate(ctx, request.AuditRequest, true, "games.center.score", "game:score", func(txCtx context.Context) error {
		return dependencies.Center.UpsertScore(txCtx, int32(gameID), playerID, request.Year, request.Week, request.Score)
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// centerGame validates and maps one registration request.
func centerGame(request CenterRequest, id int32) (gamecenterrecord.Game, error) {
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" || !colorPattern.MatchString(request.BackgroundColor) || !colorPattern.MatchString(request.TextColor) {
		return gamecenterrecord.Game{}, fiber.NewError(fiber.StatusUnprocessableEntity, "invalid game name or colors")
	}
	if request.LaunchKind != gamecenterrecord.LaunchURL && request.LaunchKind != gamecenterrecord.LaunchParameters {
		return gamecenterrecord.Game{}, fiber.NewError(fiber.StatusUnprocessableEntity, "invalid launch kind")
	}
	return gamecenterrecord.Game{ID: id, Name: request.Name, BackgroundColor: strings.ToLower(request.BackgroundColor), TextColor: strings.ToLower(request.TextColor), AssetURL: strings.TrimSpace(request.AssetURL), SupportURL: strings.TrimSpace(request.SupportURL), LaunchURL: strings.TrimSpace(request.LaunchURL), LaunchKind: request.LaunchKind, Enabled: request.Enabled, Version: request.Version}, nil
}
