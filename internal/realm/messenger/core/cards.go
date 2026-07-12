package core

import (
	"context"
	"time"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// PresenceUpdate contains one viewer-specific friend-card update.
type PresenceUpdate struct {
	// PlayerID identifies the receiving friend.
	PlayerID int64
	// Card stores the changed player's projection for that viewer.
	Card messengermodel.Card
}

// Card builds one viewer-specific friend projection.
func (service *Service) Card(ctx context.Context, viewerID int64, friendship messengermodel.Friendship) (messengermodel.Card, error) {
	record, found, err := service.players.FindByID(ctx, friendship.FriendPlayerID)
	if err != nil {
		return messengermodel.Card{}, err
	}
	if !found {
		return messengermodel.Card{}, ErrPlayerNotFound
	}
	card := service.cardFromRecord(record, friendship.Relation, friendship.CategoryID)

	return card, nil
}

// Cards returns a player's complete friend list.
func (service *Service) Cards(ctx context.Context, playerID int64) ([]messengermodel.Card, error) {
	friendships, err := service.store.ListFriends(ctx, playerID)
	if err != nil {
		return nil, err
	}
	cards := make([]messengermodel.Card, 0, len(friendships))
	for _, friendship := range friendships {
		card, cardErr := service.Card(ctx, playerID, friendship)
		if cardErr != nil {
			return nil, cardErr
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// PresenceUpdates builds viewer-specific cards for selected friends of one player.
func (service *Service) PresenceUpdates(ctx context.Context, playerID int64, include func(int64) bool) ([]PresenceUpdate, error) {
	record, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrPlayerNotFound
	}
	followers, err := service.store.ListFollowers(ctx, playerID)
	if err != nil {
		return nil, err
	}
	updates := make([]PresenceUpdate, 0, len(followers))
	for _, friendship := range followers {
		if include != nil && !include(friendship.PlayerID) {
			continue
		}
		updates = append(updates, PresenceUpdate{PlayerID: friendship.PlayerID, Card: service.cardFromRecord(record, friendship.Relation, friendship.CategoryID)})
	}
	return updates, nil
}

// Limits returns the effective, normal, and club friend limits for one player.
func (service *Service) Limits(ctx context.Context, playerID int64) (int, int, int, error) {
	record, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return 0, 0, 0, err
	}
	if !found {
		return 0, 0, 0, ErrPlayerNotFound
	}
	effective, err := service.friendLimit(ctx, record)
	return effective, service.options.MaxFriends, service.options.MaxFriendsClub, err
}

// Profile returns one public player profile with viewer relationship state.
func (service *Service) Profile(ctx context.Context, viewerID int64, targetID int64) (Profile, error) {
	record, found, err := service.players.FindByID(ctx, targetID)
	if err != nil {
		return Profile{}, err
	}
	if !found {
		return Profile{}, ErrPlayerNotFound
	}
	result := Profile{Record: record}
	if livePlayer, online := service.live.Find(targetID); online {
		result.Online = true
		_, result.InRoom = livePlayer.CurrentRoom()
	}
	friendships, err := service.store.ListFriends(ctx, viewerID)
	if err != nil {
		return Profile{}, err
	}
	for _, friendship := range friendships {
		if friendship.FriendPlayerID == targetID {
			result.IsFriend = true
			result.Relation = friendship.Relation
			break
		}
	}
	result.FriendCount, err = service.store.CountFriends(ctx, targetID)
	if err != nil {
		return Profile{}, err
	}
	result.RequestSent, err = service.store.HasRequest(ctx, viewerID, targetID)
	if err != nil {
		return Profile{}, err
	}

	return result, nil
}

// cardFromRecord maps durable and live player state into one friend card.
func (service *Service) cardFromRecord(record playerservice.Record, relation messengermodel.Relation, categoryID *int64) messengermodel.Card {
	card := messengermodel.Card{ID: record.Player.ID, Username: record.Player.Username, Look: record.Profile.Look, Motto: record.Profile.Motto, Relation: relation}
	if record.Profile.Gender != "M" {
		card.Gender = 1
	}
	if categoryID != nil {
		card.CategoryID = int32(*categoryID)
	}
	if livePlayer, online := service.live.Find(record.Player.ID); online {
		card.Online = true
		_, inRoom := livePlayer.CurrentRoom()
		card.FollowingAllowed = inRoom && !record.Profile.BlockFollowing
	}
	return card
}

// friendLimit returns one player's effective friend-list capacity.
func (service *Service) friendLimit(ctx context.Context, record playerservice.Record) (int, error) {
	if service.permissions != nil && service.unlimitedNode != "" {
		allowed, err := service.permissions.HasPermission(ctx, record.Player.ID, service.unlimitedNode)
		if err != nil {
			return 0, err
		}
		if allowed {
			return int(^uint32(0) >> 1), nil
		}
	}
	if record.Player.Club.ActiveAt(time.Now()) {
		return service.options.MaxFriendsClub, nil
	}
	return service.options.MaxFriends, nil
}

// Remove deletes both directions for a bounded friend batch.
func (service *Service) Remove(ctx context.Context, actorID int64, friendIDs []int64) ([]int64, error) {
	if len(friendIDs) == 0 || len(friendIDs) > 100 {
		return nil, ErrInvalidBatch
	}
	return service.store.RemoveFriends(ctx, actorID, friendIDs)
}

// SetRelation persists one unilateral relationship marker.
func (service *Service) SetRelation(ctx context.Context, actorID int64, friendID int64, relation messengermodel.Relation) error {
	if !relation.Valid() {
		return ErrInvalidRelation
	}
	updated, err := service.store.SetRelation(ctx, actorID, friendID, relation)
	if err != nil {
		return err
	}
	if !updated {
		return ErrNotFriend
	}
	return nil
}

// FriendIDs returns one player's directional friend ids.
func (service *Service) FriendIDs(ctx context.Context, playerID int64) ([]int64, error) {
	return service.store.ListFriendIDs(ctx, playerID)
}

// IsFriend reports whether one player lists another.
func (service *Service) IsFriend(ctx context.Context, playerID int64, friendID int64) (bool, error) {
	return service.store.IsFriend(ctx, playerID, friendID)
}
