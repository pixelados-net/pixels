package record

import "time"

// TalentRequirement describes one achievement prerequisite.
type TalentRequirement struct {
	// DefinitionID identifies the required achievement.
	DefinitionID int64 `json:"definitionId"`
	// RequiredLevel stores the minimum achievement level.
	RequiredLevel int32 `json:"requiredLevel"`
}

// TalentLevel describes one derived talent track level.
type TalentLevel struct {
	// Track identifies the track.
	Track string
	// Level stores the one-based level number.
	Level int32
	// Requirements stores achievement prerequisites.
	Requirements []TalentRequirement
	// RewardItems stores furniture definition identifiers.
	RewardItems []int64
	// RewardPerks stores permission perk names.
	RewardPerks []string
	// RewardBadges stores badge codes.
	RewardBadges []string
}

// PlayerTalent stores one player's paid track level.
type PlayerTalent struct {
	// PlayerID identifies the player.
	PlayerID int64
	// Track identifies the talent track.
	Track string
	// Level stores the highest paid level.
	Level int32
}

// Quiz describes one enabled quiz catalog entry.
type Quiz struct {
	// Code identifies the quiz.
	Code string
	// Kind selects safety or poll behavior.
	Kind string
	// Enabled controls availability.
	Enabled bool
	// Questions stores ordered questions.
	Questions []QuizQuestion
}

// QuizQuestion stores one client-known question and answer identifier.
type QuizQuestion struct {
	// ID identifies the durable question.
	ID int64
	// QuizCode identifies the parent quiz.
	QuizCode string
	// QuestionRef identifies client localization content.
	QuestionRef int32
	// CorrectAnswerID stores the expected answer.
	CorrectAnswerID int32
}

// PromoBadge describes one time-windowed badge claim.
type PromoBadge struct {
	// Code identifies the claim campaign.
	Code string
	// BadgeCode identifies the awarded badge.
	BadgeCode string
	// StartsAt stores the optional opening instant.
	StartsAt *time.Time
	// EndsAt stores the optional closing instant.
	EndsAt *time.Time
	// MaxClaims stores zero for unlimited claims.
	MaxClaims int64
	// Enabled controls availability.
	Enabled bool
}

// Catalog stores one immutable progression catalog generation.
type Catalog struct {
	// Achievements stores achievement definitions.
	Achievements []AchievementDefinition
	// Talents stores track levels.
	Talents []TalentLevel
	// Campaigns stores quest campaigns.
	Campaigns []QuestCampaign
	// Quests stores quest definitions.
	Quests []QuestDefinition
	// Quizzes stores quiz definitions.
	Quizzes []Quiz
	// Promos stores promo badge definitions.
	Promos []PromoBadge
}
