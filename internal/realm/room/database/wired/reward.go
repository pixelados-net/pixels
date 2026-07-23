package wired

import (
	"context"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// ListRewards lists normalized rewards for one WIRED effect.
func (repository *Repository) ListRewards(ctx context.Context, itemID int64) ([]roomwired.Reward, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select id,ordinal,kind,reference,amount,weight,stock from room_wired_rewards where wired_item_id=$1 order by ordinal`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rewards := make([]roomwired.Reward, 0)
	for rows.Next() {
		var reward roomwired.Reward
		if err = rows.Scan(&reward.ID, &reward.Ordinal, &reward.Kind, &reward.Reference, &reward.Amount, &reward.Weight, &reward.Stock); err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}
	return rewards, rows.Err()
}

// ReplaceRewards atomically replaces normalized reward definitions.
func (repository *Repository) ReplaceRewards(ctx context.Context, itemID int64, rewards []roomwired.Reward) error {
	work := func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		if _, err := executor.Exec(txCtx, `delete from room_wired_rewards where wired_item_id=$1`, itemID); err != nil {
			return err
		}
		for index, reward := range rewards {
			_, err := executor.Exec(txCtx, `insert into room_wired_rewards(wired_item_id,ordinal,kind,reference,amount,weight,stock) values($1,$2,$3,$4,$5,$6,$7)`, itemID, index, reward.Kind, reward.Reference, reward.Amount, reward.Weight, reward.Stock)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return repository.within(ctx, work)
}

// ClaimReward atomically selects, delivers, and records one period claim.
func (repository *Repository) ClaimReward(ctx context.Context, itemID int64, playerID int64, periodKey string, unique bool, seed uint64, deliver func(context.Context, roomwired.Reward) error) (roomwired.ClaimStatus, roomwired.Reward, error) {
	status := roomwired.ClaimUnavailable
	var selected roomwired.Reward
	err := repository.within(ctx, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		var globalLimit int64
		if err := executor.QueryRow(txCtx, `select coalesce((int_params->>2)::bigint,0) from room_wired_settings where item_id=$1 for update`, itemID).Scan(&globalLimit); err != nil {
			return err
		}
		var exists bool
		if err := executor.QueryRow(txCtx, `select exists(select 1 from room_wired_reward_claims where wired_item_id=$1 and player_id=$2 and period_key=$3)`, itemID, playerID, periodKey).Scan(&exists); err != nil {
			return err
		}
		if exists {
			status = roomwired.ClaimAlreadyReceived
			return nil
		}
		if globalLimit > 0 {
			var claims int64
			if err := executor.QueryRow(txCtx, `select count(*) from room_wired_reward_claims where wired_item_id=$1`, itemID).Scan(&claims); err != nil {
				return err
			}
			if claims >= globalLimit {
				status = roomwired.ClaimOutOfStock
				return nil
			}
		}
		rewards, err := repository.lockRewards(txCtx, executor, itemID)
		if err != nil || len(rewards) == 0 {
			return err
		}
		excluded := make(map[int64]struct{})
		if unique {
			rows, err := executor.Query(txCtx, `select reward_id from room_wired_reward_claims where wired_item_id=$1 and player_id=$2 and reward_id is not null`, itemID, playerID)
			if err != nil {
				return err
			}
			for rows.Next() {
				var rewardID int64
				if err = rows.Scan(&rewardID); err != nil {
					rows.Close()
					return err
				}
				excluded[rewardID] = struct{}{}
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return err
			}
		}
		selected, status = chooseReward(rewards, excluded, seed)
		if status == roomwired.ClaimMissed {
			_, err = executor.Exec(txCtx, `insert into room_wired_reward_claims(wired_item_id,player_id,reward_id,period_key) values($1,$2,null,$3)`, itemID, playerID, periodKey)
			return err
		}
		if status != roomwired.ClaimDelivered {
			return nil
		}
		if selected.Stock != nil {
			result, err := executor.Exec(txCtx, `update room_wired_rewards set stock=stock-1 where id=$1 and stock>0`, selected.ID)
			if err != nil {
				return err
			}
			if result.RowsAffected() == 0 {
				status = roomwired.ClaimOutOfStock
				return nil
			}
		}
		if err := deliver(txCtx, selected); err != nil {
			return err
		}
		_, err = executor.Exec(txCtx, `insert into room_wired_reward_claims(wired_item_id,player_id,reward_id,period_key) values($1,$2,$3,$4)`, itemID, playerID, selected.ID, periodKey)
		return err
	})
	return status, selected, err
}

// within reuses an active transaction or creates one.
func (repository *Repository) within(ctx context.Context, work func(context.Context) error) error {
	if _, active := postgres.ScopedExecutor(ctx); active {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// lockRewards loads candidate rewards under row locks.
func (repository *Repository) lockRewards(ctx context.Context, executor postgres.Executor, itemID int64) ([]roomwired.Reward, error) {
	rows, err := executor.Query(ctx, `select id,ordinal,kind,reference,amount,weight,stock from room_wired_rewards where wired_item_id=$1 order by ordinal for update`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rewards := make([]roomwired.Reward, 0)
	for rows.Next() {
		var reward roomwired.Reward
		if err = rows.Scan(&reward.ID, &reward.Ordinal, &reward.Kind, &reward.Reference, &reward.Amount, &reward.Weight, &reward.Stock); err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}
	return rewards, rows.Err()
}

// chooseReward selects one in-stock reward using integer weights.
func chooseReward(rewards []roomwired.Reward, excluded map[int64]struct{}, seed uint64) (roomwired.Reward, roomwired.ClaimStatus) {
	var total uint64
	for _, reward := range rewards {
		if _, found := excluded[reward.ID]; found {
			continue
		}
		if reward.Weight > 0 && (reward.Stock == nil || *reward.Stock > 0) {
			total += uint64(reward.Weight)
		}
	}
	if total == 0 {
		return roomwired.Reward{}, roomwired.ClaimOutOfStock
	}
	if total < 100 && seed%100 >= total {
		return roomwired.Reward{}, roomwired.ClaimMissed
	}
	value := seed % total
	for _, reward := range rewards {
		if _, found := excluded[reward.ID]; found {
			continue
		}
		if reward.Weight <= 0 || reward.Stock != nil && *reward.Stock <= 0 {
			continue
		}
		if value < uint64(reward.Weight) {
			return reward, roomwired.ClaimDelivered
		}
		value -= uint64(reward.Weight)
	}
	return roomwired.Reward{}, roomwired.ClaimUnavailable
}
