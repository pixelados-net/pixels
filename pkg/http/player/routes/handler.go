package routes

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// idempotencyHeader carries the caller supplied replay key.
	idempotencyHeader = "Idempotency-Key"
	// replayHeader reports that a stored response was returned.
	replayHeader = "Idempotency-Replayed"
)

// handler serves the administrative player HTTP boundary.
type handler struct {
	// players creates and reads persistent players.
	players playerservice.AdminManager
	// idempotency coordinates retry-safe creation.
	idempotency idempotencyStore
	// live stores online player projections.
	live *playerlive.Registry
	// connections stores active network sessions.
	connections *netconn.Registry
	// effects administers durable player effects.
	effects playereffect.Manager
}

// create creates one player and profile atomically.
func (handler handler) create(ctx *fiber.Ctx) error {
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid player request body")
	}

	record, claimed, err := handler.idempotency.claim(ctx.Context(), ctx.Get(idempotencyHeader), request)
	if err != nil {
		return idempotencyError(err)
	}
	if !claimed {
		return handler.replayOrRecover(ctx, record, request)
	}

	created, err := handler.players.Create(ctx.Context(), createParams(request))
	if err != nil {
		_ = handler.idempotency.release(ctx.Context(), ctx.Get(idempotencyHeader))
		return playerError(err)
	}

	response := playerResponse(created)
	if err := handler.idempotency.complete(ctx.Context(), ctx.Get(idempotencyHeader), record, response); err != nil {
		return err
	}

	setPlayerHeaders(ctx, response)
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// replayOrRecover replays a completed request or recovers a lost response.
func (handler handler) replayOrRecover(ctx *fiber.Ctx, record idempotencyRecord, request CreateRequest) error {
	if record.State == "complete" && record.Response != nil {
		ctx.Set(replayHeader, "true")
		setPlayerHeaders(ctx, *record.Response)
		return ctx.JSON(record.Response)
	}

	existing, found, err := handler.players.FindByUsername(ctx.Context(), request.Username)
	if err != nil {
		return playerError(err)
	}
	if !found {
		ctx.Set(fiber.HeaderRetryAfter, "1")
		return idempotencyError(errIdempotencyPending)
	}

	response := playerResponse(existing)
	if err := handler.idempotency.complete(ctx.Context(), ctx.Get(idempotencyHeader), record, response); err != nil {
		return err
	}
	ctx.Set(replayHeader, "true")
	setPlayerHeaders(ctx, response)

	return ctx.JSON(response)
}

// createParams maps HTTP input to the player service contract.
func createParams(request CreateRequest) playerservice.CreateParams {
	return playerservice.CreateParams{
		Username: request.Username,
		Profile: playerservice.CreateProfileParams{
			Look: request.Look, Gender: playermodel.Gender(request.Gender), Motto: request.Motto,
			HomeRoomID: request.HomeRoomID, AllowNameChange: request.AllowNameChange,
		},
	}
}

// playerResponse maps one service record to the administrative response contract.
func playerResponse(record playerservice.Record) Response {
	return Response{
		ID: record.Player.ID, Username: record.Player.Username, Look: record.Profile.Look,
		Gender: string(record.Profile.Gender), Motto: record.Profile.Motto,
		HomeRoomID: record.Profile.HomeRoomID, AllowNameChange: record.Profile.AllowNameChange,
		BubbleStyle: record.Profile.BubbleStyle, BlockFriendRequests: record.Profile.BlockFriendRequests,
		BlockRoomInvites: record.Profile.BlockRoomInvites, BlockFollowing: record.Profile.BlockFollowing,
		ClubLevel: int16(record.Player.Club.Level), ClubExpiresAt: record.Player.Club.ExpiresAt,
		AllowTrade:  record.Player.AllowTrade,
		LastLoginAt: record.Player.LastLoginAt, LastSeenAt: record.Player.LastSeenAt,
		CreatedAt: record.Player.CreatedAt, UpdatedAt: newestTime(record.Player.UpdatedAt, record.Profile.UpdatedAt),
		Version: fmt.Sprintf("%d.%d", record.Player.Version.Version, record.Profile.Version.Version),
	}
}

// setPlayerHeaders applies conditional caching headers.
func setPlayerHeaders(ctx *fiber.Ctx, response Response) {
	ctx.Set(fiber.HeaderETag, etag(response))
	ctx.Set(fiber.HeaderCacheControl, "private, no-cache")
}

// etag creates a quoted representation tag.
func etag(response Response) string {
	roomID := int64(0)
	if response.CurrentRoomID != nil {
		roomID = *response.CurrentRoomID
	}

	return fmt.Sprintf(
		"\"player-%d-%s-%t-%d\"",
		response.ID,
		response.Version,
		response.Online,
		roomID,
	)
}

// idempotencyError maps idempotency outcomes to HTTP errors.
func idempotencyError(err error) error {
	if errors.Is(err, errIdempotencyPending) || errors.Is(err, errIdempotencyConflict) {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return err
}

// playerError maps domain errors to stable HTTP statuses.
func playerError(err error) error {
	switch {
	case errors.Is(err, playerservice.ErrInvalidPlayerID), errors.Is(err, playerservice.ErrInvalidUsername),
		errors.Is(err, playerservice.ErrInvalidLook), errors.Is(err, playerservice.ErrInvalidMotto),
		errors.Is(err, playerservice.ErrInvalidGender), errors.Is(err, playerservice.ErrInvalidHomeRoomID),
		errors.Is(err, playerservice.ErrInvalidBubbleStyle):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, playerservice.ErrUsernameTaken), errors.Is(err, playerservice.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, playerservice.ErrPlayerNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	default:
		if strings.TrimSpace(err.Error()) == "" {
			return fiber.NewError(fiber.StatusInternalServerError, "player request failed")
		}
		return err
	}
}

// newestTime returns the latest of two durable mutation times.
func newestTime(first time.Time, second time.Time) time.Time {
	if second.After(first) {
		return second
	}

	return first
}
