// Package friend handles messenger friendship mutations.
package friend

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
	friendremoved "github.com/niflaot/pixels/internal/realm/messenger/friend/events/removed"
	requestaccepted "github.com/niflaot/pixels/internal/realm/messenger/friend/events/requestaccepted"
	requestdeclined "github.com/niflaot/pixels/internal/realm/messenger/friend/events/requestdeclined"
	requestsent "github.com/niflaot/pixels/internal/realm/messenger/friend/events/requestsent"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	netconn "github.com/niflaot/pixels/networking/connection"
	outaccept "github.com/niflaot/pixels/networking/outbound/messenger/friend/acceptresult"
	outnewrequest "github.com/niflaot/pixels/networking/outbound/messenger/friend/newrequest"
	outupdate "github.com/niflaot/pixels/networking/outbound/messenger/friend/update"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// RequestName identifies friend-request creation.
	RequestName command.Name = "messenger.friend.request"
	// AcceptName identifies friend-request acceptance.
	AcceptName command.Name = "messenger.friend.accept"
	// DeclineName identifies friend-request rejection.
	DeclineName command.Name = "messenger.friend.decline"
	// RemoveName identifies friendship removal.
	RemoveName command.Name = "messenger.friend.remove"
	// RelationName identifies relationship changes.
	RelationName command.Name = "messenger.friend.relation"
)

// RequestCommand requests friendship by username.
type RequestCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Username identifies the target.
	Username string
}

// CommandName returns the request command identity.
func (RequestCommand) CommandName() command.Name { return RequestName }

// BatchCommand carries a bounded player-id collection.
type BatchCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PlayerIDs stores affected players.
	PlayerIDs []int64
	// All reports an all-pending-requests decline.
	All bool
	// Name stores the command identity.
	Name command.Name
}

// CommandName returns the configured batch command identity.
func (input BatchCommand) CommandName() command.Name { return input.Name }

// RelationCommand requests one unilateral relation change.
type RelationCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PlayerID identifies the friend.
	PlayerID int64
	// Relation stores the requested relation.
	Relation messengermodel.Relation
}

// CommandName returns the relation command identity.
func (RelationCommand) CommandName() command.Name { return RelationName }

// Handler executes friendship commands.
type Handler struct {
	// Messenger stores messenger behavior.
	Messenger *messengerservice.Service
	// Delivery sends live packets.
	Delivery *delivery.Sender
	// Events publishes completed mutations.
	Events bus.Publisher
}

// HandleRequest creates and projects one friend request.
func (handler Handler) HandleRequest(ctx context.Context, input RequestCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	result, err := handler.Messenger.SendRequest(ctx, actorID, input.Username)
	if err != nil {
		return handler.sendRequestError(ctx, input.Connection, err)
	}
	if !result.Sent {
		return nil
	}
	actorCard, err := handler.Messenger.Card(ctx, result.Target.Player.ID, messengermodel.Friendship{FriendPlayerID: actorID})
	if err != nil {
		return err
	}
	packet, err := outnewrequest.Encode(actorCard.ID, actorCard.Username, actorCard.Look)
	if err != nil {
		return err
	}
	_, err = handler.Delivery.Send(ctx, result.Target.Player.ID, packet)
	handler.publish(ctx, requestsent.Name, requestsent.Payload{FromPlayerID: actorID, ToPlayerID: result.Target.Player.ID})
	return err
}

// HandleAccept accepts each supplied request and returns native failure data.
func (handler Handler) HandleAccept(ctx context.Context, input BatchCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	failures := make([]outaccept.Failure, 0)
	for _, requesterID := range input.PlayerIDs {
		result, acceptErr := handler.Messenger.Accept(ctx, actorID, requesterID)
		if acceptErr != nil || !result.Accepted {
			failures = append(failures, outaccept.Failure{PlayerID: requesterID, ErrorCode: requestErrorCode(acceptErr)})
			continue
		}
		actorPacket, encodeErr := outupdate.Encode([]outupdate.Entry{{Type: outupdate.Added, Card: delivery.FriendCard(result.ActorCard)}})
		if encodeErr != nil {
			return encodeErr
		}
		if err = input.Connection.Send(ctx, actorPacket); err != nil {
			return err
		}
		requesterPacket, encodeErr := outupdate.Encode([]outupdate.Entry{{Type: outupdate.Added, Card: delivery.FriendCard(result.RequesterCard)}})
		if encodeErr != nil {
			return encodeErr
		}
		_, _ = handler.Delivery.Send(ctx, requesterID, requesterPacket)
		handler.publish(ctx, requestaccepted.Name, requestaccepted.Payload{PlayerOneID: actorID, PlayerTwoID: requesterID})
	}
	packet, err := outaccept.Encode(failures)
	if err != nil {
		return err
	}
	if err = input.Connection.Send(ctx, packet); err != nil {
		return err
	}
	return handler.sendRequests(ctx, input.Connection, actorID)
}

// HandleDecline removes selected or all incoming requests.
func (handler Handler) HandleDecline(ctx context.Context, input BatchCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	requesterIDs := input.PlayerIDs
	if input.All {
		requests, requestErr := handler.Messenger.Requests(ctx, actorID)
		if requestErr != nil {
			return requestErr
		}
		requesterIDs = make([]int64, len(requests))
		for index, request := range requests {
			requesterIDs[index] = request.FromPlayerID
		}
	}
	_, err = handler.Messenger.Decline(ctx, actorID, input.PlayerIDs, input.All)
	if err != nil {
		return err
	}
	for _, requesterID := range requesterIDs {
		handler.publish(ctx, requestdeclined.Name, requestdeclined.Payload{FromPlayerID: requesterID, ToPlayerID: actorID})
	}
	return handler.sendRequests(ctx, input.Connection, actorID)
}

// HandleRemove atomically removes friendships and projects removals to both sides.
func (handler Handler) HandleRemove(ctx context.Context, input BatchCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	removed, err := handler.Messenger.Remove(ctx, actorID, input.PlayerIDs)
	if err != nil {
		return err
	}
	entries := make([]outupdate.Entry, len(removed))
	for index, friendID := range removed {
		entries[index] = outupdate.Entry{Type: outupdate.Removed, PlayerID: friendID}
		packet, encodeErr := outupdate.Encode([]outupdate.Entry{{Type: outupdate.Removed, PlayerID: actorID}})
		if encodeErr != nil {
			return encodeErr
		}
		_, _ = handler.Delivery.Send(ctx, friendID, packet)
		handler.publish(ctx, friendremoved.Name, friendremoved.Payload{PlayerOneID: actorID, PlayerTwoID: friendID})
	}
	packet, err := outupdate.Encode(entries)
	if err != nil {
		return err
	}
	return input.Connection.Send(ctx, packet)
}
