package database

import (
	"context"
	"fmt"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
)

// CountFriends returns one player's directional friendship count.
func (repository *Repository) CountFriends(ctx context.Context, playerID int64) (count int, err error) {
	err = repository.executor.QueryRow(ctx, `select count(*) from messenger_friendships where player_id=$1`, playerID).Scan(&count)
	if err != nil {
		err = fmt.Errorf("count player %d messenger friends: %w", playerID, err)
	}
	return
}

// IsFriend reports whether a directional friendship exists.
func (repository *Repository) IsFriend(ctx context.Context, playerID int64, friendID int64) (found bool, err error) {
	err = repository.executor.QueryRow(ctx, `select exists(select 1 from messenger_friendships where player_id=$1 and friend_player_id=$2)`, playerID, friendID).Scan(&found)
	if err != nil {
		err = fmt.Errorf("check messenger friendship %d to %d: %w", playerID, friendID, err)
	}
	return
}

// ListFriends returns directional friendship records.
func (repository *Repository) ListFriends(ctx context.Context, playerID int64) ([]messengermodel.Friendship, error) {
	rows, err := repository.executor.Query(ctx, `select player_id, friend_player_id, relation, category_id, created_at from messenger_friendships where player_id=$1 order by friend_player_id`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d messenger friends: %w", playerID, err)
	}
	defer rows.Close()
	items := make([]messengermodel.Friendship, 0)
	for rows.Next() {
		var item messengermodel.Friendship
		if err = rows.Scan(&item.PlayerID, &item.FriendPlayerID, &item.Relation, &item.CategoryID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan messenger friendship: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListFollowers returns viewer-specific rows that point at one friend.
func (repository *Repository) ListFollowers(ctx context.Context, friendID int64) ([]messengermodel.Friendship, error) {
	rows, err := repository.executor.Query(ctx, `select player_id, friend_player_id, relation, category_id, created_at from messenger_friendships where friend_player_id=$1 order by player_id`, friendID)
	if err != nil {
		return nil, fmt.Errorf("list messenger followers for player %d: %w", friendID, err)
	}
	defer rows.Close()
	items := make([]messengermodel.Friendship, 0)
	for rows.Next() {
		var item messengermodel.Friendship
		if err = rows.Scan(&item.PlayerID, &item.FriendPlayerID, &item.Relation, &item.CategoryID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan messenger follower: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListFriendIDs returns only one player's friend ids.
func (repository *Repository) ListFriendIDs(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := repository.executor.Query(ctx, `select friend_player_id from messenger_friendships where player_id=$1 order by friend_player_id`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d messenger friend ids: %w", playerID, err)
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// RelationshipSummaries returns public non-empty relationship categories.
func (repository *Repository) RelationshipSummaries(ctx context.Context, playerID int64) ([]messengermodel.RelationshipSummary, error) {
	rows, err := repository.executor.Query(ctx, `select f.relation,(count(*) over (partition by f.relation))::int,p.id,p.username,pp.look from messenger_friendships f join players p on p.id=f.friend_player_id join player_profiles pp on pp.player_id=p.id where f.player_id=$1 and f.relation between 1 and 3 order by f.relation,p.id`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d relationship summaries: %w", playerID, err)
	}
	defer rows.Close()
	items := make([]messengermodel.RelationshipSummary, 0, 3)
	var previous messengermodel.Relation
	for rows.Next() {
		var item messengermodel.RelationshipSummary
		if err = rows.Scan(&item.Relation, &item.Count, &item.Sample.PlayerID, &item.Sample.Username, &item.SampleLook); err != nil {
			return nil, fmt.Errorf("scan messenger relationship summary: %w", err)
		}
		if item.Relation != previous {
			items = append(items, item)
			previous = item.Relation
		}
	}
	return items, rows.Err()
}

// ListIgnored returns one player's directional ignore list.
func (repository *Repository) ListIgnored(ctx context.Context, playerID int64) ([]messengermodel.IgnoredPlayer, error) {
	rows, err := repository.executor.Query(ctx, `select p.id,p.username from messenger_ignored_players i join players p on p.id=i.ignored_player_id where i.player_id=$1 and p.deleted_at is null order by lower(p.username)`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d ignored users: %w", playerID, err)
	}
	defer rows.Close()
	items := make([]messengermodel.IgnoredPlayer, 0)
	for rows.Next() {
		var item messengermodel.IgnoredPlayer
		if err = rows.Scan(&item.PlayerID, &item.Username); err != nil {
			return nil, fmt.Errorf("scan ignored player: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// HasRequestEither reports whether either request direction exists.
func (repository *Repository) HasRequestEither(ctx context.Context, firstID int64, secondID int64) (found bool, err error) {
	err = repository.executor.QueryRow(ctx, `select exists(select 1 from messenger_friend_requests where (from_player_id=$1 and to_player_id=$2) or (from_player_id=$2 and to_player_id=$1))`, firstID, secondID).Scan(&found)
	return
}

// HasRequest reports whether one exact request direction exists.
func (repository *Repository) HasRequest(ctx context.Context, fromID int64, toID int64) (found bool, err error) {
	err = repository.executor.QueryRow(ctx, `select exists(select 1 from messenger_friend_requests where from_player_id=$1 and to_player_id=$2)`, fromID, toID).Scan(&found)
	return
}

// ListIncomingRequests returns requests received by a player.
func (repository *Repository) ListIncomingRequests(ctx context.Context, playerID int64) ([]messengermodel.Request, error) {
	return repository.listRequests(ctx, `select from_player_id, to_player_id, created_at from messenger_friend_requests where to_player_id=$1 order by created_at desc`, playerID)
}

// ListOutgoingRequests returns requests sent by a player.
func (repository *Repository) ListOutgoingRequests(ctx context.Context, playerID int64) ([]messengermodel.Request, error) {
	return repository.listRequests(ctx, `select from_player_id, to_player_id, created_at from messenger_friend_requests where from_player_id=$1 order by created_at desc`, playerID)
}

// listRequests reads requests using one directional query.
func (repository *Repository) listRequests(ctx context.Context, query string, playerID int64) ([]messengermodel.Request, error) {
	rows, err := repository.executor.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d friend requests: %w", playerID, err)
	}
	defer rows.Close()
	requests := make([]messengermodel.Request, 0)
	for rows.Next() {
		var item messengermodel.Request
		if err = rows.Scan(&item.FromPlayerID, &item.ToPlayerID, &item.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, item)
	}
	return requests, rows.Err()
}

// CountIncomingRequests returns requests received by a player.
func (repository *Repository) CountIncomingRequests(ctx context.Context, playerID int64) (count int, err error) {
	err = repository.executor.QueryRow(ctx, `select count(*) from messenger_friend_requests where to_player_id=$1`, playerID).Scan(&count)
	return
}

// SearchByPrefix returns active players matching a normalized username prefix.
func (repository *Repository) SearchByPrefix(ctx context.Context, prefix string, limit int) ([]messengermodel.SearchResult, error) {
	rows, err := repository.executor.Query(ctx, `select p.id, p.username, pp.motto, pp.look, case when pp.gender='M' then 0 else 1 end, pp.block_following from players p join player_profiles pp on pp.player_id=p.id where p.deleted_at is null and lower(p.username) like $1 || '%' order by lower(p.username) limit $2`, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("search messenger players by prefix %q: %w", prefix, err)
	}
	defer rows.Close()
	results := make([]messengermodel.SearchResult, 0, limit)
	for rows.Next() {
		var item messengermodel.SearchResult
		if err = rows.Scan(&item.PlayerID, &item.Username, &item.Motto, &item.Look, &item.Gender, &item.BlockFollowing); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}
