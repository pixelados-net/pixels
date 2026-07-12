package routes

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	outupdate "github.com/niflaot/pixels/networking/outbound/messenger/friend/update"
)

// friendsHandler lists durable friendships with live presence projection.
func friendsHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := pathID(ctx, "playerId")
		if err != nil {
			return err
		}
		cards, err := dependencies.Messenger.Cards(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		friends := make([]FriendResponse, len(cards))
		for index, card := range cards {
			friends[index] = FriendResponse{PlayerID: card.ID, Username: card.Username, Online: card.Online, InRoom: card.FollowingAllowed, Relation: int16(card.Relation)}
		}
		return ctx.JSON(FriendsResponse{PlayerID: playerID, Friends: friends})
	}
}

// requestsHandler lists incoming and outgoing pending requests.
func requestsHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := pathID(ctx, "playerId")
		if err != nil {
			return err
		}
		incoming, err := dependencies.Messenger.Requests(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		outgoing, err := dependencies.Messenger.OutgoingRequests(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		return ctx.JSON(RequestsResponse{PlayerID: playerID, Incoming: mapRequests(incoming), Outgoing: mapRequests(outgoing)})
	}
}

// removeHandler removes one friendship and updates online clients.
func removeHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := pathID(ctx, "playerId")
		if err != nil {
			return err
		}
		friendID, err := pathID(ctx, "friendId")
		if err != nil {
			return err
		}
		removed, err := dependencies.Messenger.Remove(ctx.Context(), playerID, []int64{friendID})
		if err != nil {
			return err
		}
		if len(removed) > 0 && dependencies.Delivery != nil {
			projectRemoval(ctx.Context(), dependencies, playerID, friendID)
		}
		return ctx.JSON(MutationResponse{PlayerID: playerID, FriendPlayerID: friendID, Removed: len(removed) > 0})
	}
}

// privacyHandler patches messenger privacy while preserving omitted fields.
func privacyHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := pathID(ctx, "playerId")
		if err != nil {
			return err
		}
		var request PrivacyRequest
		if err = ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid messenger privacy request")
		}
		profile, err := dependencies.Messenger.Profile(ctx.Context(), playerID, playerID)
		if err != nil {
			return err
		}
		params := playerservice.PrivacyParams{BlockFriendRequests: profile.Record.Profile.BlockFriendRequests, BlockRoomInvites: profile.Record.Profile.BlockRoomInvites, BlockFollowing: profile.Record.Profile.BlockFollowing}
		applyPrivacy(&params, request)
		record, err := dependencies.Messenger.UpdatePrivacy(ctx.Context(), playerID, params)
		if err != nil {
			return err
		}
		return ctx.JSON(PrivacyResponse{PlayerID: playerID, BlockFriendRequests: record.Profile.BlockFriendRequests, BlockRoomInvites: record.Profile.BlockRoomInvites, BlockFollowing: record.Profile.BlockFollowing})
	}
}

// pathID parses one positive integer path parameter.
func pathID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid messenger player id")
	}
	return value, nil
}

// mapRequests maps domain requests into HTTP records.
func mapRequests(items []messengermodel.Request) []RequestResponse {
	result := make([]RequestResponse, len(items))
	for index, item := range items {
		result[index] = RequestResponse{FromPlayerID: item.FromPlayerID, ToPlayerID: item.ToPlayerID, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339)}
	}
	return result
}

// applyPrivacy applies only supplied privacy fields.
func applyPrivacy(params *playerservice.PrivacyParams, request PrivacyRequest) {
	if request.BlockFriendRequests != nil {
		params.BlockFriendRequests = *request.BlockFriendRequests
	}
	if request.BlockRoomInvites != nil {
		params.BlockRoomInvites = *request.BlockRoomInvites
	}
	if request.BlockFollowing != nil {
		params.BlockFollowing = *request.BlockFollowing
	}
}

// projectRemoval sends symmetric friend-list removal deltas.
func projectRemoval(ctx context.Context, dependencies Dependencies, playerID int64, friendID int64) {
	playerPacket, _ := outupdate.Encode([]outupdate.Entry{{Type: outupdate.Removed, PlayerID: friendID}})
	friendPacket, _ := outupdate.Encode([]outupdate.Entry{{Type: outupdate.Removed, PlayerID: playerID}})
	_, _ = dependencies.Delivery.Send(ctx, playerID, playerPacket)
	_, _ = dependencies.Delivery.Send(ctx, friendID, friendPacket)
}
