package admin

import (
	"context"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// CreatePromo inserts one promotional badge definition.
func (repository *Repository) CreatePromo(ctx context.Context, value progressionrecord.PromoBadge) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into promo_badges(code,badge_code,starts_at,ends_at,max_claims,enabled) values($1,$2,$3,$4,$5,$6)`, value.Code, value.BadgeCode, value.StartsAt, value.EndsAt, value.MaxClaims, value.Enabled)
	return err
}

// UpdatePromo replaces one promotional badge definition.
func (repository *Repository) UpdatePromo(ctx context.Context, value progressionrecord.PromoBadge) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update promo_badges set badge_code=$2,starts_at=$3,ends_at=$4,max_claims=$5,enabled=$6 where code=$1`, value.Code, value.BadgeCode, value.StartsAt, value.EndsAt, value.MaxClaims, value.Enabled)
	return result.RowsAffected() > 0, err
}

// DisablePromo soft-disables one promotional badge definition.
func (repository *Repository) DisablePromo(ctx context.Context, code string) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update promo_badges set enabled=false where code=$1 and enabled`, code)
	return result.RowsAffected() > 0, err
}

// PromoClaims lists durable claims for one promotion.
func (repository *Repository) PromoClaims(ctx context.Context, code string) ([]progressionrecord.PromoClaim, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select player_id,claimed_at from promo_badge_claims where code=$1 order by claimed_at,player_id`, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]progressionrecord.PromoClaim, 0)
	for rows.Next() {
		var value progressionrecord.PromoClaim
		if err = rows.Scan(&value.PlayerID, &value.ClaimedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
