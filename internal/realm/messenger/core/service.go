// Package core coordinates shared messenger domain behavior.
package core

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/pkg/redis"
)

// Search returns shared cached prefix results under a per-player throttle.
func (service *Service) Search(ctx context.Context, playerID int64, term string) (SearchResult, error) {
	term = normalizeSearch(term)
	if term == "" {
		return SearchResult{Results: []messengermodel.SearchResult{}}, nil
	}
	allowed, err := service.redis.SetIfAbsent(ctx, searchThrottleKey(playerID), []byte{1}, service.options.SearchThrottle)
	if err != nil {
		return SearchResult{}, err
	}
	if !allowed {
		return SearchResult{Throttled: true}, nil
	}
	key := "messenger:search:result:" + term
	if cached, found, cacheErr := service.redis.Find(ctx, key); cacheErr == nil && found {
		var results []messengermodel.SearchResult
		if json.Unmarshal(cached, &results) == nil {
			service.projectSearch(results)
			return SearchResult{Results: results}, nil
		}
	}
	results, err := service.store.SearchByPrefix(ctx, term, service.options.MaxSearchResults)
	if err != nil {
		return SearchResult{}, err
	}
	if encoded, encodeErr := json.Marshal(results); encodeErr == nil {
		_ = service.redis.Set(ctx, key, encoded, service.options.SearchCacheTTL)
	}
	service.projectSearch(results)
	return SearchResult{Results: results}, nil
}

// projectSearch overlays process-local presence without polluting shared cache entries.
func (service *Service) projectSearch(results []messengermodel.SearchResult) {
	for index := range results {
		player, online := service.live.Find(results[index].PlayerID)
		results[index].Online = online
		if online && !results[index].BlockFollowing {
			_, results[index].CanFollow = player.CurrentRoom()
		}
	}
}

// normalizeSearch lowercases, removes spaces, and truncates a search term to fifteen runes.
func normalizeSearch(term string) string {
	runes := make([]rune, 0, 15)
	for _, value := range strings.ToLower(term) {
		if unicode.IsSpace(value) {
			continue
		}
		runes = append(runes, value)
		if len(runes) == 15 {
			break
		}
	}
	return string(runes)
}

// searchThrottleKey returns one player search throttle key.
func searchThrottleKey(playerID int64) string {
	return "messenger:search:throttle:" + integerKey(playerID)
}

// Options stores normalized messenger behavior settings.
type Options struct {
	// MaxFriends stores normal friend capacity.
	MaxFriends int
	// MaxFriendsClub stores active club friend capacity.
	MaxFriendsClub int
	// MaxSearchResults stores the search result cap.
	MaxSearchResults int
	// SearchCacheTTL stores shared result cache lifetime.
	SearchCacheTTL time.Duration
	// FriendCacheTTL stores shared durable friend-card cache lifetime.
	FriendCacheTTL time.Duration
	// SearchThrottle stores the per-player search interval.
	SearchThrottle time.Duration
	// ChatThrottle stores the per-sender private-chat interval.
	ChatThrottle time.Duration
	// ChatFilterEnabled reports whether private messages use the global filter.
	ChatFilterEnabled bool
	// ChatLogEnabled reports whether private messages are persisted.
	ChatLogEnabled bool
}

// Nodes stores messenger permission capabilities.
type Nodes struct {
	// FriendsUnlimited bypasses friend-list capacity.
	FriendsUnlimited permission.Node
	// FollowAny bypasses follow privacy.
	FollowAny permission.Node
}

// Service coordinates messenger persistence and live state.
type Service struct {
	// options stores normalized behavior limits.
	options Options
	// store persists messenger state.
	store messengermodel.Store
	// players reads durable player records.
	players playerservice.Manager
	// live reads online and room state.
	live *playerlive.Registry
	// rooms reads active populated rooms.
	rooms *roomlive.Registry
	// permissions resolves staff bypasses.
	permissions permissionservice.Checker
	// redis stores shared caches and throttles.
	redis *redis.Client
	// filter applies the global word dictionary.
	filter *chatfilter.Service
	// unlimitedNode stores the friend-capacity bypass.
	unlimitedNode permission.Node
	// followNode stores the follow-privacy bypass.
	followNode permission.Node
	// messageLog queues optional private-message persistence.
	messageLog MessageLogger
}

// MessageLogger queues optional private-message persistence.
type MessageLogger interface {
	// Enqueue queues one accepted private message without blocking delivery.
	Enqueue(fromID int64, toID int64, message string) bool
}

// New creates messenger behavior.
func New(options Options, store messengermodel.Store, players playerservice.Manager, live *playerlive.Registry, rooms *roomlive.Registry, permissions permissionservice.Checker, redisClient *redis.Client, filter *chatfilter.Service, nodes Nodes, messageLog MessageLogger) *Service {
	return &Service{options: options, store: store, players: players, live: live, rooms: rooms, permissions: permissions, redis: redisClient, filter: filter, unlimitedNode: nodes.FriendsUnlimited, followNode: nodes.FollowAny, messageLog: messageLog}
}

// RequestResult describes one friend-request attempt.
type RequestResult struct {
	// Sent reports whether a new request was persisted.
	Sent bool
	// Target stores the resolved target when available.
	Target playerservice.Record
}

// AcceptResult describes one accepted friend request.
type AcceptResult struct {
	// Accepted reports whether a pending request was resolved.
	Accepted bool
	// ActorCard stores the requester as viewed by the accepting player.
	ActorCard messengermodel.Card
	// RequesterCard stores the accepting player as viewed by the requester.
	RequesterCard messengermodel.Card
}

// SearchResult describes one search response and throttle state.
type SearchResult struct {
	// Results stores matching public players.
	Results []messengermodel.SearchResult
	// Throttled reports a silently ignored request.
	Throttled bool
}

var (
	// ErrInvalidUsername reports malformed friend-search or request input.
	ErrInvalidUsername = errors.New("invalid messenger username")
	// ErrPlayerNotFound reports a missing target player.
	ErrPlayerNotFound = errors.New("messenger player not found")
	// ErrRequestsBlocked reports target friend-request privacy.
	ErrRequestsBlocked = errors.New("target does not accept friend requests")
	// ErrOwnListFull reports requester friend-list capacity.
	ErrOwnListFull = errors.New("own messenger friend list is full")
	// ErrTargetListFull reports target friend-list capacity.
	ErrTargetListFull = errors.New("target messenger friend list is full")
	// ErrNotFriend reports behavior requiring an existing friendship.
	ErrNotFriend = errors.New("players are not messenger friends")
	// ErrInvalidRelation reports a relation outside Nitro's supported range.
	ErrInvalidRelation = errors.New("invalid messenger relationship")
	// ErrInvalidBatch reports an empty or oversized player batch.
	ErrInvalidBatch = errors.New("invalid messenger player batch")
	// ErrNoActiveRoom reports an invitation without room presence.
	ErrNoActiveRoom = errors.New("messenger inviter has no active room")
	// ErrFriendOffline reports following an offline friend.
	ErrFriendOffline = errors.New("messenger friend is offline")
	// ErrFollowingBlocked reports target follow privacy.
	ErrFollowingBlocked = errors.New("messenger friend blocks following")
	// ErrFriendNotInRoom reports following a friend outside rooms.
	ErrFriendNotInRoom = errors.New("messenger friend is not in a room")
	// ErrNoPopulatedRoom reports a find-new-friends miss.
	ErrNoPopulatedRoom = errors.New("no populated messenger room available")
	// ErrInvalidMessage reports empty private communication.
	ErrInvalidMessage = errors.New("invalid messenger message")
)
