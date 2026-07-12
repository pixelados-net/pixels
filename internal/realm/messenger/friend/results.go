package friend

import (
	"context"
	"errors"

	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
	relationchanged "github.com/niflaot/pixels/internal/realm/messenger/friend/events/relationchanged"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	netconn "github.com/niflaot/pixels/networking/connection"
	outerror "github.com/niflaot/pixels/networking/outbound/messenger/friend/requesterror"
	outrequests "github.com/niflaot/pixels/networking/outbound/messenger/friend/requests"
	outupdate "github.com/niflaot/pixels/networking/outbound/messenger/friend/update"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// errorOwnListFull is Nitro's own-list-capacity code.
	errorOwnListFull int32 = 1
	// errorTargetListFull is Nitro's target-list-capacity code.
	errorTargetListFull int32 = 2
	// errorRequestsBlocked is Nitro's target-privacy code.
	errorRequestsBlocked int32 = 3
	// errorTargetNotFound is Nitro's missing-target code.
	errorTargetNotFound int32 = 4
)

// HandleRelation persists and projects one unilateral relationship marker.
func (handler Handler) HandleRelation(ctx context.Context, input RelationCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	if err = handler.Messenger.SetRelation(ctx, actorID, input.PlayerID, input.Relation); err != nil {
		return err
	}
	card, err := handler.Messenger.Card(ctx, actorID, messengermodel.Friendship{FriendPlayerID: input.PlayerID, Relation: input.Relation})
	if err != nil {
		return err
	}
	packet, err := outupdate.Encode([]outupdate.Entry{{Type: outupdate.Changed, Card: delivery.FriendCard(card)}})
	if err != nil {
		return err
	}
	handler.publish(ctx, relationchanged.Name, relationchanged.Payload{PlayerID: actorID, FriendID: input.PlayerID, Relation: input.Relation})
	return input.Connection.Send(ctx, packet)
}

// sendRequestError maps expected request failures to Nitro error codes.
func (handler Handler) sendRequestError(ctx context.Context, connection netconn.Context, failure error) error {
	packet, err := outerror.Encode(0, requestErrorCode(failure))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// sendRequests refreshes incoming requests after a mutation.
func (handler Handler) sendRequests(ctx context.Context, connection netconn.Context, playerID int64) error {
	cards, err := handler.Messenger.RequestCards(ctx, playerID)
	if err != nil {
		return err
	}
	requests := make([]outrequests.Request, len(cards))
	for index, card := range cards {
		requests[index] = outrequests.Request{PlayerID: card.ID, Username: card.Username, Look: card.Look}
	}
	packet, err := outrequests.Encode(int32(len(requests)), requests)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// requestErrorCode maps messenger failures to Nitro's documented codes.
func requestErrorCode(failure error) int32 {
	switch {
	case errors.Is(failure, messengerservice.ErrOwnListFull):
		return errorOwnListFull
	case errors.Is(failure, messengerservice.ErrTargetListFull):
		return errorTargetListFull
	case errors.Is(failure, messengerservice.ErrRequestsBlocked):
		return errorRequestsBlocked
	default:
		return errorTargetNotFound
	}
}

// publish emits one completed messenger mutation.
func (handler Handler) publish(ctx context.Context, name bus.Name, payload any) {
	if handler.Events != nil {
		_ = handler.Events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}
