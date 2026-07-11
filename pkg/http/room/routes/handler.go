package routes

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
)

// listHandler lists rooms.
func listHandler(rooms roomservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		items, err := listRooms(ctx, rooms)
		if err != nil {
			return err
		}

		return ctx.JSON(ListResponse{Total: len(items), Items: roomResponses(items)})
	}
}

// detailHandler reads one room.
func detailHandler(rooms roomservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		room, err := findRoom(ctx, rooms)
		if err != nil {
			return err
		}

		return ctx.JSON(roomResponse(room))
	}
}

// occupancyHandler reads runtime occupancy.
func occupancyHandler(rooms roomservice.Manager, runtime *roomlive.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		room, err := findRoom(ctx, rooms)
		if err != nil {
			return err
		}

		active, found := runtime.Find(room.ID)
		if !found {
			return ctx.JSON(OccupancyResponse{RoomID: room.ID, MaxUsers: room.MaxUsers})
		}

		return ctx.JSON(occupancyResponse(active.Occupancy()))
	}
}

// closeHandler closes one active room.
func closeHandler(runtime *roomlive.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := roomIDParam(ctx)
		if err != nil {
			return err
		}

		occupancy, found, err := runtime.Close(ctx.Context(), roomID)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "active room not found")
		}

		return ctx.JSON(ActionResponse{Matched: occupancy.Count, Sent: occupancy.Count})
	}
}

// forwardHandler forwards active room occupants to another room.
func forwardHandler(runtime *roomlive.Registry, connections *netconn.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		sourceID, request, err := forwardInput(ctx)
		if err != nil {
			return err
		}

		active, found := runtime.Find(sourceID)
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "active room not found")
		}

		response := forwardOccupants(ctx, active.Occupants(), request.TargetRoomID, connections)
		_, _, _ = runtime.Close(ctx.Context(), sourceID)

		return ctx.JSON(response)
	}
}

// categoriesHandler lists navigator categories.
func categoriesHandler(rooms roomservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		categories, err := rooms.ListCategories(ctx.Context())
		if err != nil {
			return err
		}

		return ctx.JSON(CategoryListResponse{Total: len(categories), Items: categoryResponses(categories)})
	}
}

// liftedHandler lists navigator lifted rooms.
func liftedHandler(navigator navservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		lifted, err := navigator.ListLiftedRooms(ctx.Context())
		if err != nil {
			return err
		}

		return ctx.JSON(LiftedListResponse{Total: len(lifted), Items: liftedResponses(lifted)})
	}
}

// listRooms selects list behavior from query params.
func listRooms(ctx *fiber.Ctx, rooms roomservice.Manager) ([]roommodel.Room, error) {
	limit := ctx.QueryInt("limit", 50)
	query := ctx.Query("query")
	if query != "" {
		return rooms.Search(ctx.Context(), query, limit)
	}

	return rooms.ListPopular(ctx.Context(), limit)
}

// findRoom resolves the route room id.
func findRoom(ctx *fiber.Ctx, rooms roomservice.Manager) (roommodel.Room, error) {
	roomID, err := roomIDParam(ctx)
	if err != nil {
		return roommodel.Room{}, err
	}

	room, found, err := rooms.FindByID(ctx.Context(), roomID)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, fiber.NewError(fiber.StatusNotFound, "room not found")
	}

	return room, nil
}

// roomIDParam parses a room id route parameter.
func roomIDParam(ctx *fiber.Ctx) (int64, error) {
	roomID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || roomID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid room id")
	}

	return roomID, nil
}

// forwardInput parses forwarding input.
func forwardInput(ctx *fiber.Ctx) (int64, ForwardRequest, error) {
	sourceID, err := roomIDParam(ctx)
	if err != nil {
		return 0, ForwardRequest{}, err
	}

	var request ForwardRequest
	if err := ctx.BodyParser(&request); err != nil {
		return 0, ForwardRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid room forward request body")
	}
	if request.TargetRoomID <= 0 || request.TargetRoomID == sourceID {
		return 0, ForwardRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid target room id")
	}

	return sourceID, request, nil
}

// ignored reports whether an error can be ignored.
func ignored(err error) bool {
	return err == nil || errors.Is(err, netconn.ErrDisposed)
}

// liftedResponses maps lifted room rows.
func liftedResponses(rooms []navmodel.LiftedRoom) []LiftedResponse {
	items := make([]LiftedResponse, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, LiftedResponse{ID: room.ID, RoomID: room.RoomID, AreaID: room.AreaID, Image: room.Image, Caption: room.Caption})
	}

	return items
}

// forwardOccupants sends room forward packets.
func forwardOccupants(ctx *fiber.Ctx, occupants []roomlive.Occupant, targetID int64, connections *netconn.Registry) ActionResponse {
	packet, err := outforward.Encode(int32(targetID))
	if err != nil {
		return ActionResponse{Matched: len(occupants), Errors: len(occupants)}
	}

	response := ActionResponse{Matched: len(occupants)}
	for _, occupant := range occupants {
		connection, found := connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			response.Errors++
			continue
		}
		if err := connection.Send(ctx.Context(), packet); ignored(err) {
			response.Sent++
		} else {
			response.Errors++
		}
	}

	return response
}
