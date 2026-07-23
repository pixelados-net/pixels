package database

import (
	"context"
	"fmt"

	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// Prizes lists enabled recycler pool entries in rarity order.
func (repository *Repository) Prizes(ctx context.Context) ([]craftingrecord.Prize, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select p.tier,p.reward_definition_id,d.name,d.kind,d.sprite_id from crafting_recycler_prizes p join furniture_definitions d on d.id=p.reward_definition_id and d.deleted_at is null order by p.tier desc,p.reward_definition_id`)
	if err != nil {
		return nil, fmt.Errorf("list crafting recycler prizes: %w", err)
	}
	defer rows.Close()
	prizes := make([]craftingrecord.Prize, 0, 16)
	for rows.Next() {
		var prize craftingrecord.Prize
		if err = rows.Scan(&prize.Tier, &prize.RewardDefinitionID, &prize.RewardName, &prize.TypeCode, &prize.SpriteID); err != nil {
			return nil, err
		}
		prizes = append(prizes, prize)
	}
	return prizes, rows.Err()
}

// AddPrize inserts one recycler prize idempotently.
func (repository *Repository) AddPrize(ctx context.Context, tier int32, definitionID int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `insert into crafting_recycler_prizes(tier,reward_definition_id) values($1,$2) on conflict do nothing`, tier, definitionID)
	return tag.RowsAffected() > 0, err
}

// DeletePrize removes one recycler prize idempotently.
func (repository *Repository) DeletePrize(ctx context.Context, tier int32, definitionID int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `delete from crafting_recycler_prizes where tier=$1 and reward_definition_id=$2`, tier, definitionID)
	return tag.RowsAffected() > 0, err
}

// InsertAudit appends one administrative mutation record.
func (repository *Repository) InsertAudit(ctx context.Context, audit craftingrecord.Audit) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into crafting_audit(actor_player_id,action,entity_kind,entity_id,reason) values(nullif($1,0),$2,$3,nullif($4,0),$5)`, audit.ActorPlayerID, audit.Action, audit.EntityKind, audit.EntityID, audit.Reason)
	return err
}
