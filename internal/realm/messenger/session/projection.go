package session

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfriends "github.com/niflaot/pixels/networking/outbound/messenger/friend/list"
	outrequests "github.com/niflaot/pixels/networking/outbound/messenger/friend/requests"
)

// sendFriends sends one complete friend-list fragment.
func (handler Handler) sendFriends(ctx context.Context, connection netconn.Context, playerID int64) error {
	cards, err := handler.Messenger.Cards(ctx, playerID)
	if err != nil {
		return err
	}
	packet, err := outfriends.Encode(1, 0, delivery.FriendCards(cards))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// sendRequests sends incoming request cards and count.
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
