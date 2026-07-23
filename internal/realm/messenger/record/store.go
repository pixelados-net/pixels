package record

import (
	"context"
)

// Store reads and mutates messenger persistence.
type Store interface {
	// CountFriends returns one player's directional friendship count.
	CountFriends(context.Context, int64) (int, error)
	// IsFriend reports whether a directional friendship exists.
	IsFriend(context.Context, int64, int64) (bool, error)
	// ListFriends returns directional friendship records.
	ListFriends(context.Context, int64) ([]Friendship, error)
	// ListFollowers returns viewer-specific rows that point at one friend.
	ListFollowers(context.Context, int64) ([]Friendship, error)
	// ListFriendIDs returns only one player's friend ids.
	ListFriendIDs(context.Context, int64) ([]int64, error)
	// HasRequestEither reports whether either request direction exists.
	HasRequestEither(context.Context, int64, int64) (bool, error)
	// HasRequest reports whether one exact request direction exists.
	HasRequest(context.Context, int64, int64) (bool, error)
	// CreateRequest inserts a pending request once.
	CreateRequest(context.Context, int64, int64) (bool, error)
	// ListIncomingRequests returns requests received by a player.
	ListIncomingRequests(context.Context, int64) ([]Request, error)
	// ListOutgoingRequests returns requests sent by a player.
	ListOutgoingRequests(context.Context, int64) ([]Request, error)
	// CountIncomingRequests returns requests received by a player.
	CountIncomingRequests(context.Context, int64) (int, error)
	// AcceptRequest atomically resolves a request and creates both friendship rows.
	AcceptRequest(context.Context, int64, int64) (bool, error)
	// DeclineRequests deletes requests received from supplied players or all requests.
	DeclineRequests(context.Context, int64, []int64, bool) (int64, error)
	// RemoveFriends atomically deletes both friendship directions.
	RemoveFriends(context.Context, int64, []int64) ([]int64, error)
	// SetRelation updates one directional friendship marker.
	SetRelation(context.Context, int64, int64, Relation) (bool, error)
	// RelationshipSummaries returns public non-empty relationship categories.
	RelationshipSummaries(context.Context, int64) ([]RelationshipSummary, error)
	// ListIgnored returns one player's directional ignore list.
	ListIgnored(context.Context, int64) ([]IgnoredPlayer, error)
	// AddIgnored persists one directional ignore once.
	AddIgnored(context.Context, int64, int64) (bool, error)
	// RemoveIgnored deletes one directional ignore.
	RemoveIgnored(context.Context, int64, int64) (bool, error)
	// SearchByPrefix returns active players matching a normalized username prefix.
	SearchByPrefix(context.Context, string, int) ([]SearchResult, error)
	// LogPrivateMessage persists one optional private-chat record.
	LogPrivateMessage(context.Context, int64, int64, string) error
}
