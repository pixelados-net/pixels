package record

import (
	"context"
	"time"
)

// QuizResult stores one player's latest durable quiz outcome.
type QuizResult struct {
	// PlayerID identifies the player.
	PlayerID int64
	// QuizCode identifies the quiz.
	QuizCode string
	// Passed reports whether any attempt succeeded.
	Passed bool
	// FailedQuestionRefs stores the latest failed question identifiers.
	FailedQuestionRefs []int32
	// AttemptedAt stores the latest attempt instant.
	AttemptedAt time.Time
	// PassedAt stores the first successful attempt instant.
	PassedAt *time.Time
}

// PromoClaim stores one durable promotional claim.
type PromoClaim struct {
	// PlayerID identifies the claimant.
	PlayerID int64
	// ClaimedAt stores the claim instant.
	ClaimedAt time.Time
}

// AdminStore persists progression catalog administration and support reads.
type AdminStore interface {
	// WithinTransaction runs work in one shared PostgreSQL transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// InsertAudit appends one administrative mutation record.
	InsertAudit(context.Context, int64, string, string, string) error
	// CreateAchievement inserts one achievement definition.
	CreateAchievement(context.Context, AchievementDefinition) (AchievementDefinition, error)
	// UpdateAchievement replaces mutable fields under optimistic locking.
	UpdateAchievement(context.Context, AchievementDefinition, int64) (AchievementDefinition, error)
	// DisableAchievement soft-disables one achievement definition.
	DisableAchievement(context.Context, int64) (bool, error)
	// UpsertAchievementLevel creates or replaces one cumulative level.
	UpsertAchievementLevel(context.Context, AchievementLevel) error
	// DeleteAchievementLevel removes only the current highest level.
	DeleteAchievementLevel(context.Context, int64, int32) (bool, error)
	// UpsertTalentLevel creates or replaces one talent level.
	UpsertTalentLevel(context.Context, TalentLevel) error
	// CreateCampaign inserts one quest campaign.
	CreateCampaign(context.Context, QuestCampaign) error
	// UpdateCampaign replaces one quest campaign.
	UpdateCampaign(context.Context, QuestCampaign) (bool, error)
	// DisableCampaign soft-disables one quest campaign.
	DisableCampaign(context.Context, string) (bool, error)
	// CreateQuest inserts one quest definition.
	CreateQuest(context.Context, QuestDefinition) (QuestDefinition, error)
	// UpdateQuest replaces mutable fields under optimistic locking.
	UpdateQuest(context.Context, QuestDefinition, int64) (QuestDefinition, error)
	// DisableQuest soft-disables one quest definition.
	DisableQuest(context.Context, int64) (bool, error)
	// CreateQuiz inserts one quiz definition.
	CreateQuiz(context.Context, Quiz) error
	// DisableQuiz soft-disables one quiz definition.
	DisableQuiz(context.Context, string) (bool, error)
	// CreateQuizQuestion inserts one quiz question.
	CreateQuizQuestion(context.Context, QuizQuestion) (QuizQuestion, error)
	// UpdateQuizQuestion replaces one quiz question.
	UpdateQuizQuestion(context.Context, QuizQuestion) (bool, error)
	// DeleteQuizQuestion removes one quiz question.
	DeleteQuizQuestion(context.Context, string, int64) (bool, error)
	// QuizResult reads one player's durable quiz outcome.
	QuizResult(context.Context, int64, string) (QuizResult, bool, error)
	// CreatePromo inserts one promotional badge definition.
	CreatePromo(context.Context, PromoBadge) error
	// UpdatePromo replaces one promotional badge definition.
	UpdatePromo(context.Context, PromoBadge) (bool, error)
	// DisablePromo soft-disables one promotional badge definition.
	DisablePromo(context.Context, string) (bool, error)
	// PromoClaims lists durable claims for one promotion.
	PromoClaims(context.Context, string) ([]PromoClaim, error)
}
