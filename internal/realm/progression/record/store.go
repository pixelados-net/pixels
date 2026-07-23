package record

import (
	"context"
	"time"
)

// Store persists progression catalogs and player state.
type Store interface {
	// WithinTransaction runs work in one shared PostgreSQL transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// Catalog loads one complete immutable catalog generation.
	Catalog(context.Context) (Catalog, error)
	// PlayerAchievements lists one player's progress joined to definitions.
	PlayerAchievements(context.Context, int64) ([]PlayerAchievement, error)
	// MutateProgress locks and increments one progress row atomically.
	MutateProgress(context.Context, int64, AchievementDefinition, int64, bool) (ProgressMutation, error)
	// SetProgress forces one progress row and returns crossed levels.
	SetProgress(context.Context, int64, AchievementDefinition, int64, int32, bool) (ProgressMutation, error)
	// ResetProgress deletes one player's progress row.
	ResetProgress(context.Context, int64, int64) (bool, error)
	// AddScore increments one player's durable achievement score.
	AddScore(context.Context, int64, int32) (int32, error)
	// PlayerTalents lists one player's paid talent levels.
	PlayerTalents(context.Context, int64) ([]PlayerTalent, error)
	// SetTalent advances one player's paid talent level.
	SetTalent(context.Context, int64, string, int32) (bool, error)
	// ForceTalent replaces one player's paid talent level exactly.
	ForceTalent(context.Context, int64, string, int32) error
	// ActiveQuest loads the single active quest.
	ActiveQuest(context.Context, int64) (PlayerQuestState, bool, error)
	// QuestProgress lists durable quest progress.
	QuestProgress(context.Context, int64) ([]PlayerQuestState, error)
	// ActivateQuest replaces one player's active quest.
	ActivateQuest(context.Context, int64, int64) (int64, error)
	// IncrementQuest locks and increments active quest progress up to its goal.
	IncrementQuest(context.Context, int64, int64, int64, int64) (PlayerQuestState, error)
	// CompleteQuest marks one active quest complete.
	CompleteQuest(context.Context, int64, int64) (bool, error)
	// CancelQuest clears one active quest.
	CancelQuest(context.Context, int64) (int64, error)
	// DailyQuestRejected reports whether the player discarded today's offer.
	DailyQuestRejected(context.Context, int64, time.Time) (bool, error)
	// RejectDailyQuest durably discards today's daily offer.
	RejectDailyQuest(context.Context, int64, time.Time) error
	// QuizPassed reports whether one player already passed a quiz.
	QuizPassed(context.Context, int64, string) (bool, error)
	// SaveQuizResult records one quiz attempt idempotently.
	SaveQuizResult(context.Context, int64, string, bool, []int32) (bool, error)
	// ClaimPromo claims one promo atomically under its global cap.
	ClaimPromo(context.Context, int64, PromoBadge, bool) (bool, error)
	// PromoClaimed reports whether one player already claimed a promotion.
	PromoClaimed(context.Context, int64, string) (bool, error)
	// InsertAudit appends one administrative mutation record.
	InsertAudit(context.Context, int64, string, string, string) error
}
