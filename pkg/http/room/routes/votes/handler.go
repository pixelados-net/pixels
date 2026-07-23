package votes

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
)

// castHandler casts one permanent room vote.
func castHandler(votes roomvotes.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request CastRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid room vote request body")
		}
		result, err := votes.Cast(ctx.Context(), request.RoomID, request.PlayerID)
		if err != nil {
			return voteError(err)
		}

		return ctx.JSON(CastResponse{RoomID: request.RoomID, PlayerID: request.PlayerID, Score: result.Score, Inserted: result.Inserted})
	}
}

// statusHandler reads one player's room vote state.
func statusHandler(votes roomvotes.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, playerID, err := statusInput(ctx)
		if err != nil {
			return err
		}
		state, err := votes.State(ctx.Context(), roomID, playerID)
		if err != nil {
			return voteError(err)
		}

		return ctx.JSON(StatusResponse{RoomID: roomID, PlayerID: playerID, Score: state.Score, CanVote: state.CanVote, Voted: state.Voted})
	}
}

// listHandler lists durable room votes.
func listHandler(votes roomvotes.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		query, err := listInput(ctx)
		if err != nil {
			return err
		}
		items, err := votes.List(ctx.Context(), query)
		if err != nil {
			return voteError(err)
		}
		response := make([]VoteResponse, len(items))
		for index := range items {
			response[index] = VoteResponse{RoomID: items[index].RoomID, PlayerID: items[index].PlayerID, CreatedAt: items[index].CreatedAt}
		}

		return ctx.JSON(ListResponse{Total: len(response), Items: response})
	}
}

// statusInput parses required status query fields.
func statusInput(ctx *fiber.Ctx) (int64, int64, error) {
	roomID, err := positiveQuery(ctx, "roomId")
	if err != nil {
		return 0, 0, err
	}
	playerID, err := positiveQuery(ctx, "playerId")
	if err != nil {
		return 0, 0, err
	}

	return roomID, playerID, nil
}

// listInput parses optional room vote list filters.
func listInput(ctx *fiber.Ctx) (roomvotes.Query, error) {
	roomID, err := positiveQuery(ctx, "roomId")
	if err != nil {
		return roomvotes.Query{}, err
	}
	query := roomvotes.Query{RoomID: roomID, Limit: ctx.QueryInt("limit", roomvotes.DefaultLimit)}
	if raw := ctx.Query("playerId"); raw != "" {
		playerID, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || playerID <= 0 {
			return roomvotes.Query{}, fiber.NewError(fiber.StatusBadRequest, "invalid playerId query parameter")
		}
		query.PlayerID = &playerID
	}
	if raw := ctx.Query("before"); raw != "" {
		before, parseErr := time.Parse(time.RFC3339, raw)
		if parseErr != nil {
			return roomvotes.Query{}, fiber.NewError(fiber.StatusBadRequest, "invalid before query parameter")
		}
		query.Before = &before
	}
	if _, err := query.Normalize(); err != nil {
		return roomvotes.Query{}, voteError(err)
	}

	return query, nil
}

// positiveQuery parses one required positive integer query parameter.
func positiveQuery(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Query(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name+" query parameter")
	}

	return value, nil
}

// voteError maps vote domain errors to meaningful HTTP failures.
func voteError(err error) error {
	switch {
	case errors.Is(err, roomvotes.ErrInvalidRoomID), errors.Is(err, roomvotes.ErrInvalidPlayerID):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, roomvotes.ErrRoomNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	default:
		return err
	}
}
