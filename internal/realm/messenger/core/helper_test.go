package core

import (
	"context"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
)

// fakeStore supplies in-memory messenger persistence for tests.
type fakeStore struct {
	// friends stores directional friendship existence.
	friends map[[2]int64]bool
	// relations stores directional relationship markers.
	relations map[[2]int64]messengermodel.Relation
	// friendships stores follower projections by target player.
	friendships map[int64][]messengermodel.Friendship
	// requests stores pending requests.
	requests []messengermodel.Request
	// records stores unused player records for benchmark setup.
	records map[int64]any
	// searches counts prefix queries.
	searches int
	// ignored stores directional ignored users.
	ignored map[[2]int64]bool
}

// newFakeStore creates empty messenger persistence.
func newFakeStore() *fakeStore {
	return &fakeStore{friends: make(map[[2]int64]bool), relations: make(map[[2]int64]messengermodel.Relation), friendships: make(map[int64][]messengermodel.Friendship), records: make(map[int64]any), ignored: make(map[[2]int64]bool)}
}

// CountFriends returns directional friendship count.
func (store *fakeStore) CountFriends(_ context.Context, playerID int64) (int, error) {
	count := 0
	for pair, found := range store.friends {
		if found && pair[0] == playerID {
			count++
		}
	}
	return count, nil
}

// IsFriend reports directional friendship existence.
func (store *fakeStore) IsFriend(_ context.Context, playerID int64, friendID int64) (bool, error) {
	return store.friends[[2]int64{playerID, friendID}], nil
}

// ListFriends returns directional friendship records.
func (store *fakeStore) ListFriends(_ context.Context, playerID int64) ([]messengermodel.Friendship, error) {
	items := make([]messengermodel.Friendship, 0)
	for pair, found := range store.friends {
		if found && pair[0] == playerID {
			items = append(items, messengermodel.Friendship{PlayerID: playerID, FriendPlayerID: pair[1], Relation: store.relations[pair]})
		}
	}
	return items, nil
}

// ListFollowers returns viewer-specific rows pointing at one player.
func (store *fakeStore) ListFollowers(_ context.Context, playerID int64) ([]messengermodel.Friendship, error) {
	return store.friendships[playerID], nil
}

// ListFriendIDs returns directional friend ids.
func (store *fakeStore) ListFriendIDs(ctx context.Context, playerID int64) ([]int64, error) {
	items, _ := store.ListFriends(ctx, playerID)
	ids := make([]int64, len(items))
	for index, item := range items {
		ids[index] = item.FriendPlayerID
	}
	return ids, nil
}

// HasRequestEither reports pending requests in either direction.
func (store *fakeStore) HasRequestEither(_ context.Context, firstID int64, secondID int64) (bool, error) {
	for _, request := range store.requests {
		if request.FromPlayerID == firstID && request.ToPlayerID == secondID || request.FromPlayerID == secondID && request.ToPlayerID == firstID {
			return true, nil
		}
	}
	return false, nil
}

// HasRequest reports pending requests in one direction.
func (store *fakeStore) HasRequest(_ context.Context, fromID int64, toID int64) (bool, error) {
	for _, request := range store.requests {
		if request.FromPlayerID == fromID && request.ToPlayerID == toID {
			return true, nil
		}
	}
	return false, nil
}

// CreateRequest inserts one pending request.
func (store *fakeStore) CreateRequest(_ context.Context, fromID int64, toID int64) (bool, error) {
	store.requests = append(store.requests, messengermodel.Request{FromPlayerID: fromID, ToPlayerID: toID})
	return true, nil
}

// ListIncomingRequests returns received requests.
func (store *fakeStore) ListIncomingRequests(_ context.Context, playerID int64) ([]messengermodel.Request, error) {
	return filterRequests(store.requests, playerID, true), nil
}

// ListOutgoingRequests returns sent requests.
func (store *fakeStore) ListOutgoingRequests(_ context.Context, playerID int64) ([]messengermodel.Request, error) {
	return filterRequests(store.requests, playerID, false), nil
}

// CountIncomingRequests counts received requests.
func (store *fakeStore) CountIncomingRequests(ctx context.Context, playerID int64) (int, error) {
	items, _ := store.ListIncomingRequests(ctx, playerID)
	return len(items), nil
}

// AcceptRequest resolves a request and creates both friendship directions.
func (store *fakeStore) AcceptRequest(_ context.Context, actorID int64, requesterID int64) (bool, error) {
	for index, request := range store.requests {
		if request.FromPlayerID == requesterID && request.ToPlayerID == actorID {
			store.requests = append(store.requests[:index], store.requests[index+1:]...)
			store.friends[[2]int64{actorID, requesterID}] = true
			store.friends[[2]int64{requesterID, actorID}] = true
			return true, nil
		}
	}
	return false, nil
}

// DeclineRequests removes selected or all received requests.
func (store *fakeStore) DeclineRequests(_ context.Context, actorID int64, requesterIDs []int64, all bool) (int64, error) {
	remaining := store.requests[:0]
	var deleted int64
	for _, request := range store.requests {
		remove := request.ToPlayerID == actorID && (all || containsID(requesterIDs, request.FromPlayerID))
		if remove {
			deleted++
		} else {
			remaining = append(remaining, request)
		}
	}
	store.requests = remaining
	return deleted, nil
}

// RemoveFriends removes both directions for supplied friends.
func (store *fakeStore) RemoveFriends(_ context.Context, actorID int64, friendIDs []int64) ([]int64, error) {
	removed := make([]int64, 0, len(friendIDs))
	for _, friendID := range friendIDs {
		if store.friends[[2]int64{actorID, friendID}] {
			removed = append(removed, friendID)
		}
		delete(store.friends, [2]int64{actorID, friendID})
		delete(store.friends, [2]int64{friendID, actorID})
	}
	return removed, nil
}

// SetRelation updates one directional marker.
func (store *fakeStore) SetRelation(_ context.Context, playerID int64, friendID int64, relation messengermodel.Relation) (bool, error) {
	pair := [2]int64{playerID, friendID}
	if !store.friends[pair] {
		return false, nil
	}
	store.relations[pair] = relation
	return true, nil
}

// RelationshipSummaries returns no public summaries.
func (*fakeStore) RelationshipSummaries(context.Context, int64) ([]messengermodel.RelationshipSummary, error) {
	return nil, nil
}

// ListIgnored returns directional ignored users.
func (store *fakeStore) ListIgnored(_ context.Context, playerID int64) ([]messengermodel.IgnoredPlayer, error) {
	items := make([]messengermodel.IgnoredPlayer, 0)
	for pair, ignored := range store.ignored {
		if ignored && pair[0] == playerID {
			items = append(items, messengermodel.IgnoredPlayer{PlayerID: pair[1]})
		}
	}
	return items, nil
}

// AddIgnored persists one directional ignore.
func (store *fakeStore) AddIgnored(_ context.Context, playerID int64, targetID int64) (bool, error) {
	pair := [2]int64{playerID, targetID}
	added := !store.ignored[pair]
	store.ignored[pair] = true
	return added, nil
}

// RemoveIgnored deletes one directional ignore.
func (store *fakeStore) RemoveIgnored(_ context.Context, playerID int64, targetID int64) (bool, error) {
	pair := [2]int64{playerID, targetID}
	removed := store.ignored[pair]
	delete(store.ignored, pair)
	return removed, nil
}

// SearchByPrefix returns one deterministic match.
func (store *fakeStore) SearchByPrefix(context.Context, string, int) ([]messengermodel.SearchResult, error) {
	store.searches++
	return []messengermodel.SearchResult{{PlayerID: 2, Username: "alice", Look: "hd-180-1"}}, nil
}

// LogPrivateMessage accepts optional logging.
func (*fakeStore) LogPrivateMessage(context.Context, int64, int64, string) error { return nil }

// filterRequests selects one request direction.
func filterRequests(items []messengermodel.Request, playerID int64, incoming bool) []messengermodel.Request {
	result := make([]messengermodel.Request, 0)
	for _, item := range items {
		if incoming && item.ToPlayerID == playerID || !incoming && item.FromPlayerID == playerID {
			result = append(result, item)
		}
	}
	return result
}

// containsID reports whether a player id is present.
func containsID(items []int64, value int64) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
