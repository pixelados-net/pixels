package core

import (
	"context"
	"math/rand/v2"
	"strconv"
	"strings"
	"unicode/utf8"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// InviteResult describes one room-invite batch.
type InviteResult struct {
	// Delivered stores online friend recipients.
	Delivered []int64
	// NotFriends stores recipients outside the actor's friend list.
	NotFriends []int64
	// Blocked stores recipients whose privacy rejected delivery.
	Blocked []int64
	// RoomID stores the actor's current room.
	RoomID int64
	// Message stores the accepted filtered message.
	Message string
}

// FollowResult describes one accepted follow lookup.
type FollowResult struct {
	// RoomID stores the destination room when forwarding is needed.
	RoomID int64
	// SameRoom reports that actor and friend already share a room.
	SameRoom bool
}

// Profile stores one public profile projection.
type Profile struct {
	// Record stores durable identity and presentation.
	Record playerservice.Record
	// Online reports active player presence.
	Online bool
	// InRoom reports active room presence.
	InRoom bool
	// IsFriend reports whether the viewer lists the target.
	IsFriend bool
	// Relation stores the viewer's unilateral relationship marker.
	Relation messengermodel.Relation
	// FriendCount stores the target player's friendship count.
	FriendCount int
	// RequestSent reports whether the viewer has a pending request to the target.
	RequestSent bool
}

// PrivateMessage describes one accepted live private message.
type PrivateMessage struct {
	// Message stores filtered message text.
	Message string
	// Deliver reports whether the recipient is online.
	Deliver bool
	// Throttled reports a silently ignored send.
	Throttled bool
}

// Invite validates one room invitation batch.
func (service *Service) Invite(ctx context.Context, actorID int64, targetIDs []int64, message string) (InviteResult, error) {
	if len(targetIDs) == 0 || len(targetIDs) > 100 {
		return InviteResult{}, ErrInvalidBatch
	}
	actor, online := service.live.Find(actorID)
	if !online {
		return InviteResult{}, ErrNoActiveRoom
	}
	roomID, inRoom := actor.CurrentRoom()
	if !inRoom {
		return InviteResult{}, ErrNoActiveRoom
	}
	message = truncate(strings.TrimSpace(message), 255)
	if service.filter != nil {
		message, _ = service.filter.Censor(message)
	}
	if message == "" {
		return InviteResult{}, ErrInvalidMessage
	}
	result := InviteResult{RoomID: roomID, Message: message, Delivered: make([]int64, 0, len(targetIDs))}
	for _, targetID := range targetIDs {
		friend, err := service.store.IsFriend(ctx, actorID, targetID)
		if err != nil {
			return InviteResult{}, err
		}
		if !friend {
			result.NotFriends = append(result.NotFriends, targetID)
			continue
		}
		record, found, err := service.players.FindByID(ctx, targetID)
		if err != nil {
			return InviteResult{}, err
		}
		if !found || record.Profile.BlockRoomInvites {
			result.Blocked = append(result.Blocked, targetID)
			continue
		}
		if _, online = service.live.Find(targetID); online {
			result.Delivered = append(result.Delivered, targetID)
		}
	}
	return result, nil
}

// Follow resolves one friend room destination under privacy rules.
func (service *Service) Follow(ctx context.Context, actorID int64, friendID int64) (FollowResult, error) {
	friend, err := service.store.IsFriend(ctx, actorID, friendID)
	if err != nil {
		return FollowResult{}, err
	}
	if !friend {
		return FollowResult{}, ErrNotFriend
	}
	target, online := service.live.Find(friendID)
	if !online {
		return FollowResult{}, ErrFriendOffline
	}
	record, found, err := service.players.FindByID(ctx, friendID)
	if err != nil || !found {
		if err != nil {
			return FollowResult{}, err
		}
		return FollowResult{}, ErrPlayerNotFound
	}
	if record.Profile.BlockFollowing {
		allowed := false
		if service.permissions != nil && service.followNode != "" {
			allowed, err = service.permissions.HasPermission(ctx, actorID, service.followNode)
		}
		if err != nil {
			return FollowResult{}, err
		}
		if !allowed {
			return FollowResult{}, ErrFollowingBlocked
		}
	}
	roomID, inRoom := target.CurrentRoom()
	if !inRoom {
		return FollowResult{}, ErrFriendNotInRoom
	}
	if actor, actorOnline := service.live.Find(actorID); actorOnline {
		current, present := actor.CurrentRoom()
		if present && current == roomID {
			return FollowResult{SameRoom: true}, nil
		}
	}
	return FollowResult{RoomID: roomID}, nil
}

// FindPopulatedRoom returns one random active categorized room with occupants.
func (service *Service) FindPopulatedRoom() (int64, error) {
	candidates := make([]int64, 0)
	for _, active := range service.rooms.Snapshot() {
		snapshot := active.Snapshot()
		if snapshot.CategoryID != nil && active.Occupancy().Count > 0 {
			candidates = append(candidates, snapshot.ID)
		}
	}
	if len(candidates) == 0 {
		return 0, ErrNoPopulatedRoom
	}
	return candidates[rand.IntN(len(candidates))], nil
}

// SendPrivate validates one live private message.
func (service *Service) SendPrivate(ctx context.Context, fromID int64, toID int64, message string) (PrivateMessage, error) {
	allowed, err := service.redis.SetIfAbsent(ctx, "messenger:chat:throttle:"+integerKey(fromID), []byte{1}, service.options.ChatThrottle)
	if err != nil {
		return PrivateMessage{}, err
	}
	if !allowed {
		return PrivateMessage{Throttled: true}, nil
	}
	friend, err := service.store.IsFriend(ctx, fromID, toID)
	if err != nil {
		return PrivateMessage{}, err
	}
	if !friend {
		return PrivateMessage{}, ErrNotFriend
	}
	if target, online := service.live.Find(toID); online && target.IsIgnoring(fromID) {
		return PrivateMessage{Throttled: true}, nil
	}
	message = truncate(strings.TrimSpace(message), 255)
	if message == "" {
		return PrivateMessage{}, ErrInvalidMessage
	}
	if service.options.ChatFilterEnabled && service.filter != nil {
		message, _ = service.filter.Censor(message)
	}
	if service.options.ChatLogEnabled && service.messageLog != nil {
		service.messageLog.Enqueue(fromID, toID, message)
	}
	_, online := service.live.Find(toID)
	return PrivateMessage{Message: message, Deliver: online}, nil
}

// UpdatePrivacy persists and projects messenger privacy fields.
func (service *Service) UpdatePrivacy(ctx context.Context, playerID int64, params playerservice.PrivacyParams) (playerservice.Record, error) {
	record, err := service.players.UpdatePrivacy(ctx, playerID, params)
	if err != nil {
		return playerservice.Record{}, err
	}
	if livePlayer, found := service.live.Find(playerID); found {
		_ = livePlayer.ReplaceSnapshot(playerlive.SnapshotFromRecord(record))
	}
	return record, nil
}

// SetRoomInvites persists Nitro's native room-invite blocking preference.
func (service *Service) SetRoomInvites(ctx context.Context, playerID int64, blocked bool) (playerservice.Record, error) {
	record, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return playerservice.Record{}, err
	}
	if !found {
		return playerservice.Record{}, ErrPlayerNotFound
	}
	return service.UpdatePrivacy(ctx, playerID, playerservice.PrivacyParams{
		BlockFriendRequests: record.Profile.BlockFriendRequests,
		BlockRoomInvites:    blocked,
		BlockFollowing:      record.Profile.BlockFollowing,
	})
}

// integerKey formats an id for Redis keys.
func integerKey(value int64) string { return strconv.FormatInt(value, 10) }

// truncate limits text by Unicode rune count.
func truncate(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}
