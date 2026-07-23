package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// PlayerTalents lists one player's paid talent levels.
func (repository *Repository) PlayerTalents(ctx context.Context, playerID int64) ([]progressionrecord.PlayerTalent, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select player_id,track,level from player_talent_levels where player_id=$1 order by track`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]progressionrecord.PlayerTalent, 0)
	for rows.Next() {
		var value progressionrecord.PlayerTalent
		if err = rows.Scan(&value.PlayerID, &value.Track, &value.Level); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// SetTalent advances one player's talent level monotonically.
func (repository *Repository) SetTalent(ctx context.Context, playerID int64, track string, level int32) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `insert into player_talent_levels(player_id,track,level) values($1,$2,$3) on conflict(player_id,track) do update set level=excluded.level,updated_at=now() where player_talent_levels.level<excluded.level`, playerID, track, level)
	return result.RowsAffected() > 0, err
}

// ForceTalent replaces one player's paid talent level exactly.
func (repository *Repository) ForceTalent(ctx context.Context, playerID int64, track string, level int32) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into player_talent_levels(player_id,track,level) values($1,$2,$3) on conflict(player_id,track) do update set level=excluded.level,updated_at=now()`, playerID, track, level)
	return err
}

// QuizPassed reports whether one player already passed a quiz.
func (repository *Repository) QuizPassed(ctx context.Context, playerID int64, code string) (bool, error) {
	var passed bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select passed from player_quiz_results where player_id=$1 and quiz_code=$2`, playerID, code).Scan(&passed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return passed, nil
}

// SaveQuizResult records one quiz attempt idempotently.
func (repository *Repository) SaveQuizResult(ctx context.Context, playerID int64, code string, passed bool, failed []int32) (bool, error) {
	if failed == nil {
		failed = []int32{}
	}
	result, err := repository.executorFor(ctx).Exec(ctx, `insert into player_quiz_results(player_id,quiz_code,passed,failed_question_refs,passed_at) values($1,$2,$3,$4,case when $3 then now() else null end) on conflict(player_id,quiz_code) do update set passed=player_quiz_results.passed or excluded.passed,failed_question_refs=excluded.failed_question_refs,attempted_at=now(),passed_at=case when player_quiz_results.passed_at is not null then player_quiz_results.passed_at when excluded.passed then now() else null end`, playerID, code, passed, failed)
	return result.RowsAffected() > 0, err
}

// ClaimPromo claims one promotional badge under a global cap.
func (repository *Repository) ClaimPromo(ctx context.Context, playerID int64, promo progressionrecord.PromoBadge, force bool) (bool, error) {
	executor := repository.executorFor(ctx)
	var code string
	if err := executor.QueryRow(ctx, `select code from promo_badges where code=$1 for update`, promo.Code).Scan(&code); err != nil {
		return false, err
	}
	result, err := executor.Exec(ctx, `insert into promo_badge_claims(player_id,code) select $1,$2 where $3 or $4=0 or (select count(*) from promo_badge_claims where code=$2)<$4 on conflict do nothing`, playerID, code, force, promo.MaxClaims)
	return result.RowsAffected() > 0, err
}

// PromoClaimed reports whether one player already claimed a promotion.
func (repository *Repository) PromoClaimed(ctx context.Context, playerID int64, code string) (bool, error) {
	var claimed bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from promo_badge_claims where player_id=$1 and code=$2)`, playerID, code).Scan(&claimed)
	return claimed, err
}

// InsertAudit appends one administrative mutation record.
func (repository *Repository) InsertAudit(ctx context.Context, actorID int64, action string, entity string, reason string) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into progression_audit(actor_player_id,action,entity,reason) values(nullif($1,0),$2,$3,$4)`, actorID, action, entity, reason)
	return err
}
