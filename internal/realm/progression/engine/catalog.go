// Package engine implements cached event-driven player progression.
package engine

import (
	"context"
	"strings"
	"sync/atomic"

	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// CatalogSource loads one complete progression catalog generation.
type CatalogSource interface {
	// Catalog loads all progression definitions.
	Catalog(context.Context) (progressionrecord.Catalog, error)
}

// Generation stores immutable lookup indexes.
type Generation struct {
	// Catalog stores the raw grouped catalog.
	Catalog progressionrecord.Catalog
	// AchievementByID indexes achievement definitions.
	AchievementByID map[int64]*progressionrecord.AchievementDefinition
	// AchievementsByTrigger indexes enabled achievement fan-out.
	AchievementsByTrigger map[string][]*progressionrecord.AchievementDefinition
	// QuestByID indexes enabled quests.
	QuestByID map[int64]*progressionrecord.QuestDefinition
	// QuestsByTrigger indexes enabled quest fan-out.
	QuestsByTrigger map[string][]*progressionrecord.QuestDefinition
	// CampaignByCode indexes enabled campaigns.
	CampaignByCode map[string]*progressionrecord.QuestCampaign
	// QuizByCode indexes enabled quizzes.
	QuizByCode map[string]*progressionrecord.Quiz
	// PromoByCode indexes enabled promotions.
	PromoByCode map[string]*progressionrecord.PromoBadge
	// TalentByTrack indexes ordered talent levels.
	TalentByTrack map[string][]progressionrecord.TalentLevel
	// TalentTracksByAchievement indexes relevant track names.
	TalentTracksByAchievement map[int64][]string
}

// Catalog caches immutable progression definitions.
type Catalog struct {
	// source loads durable generations.
	source CatalogSource
	// current stores the active immutable generation.
	current atomic.Pointer[Generation]
	// metrics stores process-wide catalog telemetry.
	metrics *progressionobservability.Metrics
}

// NewCatalog creates an unloaded progression catalog.
func NewCatalog(source CatalogSource) *Catalog { return &Catalog{source: source} }

// Reload atomically installs one freshly loaded generation.
func (catalog *Catalog) Reload(ctx context.Context) error {
	loaded, err := catalog.source.Catalog(ctx)
	if err != nil {
		return err
	}
	generation := buildGeneration(loaded)
	catalog.current.Store(&generation)
	if catalog.metrics != nil {
		catalog.metrics.SetCacheDefinitions(len(loaded.Achievements))
	}
	return nil
}

// SetMetrics attaches process-wide telemetry before the first reload.
func (catalog *Catalog) SetMetrics(metrics *progressionobservability.Metrics) {
	catalog.metrics = metrics
}

// Current returns the active immutable generation.
func (catalog *Catalog) Current() *Generation { return catalog.current.Load() }

// Achievements resolves one trigger without allocating.
func (catalog *Catalog) Achievements(trigger string) []*progressionrecord.AchievementDefinition {
	generation := catalog.current.Load()
	if generation == nil {
		return nil
	}
	return generation.AchievementsByTrigger[trigger]
}

// Quests resolves one trigger without allocating.
func (catalog *Catalog) Quests(trigger string) []*progressionrecord.QuestDefinition {
	generation := catalog.current.Load()
	if generation == nil {
		return nil
	}
	return generation.QuestsByTrigger[trigger]
}

// buildGeneration creates immutable indexes with normalized keys.
func buildGeneration(loaded progressionrecord.Catalog) Generation {
	generation := Generation{
		Catalog: loaded, AchievementByID: make(map[int64]*progressionrecord.AchievementDefinition, len(loaded.Achievements)),
		AchievementsByTrigger: make(map[string][]*progressionrecord.AchievementDefinition), QuestByID: make(map[int64]*progressionrecord.QuestDefinition, len(loaded.Quests)),
		QuestsByTrigger: make(map[string][]*progressionrecord.QuestDefinition), CampaignByCode: make(map[string]*progressionrecord.QuestCampaign, len(loaded.Campaigns)),
		QuizByCode: make(map[string]*progressionrecord.Quiz, len(loaded.Quizzes)), PromoByCode: make(map[string]*progressionrecord.PromoBadge, len(loaded.Promos)),
		TalentByTrack: make(map[string][]progressionrecord.TalentLevel), TalentTracksByAchievement: make(map[int64][]string),
	}
	for index := range loaded.Achievements {
		definition := &generation.Catalog.Achievements[index]
		generation.AchievementByID[definition.ID] = definition
		if definition.Enabled {
			key := strings.TrimSpace(definition.TriggerKey)
			generation.AchievementsByTrigger[key] = append(generation.AchievementsByTrigger[key], definition)
		}
	}
	indexSecondary(&generation)
	return generation
}

// indexSecondary builds quest, quiz, promo, and talent lookup maps.
func indexSecondary(generation *Generation) {
	for index := range generation.Catalog.Campaigns {
		campaign := &generation.Catalog.Campaigns[index]
		if campaign.Enabled {
			generation.CampaignByCode[campaign.Code] = campaign
		}
	}
	for index := range generation.Catalog.Quests {
		quest := &generation.Catalog.Quests[index]
		generation.QuestByID[quest.ID] = quest
		if quest.Enabled {
			generation.QuestsByTrigger[quest.TriggerKey] = append(generation.QuestsByTrigger[quest.TriggerKey], quest)
		}
	}
	for index := range generation.Catalog.Quizzes {
		quiz := &generation.Catalog.Quizzes[index]
		if quiz.Enabled {
			generation.QuizByCode[strings.ToUpper(quiz.Code)] = quiz
		}
	}
	for index := range generation.Catalog.Promos {
		promo := &generation.Catalog.Promos[index]
		if promo.Enabled {
			generation.PromoByCode[strings.ToUpper(promo.Code)] = promo
		}
	}
	indexTalents(generation)
}

// indexTalents builds track and reverse achievement indexes.
func indexTalents(generation *Generation) {
	for _, level := range generation.Catalog.Talents {
		generation.TalentByTrack[level.Track] = append(generation.TalentByTrack[level.Track], level)
		for _, requirement := range level.Requirements {
			tracks := generation.TalentTracksByAchievement[requirement.DefinitionID]
			found := false
			for _, track := range tracks {
				found = found || track == level.Track
			}
			if !found {
				generation.TalentTracksByAchievement[requirement.DefinitionID] = append(tracks, level.Track)
			}
		}
	}
}
