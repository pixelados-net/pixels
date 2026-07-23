package core

import (
	"context"
	"strings"
	"unicode/utf8"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// SendRequest validates and persists one friend request.
func (service *Service) SendRequest(ctx context.Context, fromID int64, username string) (RequestResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || utf8.RuneCountInString(username) > 15 {
		return RequestResult{}, ErrInvalidUsername
	}
	target, found, err := service.players.FindByUsername(ctx, username)
	if err != nil {
		return RequestResult{}, err
	}
	if !found {
		return RequestResult{}, ErrPlayerNotFound
	}
	result := RequestResult{Target: target}
	if target.Player.ID == fromID {
		return result, nil
	}
	friends, err := service.store.IsFriend(ctx, fromID, target.Player.ID)
	if err != nil || friends {
		return result, err
	}
	pending, err := service.store.HasRequestEither(ctx, fromID, target.Player.ID)
	if err != nil || pending {
		return result, err
	}
	if target.Profile.BlockFriendRequests {
		return result, ErrRequestsBlocked
	}
	actor, found, err := service.players.FindByID(ctx, fromID)
	if err != nil {
		return result, err
	}
	if !found {
		return result, ErrPlayerNotFound
	}
	if full, fullErr := service.listFull(ctx, actor); fullErr != nil || full {
		if fullErr != nil {
			return result, fullErr
		}
		return result, ErrOwnListFull
	}
	if full, fullErr := service.listFull(ctx, target); fullErr != nil || full {
		if fullErr != nil {
			return result, fullErr
		}
		return result, ErrTargetListFull
	}
	result.Sent, err = service.store.CreateRequest(ctx, fromID, target.Player.ID)
	return result, err
}

// Accept validates limits and atomically accepts one incoming request.
func (service *Service) Accept(ctx context.Context, actorID int64, requesterID int64) (AcceptResult, error) {
	actor, actorFound, err := service.players.FindByID(ctx, actorID)
	if err != nil || !actorFound {
		if err != nil {
			return AcceptResult{}, err
		}
		return AcceptResult{}, ErrPlayerNotFound
	}
	requester, requesterFound, err := service.players.FindByID(ctx, requesterID)
	if err != nil || !requesterFound {
		if err != nil {
			return AcceptResult{}, err
		}
		return AcceptResult{}, ErrPlayerNotFound
	}
	if full, fullErr := service.listFull(ctx, actor); fullErr != nil || full {
		if fullErr != nil {
			return AcceptResult{}, fullErr
		}
		return AcceptResult{}, ErrOwnListFull
	}
	if full, fullErr := service.listFull(ctx, requester); fullErr != nil || full {
		if fullErr != nil {
			return AcceptResult{}, fullErr
		}
		return AcceptResult{}, ErrTargetListFull
	}
	accepted, err := service.store.AcceptRequest(ctx, actorID, requesterID)
	if err != nil || !accepted {
		return AcceptResult{Accepted: accepted}, err
	}
	service.invalidateCards(ctx, actorID, requesterID)
	return AcceptResult{Accepted: true, ActorCard: service.cardFromRecord(requester, messengermodel.RelationNone, nil), RequesterCard: service.cardFromRecord(actor, messengermodel.RelationNone, nil)}, nil
}

// Decline deletes selected or all incoming friend requests.
func (service *Service) Decline(ctx context.Context, actorID int64, requesterIDs []int64, all bool) (int64, error) {
	if !all && (len(requesterIDs) == 0 || len(requesterIDs) > 100) {
		return 0, ErrInvalidBatch
	}
	return service.store.DeclineRequests(ctx, actorID, requesterIDs, all)
}

// Requests returns incoming pending requests.
func (service *Service) Requests(ctx context.Context, playerID int64) ([]messengermodel.Request, error) {
	return service.store.ListIncomingRequests(ctx, playerID)
}

// OutgoingRequests returns friend requests sent by one player.
func (service *Service) OutgoingRequests(ctx context.Context, playerID int64) ([]messengermodel.Request, error) {
	return service.store.ListOutgoingRequests(ctx, playerID)
}

// RequestCards returns incoming requests enriched with requester identity.
func (service *Service) RequestCards(ctx context.Context, playerID int64) ([]messengermodel.Card, error) {
	requests, err := service.store.ListIncomingRequests(ctx, playerID)
	if err != nil {
		return nil, err
	}
	cards := make([]messengermodel.Card, 0, len(requests))
	for _, request := range requests {
		record, found, findErr := service.players.FindByID(ctx, request.FromPlayerID)
		if findErr != nil {
			return nil, findErr
		}
		if found {
			cards = append(cards, service.cardFromRecord(record, messengermodel.RelationNone, nil))
		}
	}
	return cards, nil
}

// PendingCount returns incoming pending request count.
func (service *Service) PendingCount(ctx context.Context, playerID int64) (int, error) {
	return service.store.CountIncomingRequests(ctx, playerID)
}

// listFull reports whether one player reached the effective friend capacity.
func (service *Service) listFull(ctx context.Context, record playerservice.Record) (bool, error) {
	limit, err := service.friendLimit(ctx, record)
	if err != nil {
		return false, err
	}
	count, err := service.store.CountFriends(ctx, record.Player.ID)
	return count >= limit, err
}
