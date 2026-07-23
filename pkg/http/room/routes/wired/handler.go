package wired

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// Handler serves protected WIRED administration operations.
type Handler struct {
	// dependencies stores focused domain boundaries.
	dependencies Dependencies
}

// registry returns the stable canonical manifest.
func (handler Handler) registry(ctx *fiber.Ctx) error {
	items := handler.dependencies.Registry.Manifest()
	return ctx.JSON(RegistryResponse{Source: "Arcturus Community and Nitro audited manifest", Total: len(items), Items: items})
}

// room returns every configured node and compilation state.
func (handler Handler) room(ctx *fiber.Ctx) error {
	roomID, err := positiveParam(ctx, "id")
	if err != nil {
		return err
	}
	configs, err := handler.dependencies.Store.LoadRoom(ctx.Context(), roomID)
	if err != nil {
		return err
	}
	items := make([]ConfigResponse, 0, len(configs))
	for _, config := range configs {
		descriptor, found := handler.dependencies.Registry.Resolve(config.Interaction)
		if found {
			items = append(items, ConfigResponse{Config: config, Descriptor: descriptor})
		}
	}
	return ctx.JSON(RoomResponse{RoomID: roomID, Loaded: handler.dependencies.Engine.Loaded(roomID), Items: items})
}

// item returns one persisted WIRED node.
func (handler Handler) item(ctx *fiber.Ctx) error {
	roomID, itemID, err := handler.identifiers(ctx)
	if err != nil {
		return err
	}
	config, found, err := handler.dependencies.Store.Find(ctx.Context(), roomID, itemID)
	if err != nil {
		return err
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "WIRED item not found")
	}
	descriptor, found := handler.dependencies.Registry.Resolve(config.Interaction)
	if !found {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "unsupported WIRED item")
	}
	return ctx.JSON(ConfigResponse{Config: config, Descriptor: descriptor})
}

// save validates and atomically replaces one node configuration.
func (handler Handler) save(ctx *fiber.Ctx) error {
	roomID, itemID, err := handler.identifiers(ctx)
	if err != nil {
		return err
	}
	var request ConfigRequest
	if err = ctx.BodyParser(&request); err != nil || request.ExpectedVersion < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid WIRED request")
	}
	config, found, err := handler.dependencies.Store.Find(ctx.Context(), roomID, itemID)
	if err != nil || !found {
		return handler.findError(err)
	}
	config.IntParams = append([]int32(nil), request.IntParams...)
	config.StringParam = request.StringParam
	config.SelectionMode = request.SelectionMode
	config.DelayPulses = request.DelayPulses
	config.Targets = targets(request.TargetIDs)
	if _, err = handler.dependencies.Compiler.CompileNode(config); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	saved, err := handler.dependencies.Store.Save(ctx.Context(), config, request.ExpectedVersion)
	if err != nil {
		return persistenceError(err)
	}
	if err = handler.dependencies.Engine.Reload(ctx.Context(), roomID, time.Now()); err != nil {
		return err
	}
	descriptor, _ := handler.dependencies.Registry.Resolve(saved.Interaction)
	return ctx.JSON(ConfigResponse{Config: saved, Descriptor: descriptor})
}

// snapshot captures selected furniture state and reloads the active generation.
func (handler Handler) snapshot(ctx *fiber.Ctx) error {
	roomID, itemID, err := handler.identifiers(ctx)
	if err != nil {
		return err
	}
	if _, err = handler.dependencies.Store.Capture(ctx.Context(), roomID, itemID); err != nil {
		return persistenceError(err)
	}
	if err = handler.dependencies.Engine.Reload(ctx.Context(), roomID, time.Now()); err != nil {
		return err
	}
	return ctx.JSON(ActionResponse{Success: true})
}

// reload recompiles one room from durable configuration.
func (handler Handler) reload(ctx *fiber.Ctx) error {
	roomID, err := positiveParam(ctx, "id")
	if err != nil {
		return err
	}
	if err = handler.dependencies.Engine.Reload(ctx.Context(), roomID, time.Now()); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	return ctx.JSON(ActionResponse{Success: true})
}

// game performs one explicit QA lifecycle transition.
func (handler Handler) game(ctx *fiber.Ctx) error {
	if handler.dependencies.Games == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "WIRED game service unavailable")
	}
	roomID, err := positiveParam(ctx, "id")
	if err != nil {
		return err
	}
	switch ctx.Params("action") {
	case "start":
		err = handler.dependencies.Games.Start(ctx.Context(), roomID)
	case "end":
		err = handler.dependencies.Games.End(ctx.Context(), roomID)
	case "reset":
		err = handler.dependencies.Games.Reset(ctx.Context(), roomID)
	default:
		return fiber.NewError(fiber.StatusNotFound, "unknown WIRED game action")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	return ctx.JSON(ActionResponse{Success: true})
}

// visibility persists room-level WIRED configuration-box visibility.
func (handler Handler) visibility(ctx *fiber.Ctx) error {
	if handler.dependencies.Rooms == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "room settings service unavailable")
	}
	roomID, err := positiveParam(ctx, "id")
	if err != nil {
		return err
	}
	var request VisibilityRequest
	if err = ctx.BodyParser(&request); err != nil || request.ExpectedVersion < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid WIRED visibility request")
	}
	updated, err := handler.dependencies.Rooms.Update(ctx.Context(), roomID, request.ExpectedVersion, roomservice.UpdateParams{HideWired: &request.HideBoxes})
	if err != nil {
		if errors.Is(err, roomservice.ErrRoomNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		if errors.Is(err, roomservice.ErrVersionConflict) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return err
	}
	return ctx.JSON(VisibilityResponse{RoomID: updated.ID, HideBoxes: updated.HideWired, Version: updated.Version.Version})
}

// traces returns a sanitized oldest-first trace ring.
func (handler Handler) traces(ctx *fiber.Ctx) error {
	roomID, err := positiveParam(ctx, "id")
	if err != nil {
		return err
	}
	traces := handler.dependencies.Engine.Traces(roomID)
	result := make([]TraceResponse, len(traces))
	for index, trace := range traces {
		result[index] = TraceResponse{ID: trace.ID, Kind: uint8(trace.Kind), Stacks: trace.Stacks, Effects: trace.Effects, BudgetExhausted: trace.BudgetExhausted, StartedAt: trace.StartedAt, Duration: trace.Duration}
	}
	return ctx.JSON(result)
}

// metrics returns process-wide low-cardinality WIRED execution counters.
func (handler Handler) metrics(ctx *fiber.Ctx) error {
	if _, err := positiveParam(ctx, "id"); err != nil {
		return err
	}
	return ctx.JSON(handler.dependencies.Engine.Metrics())
}

// identifiers parses a positive room and furniture identifier.
func (handler Handler) identifiers(ctx *fiber.Ctx) (int64, int64, error) {
	roomID, err := positiveParam(ctx, "id")
	if err != nil {
		return 0, 0, err
	}
	itemID, err := positiveParam(ctx, "itemId")
	return roomID, itemID, err
}

// positiveParam parses one positive route identifier.
func positiveParam(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

// targets maps request identifiers into ordered records.
func targets(ids []int64) []record.Target {
	result := make([]record.Target, len(ids))
	for index, id := range ids {
		result[index] = record.Target{ItemID: id}
	}
	return result
}

// findError maps an optional missing item result.
func (handler Handler) findError(err error) error {
	if err != nil {
		return err
	}
	return fiber.NewError(fiber.StatusNotFound, "WIRED item not found")
}

// persistenceError maps stable optimistic and target failures.
func persistenceError(err error) error {
	switch {
	case errors.Is(err, record.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, record.ErrItemMissing), errors.Is(err, record.ErrTargetMissing):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, configuration.ErrInvalid):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	default:
		return err
	}
}
