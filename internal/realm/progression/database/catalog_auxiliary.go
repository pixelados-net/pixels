package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// quests loads campaigns and definitions in two grouped queries.
func (repository *Repository) quests(ctx context.Context) ([]progressionrecord.QuestCampaign, []progressionrecord.QuestDefinition, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select code,seasonal,starts_at,ends_at,timing_code,enabled from quest_campaigns order by code`)
	if err != nil {
		return nil, nil, err
	}
	campaigns := make([]progressionrecord.QuestCampaign, 0)
	for rows.Next() {
		var campaign progressionrecord.QuestCampaign
		var starts, ends pgtype.Timestamptz
		if err = rows.Scan(&campaign.Code, &campaign.Seasonal, &starts, &ends, &campaign.TimingCode, &campaign.Enabled); err != nil {
			rows.Close()
			return nil, nil, err
		}
		campaign.StartsAt, campaign.EndsAt = timePointer(starts), timePointer(ends)
		campaigns = append(campaigns, campaign)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	rows, err = repository.executorFor(ctx).Query(ctx, `select id,campaign_code,series_number,name,localization_code,trigger_key,goal_amount,goal_data,reward_kind,reward_currency_type,reward_amount,reward_badge,coalesce(reward_definition_id,0),coalesce(reward_room_id,0),daily,easy,sort_order,enabled,version from quest_definitions order by campaign_code,series_number,id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	quests := make([]progressionrecord.QuestDefinition, 0)
	for rows.Next() {
		var quest progressionrecord.QuestDefinition
		if err = rows.Scan(&quest.ID, &quest.CampaignCode, &quest.SeriesNumber, &quest.Name, &quest.LocalizationCode, &quest.TriggerKey, &quest.GoalAmount, &quest.GoalData, &quest.RewardKind, &quest.RewardCurrencyType, &quest.RewardAmount, &quest.RewardBadge, &quest.RewardDefinitionID, &quest.RewardRoomID, &quest.Daily, &quest.Easy, &quest.SortOrder, &quest.Enabled, &quest.Version); err != nil {
			return nil, nil, err
		}
		quests = append(quests, quest)
	}
	return campaigns, quests, rows.Err()
}

// quizzes loads quiz questions grouped by quiz code.
func (repository *Repository) quizzes(ctx context.Context) ([]progressionrecord.Quiz, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select q.code,q.kind,q.enabled,qq.id,qq.question_ref,qq.correct_answer_id from quizzes q left join quiz_questions qq on qq.quiz_code=q.code order by q.code,qq.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	quizzes := make([]progressionrecord.Quiz, 0)
	var current *progressionrecord.Quiz
	for rows.Next() {
		var quiz progressionrecord.Quiz
		var id pgtype.Int8
		var reference, answer pgtype.Int4
		if err = rows.Scan(&quiz.Code, &quiz.Kind, &quiz.Enabled, &id, &reference, &answer); err != nil {
			return nil, err
		}
		if current == nil || current.Code != quiz.Code {
			quizzes = append(quizzes, quiz)
			current = &quizzes[len(quizzes)-1]
		}
		if id.Valid {
			current.Questions = append(current.Questions, progressionrecord.QuizQuestion{ID: id.Int64, QuizCode: quiz.Code, QuestionRef: reference.Int32, CorrectAnswerID: answer.Int32})
		}
	}
	return quizzes, rows.Err()
}

// promos loads promotional badge definitions.
func (repository *Repository) promos(ctx context.Context) ([]progressionrecord.PromoBadge, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select code,badge_code,starts_at,ends_at,max_claims,enabled from promo_badges order by code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	promos := make([]progressionrecord.PromoBadge, 0)
	for rows.Next() {
		var promo progressionrecord.PromoBadge
		var starts, ends pgtype.Timestamptz
		if err = rows.Scan(&promo.Code, &promo.BadgeCode, &starts, &ends, &promo.MaxClaims, &promo.Enabled); err != nil {
			return nil, err
		}
		promo.StartsAt, promo.EndsAt = timePointer(starts), timePointer(ends)
		promos = append(promos, promo)
	}
	return promos, rows.Err()
}

// timePointer converts one nullable PostgreSQL timestamp.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
