package database

import (
	"context"
	"fmt"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// CreateRequest inserts a pending request once.
func (repository *Repository) CreateRequest(ctx context.Context, fromID int64, toID int64) (bool, error) {
	tag, err := repository.executor.Exec(ctx, `insert into messenger_friend_requests (from_player_id,to_player_id) values ($1,$2) on conflict do nothing`, fromID, toID)
	if err != nil {
		return false, fmt.Errorf("create friend request %d to %d: %w", fromID, toID, err)
	}
	return tag.RowsAffected() == 1, nil
}

// AcceptRequest atomically resolves a request and creates both friendship rows.
func (repository *Repository) AcceptRequest(ctx context.Context, actorID int64, requesterID int64) (bool, error) {
	accepted := false
	err := repository.withinTx(ctx, func(txCtx context.Context, executor postgres.Executor) error {
		tag, err := executor.Exec(txCtx, `delete from messenger_friend_requests where (from_player_id=$2 and to_player_id=$1) or (from_player_id=$1 and to_player_id=$2)`, actorID, requesterID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		_, err = executor.Exec(txCtx, `insert into messenger_friendships (player_id,friend_player_id) values ($1,$2),($2,$1) on conflict do nothing`, actorID, requesterID)
		accepted = err == nil
		return err
	})
	if err != nil {
		return false, fmt.Errorf("accept friend request %d from %d: %w", actorID, requesterID, err)
	}
	return accepted, nil
}

// DeclineRequests deletes requests received from supplied players or all requests.
func (repository *Repository) DeclineRequests(ctx context.Context, actorID int64, requesterIDs []int64, all bool) (int64, error) {
	query := `delete from messenger_friend_requests where to_player_id=$1 and from_player_id=any($2::bigint[])`
	arguments := []any{actorID, requesterIDs}
	if all {
		query = `delete from messenger_friend_requests where to_player_id=$1`
		arguments = arguments[:1]
	}
	tag, err := repository.executor.Exec(ctx, query, arguments...)
	if err != nil {
		return 0, fmt.Errorf("decline player %d friend requests: %w", actorID, err)
	}
	return tag.RowsAffected(), nil
}

// RemoveFriends atomically deletes both friendship directions.
func (repository *Repository) RemoveFriends(ctx context.Context, actorID int64, friendIDs []int64) ([]int64, error) {
	removed := make([]int64, 0, len(friendIDs))
	err := repository.withinTx(ctx, func(txCtx context.Context, executor postgres.Executor) error {
		rows, err := executor.Query(txCtx, `delete from messenger_friendships where (player_id=$1 and friend_player_id=any($2::bigint[])) or (friend_player_id=$1 and player_id=any($2::bigint[])) returning case when player_id=$1 then friend_player_id else player_id end`, actorID, friendIDs)
		if err != nil {
			return err
		}
		defer rows.Close()
		seen := make(map[int64]struct{}, len(friendIDs))
		for rows.Next() {
			var id int64
			if err = rows.Scan(&id); err != nil {
				return err
			}
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				removed = append(removed, id)
			}
		}
		return rows.Err()
	})
	return removed, err
}

// SetRelation updates one directional friendship marker.
func (repository *Repository) SetRelation(ctx context.Context, playerID int64, friendID int64, relation messengermodel.Relation) (bool, error) {
	tag, err := repository.executor.Exec(ctx, `update messenger_friendships set relation=$3 where player_id=$1 and friend_player_id=$2`, playerID, friendID, relation)
	if err != nil {
		return false, fmt.Errorf("set messenger relation %d to %d: %w", playerID, friendID, err)
	}
	return tag.RowsAffected() == 1, nil
}

// AddIgnored persists one directional ignore once.
func (repository *Repository) AddIgnored(ctx context.Context, playerID int64, ignoredPlayerID int64) (bool, error) {
	tag, err := repository.executor.Exec(ctx, `insert into messenger_ignored_players (player_id,ignored_player_id) values ($1,$2) on conflict do nothing`, playerID, ignoredPlayerID)
	if err != nil {
		return false, fmt.Errorf("ignore player %d for %d: %w", ignoredPlayerID, playerID, err)
	}
	return tag.RowsAffected() == 1, nil
}

// RemoveIgnored deletes one directional ignore.
func (repository *Repository) RemoveIgnored(ctx context.Context, playerID int64, ignoredPlayerID int64) (bool, error) {
	tag, err := repository.executor.Exec(ctx, `delete from messenger_ignored_players where player_id=$1 and ignored_player_id=$2`, playerID, ignoredPlayerID)
	if err != nil {
		return false, fmt.Errorf("unignore player %d for %d: %w", ignoredPlayerID, playerID, err)
	}
	return tag.RowsAffected() == 1, nil
}

// LogPrivateMessage persists one optional private-chat record.
func (repository *Repository) LogPrivateMessage(ctx context.Context, fromID int64, toID int64, message string) error {
	_, err := repository.executor.Exec(ctx, `insert into messenger_private_messages (from_player_id,to_player_id,message) values ($1,$2,$3)`, fromID, toID, message)
	if err != nil {
		return fmt.Errorf("log private messenger message: %w", err)
	}
	return nil
}
