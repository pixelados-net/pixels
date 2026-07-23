package quest

import (
	"context"
	"hash/fnv"
	"sort"
	"strconv"
	"time"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// Daily selects one stable daily quest for a player and UTC date.
func (service *Service) Daily(ctx context.Context, playerID int64, day time.Time, easy bool) (progressionrecord.QuestDefinition, bool, error) {
	rejected, err := service.store.DailyQuestRejected(ctx, playerID, day)
	if err != nil || rejected {
		return progressionrecord.QuestDefinition{}, false, err
	}
	generation := service.catalog.Current()
	if generation == nil {
		return progressionrecord.QuestDefinition{}, false, nil
	}
	pool := make([]progressionrecord.QuestDefinition, 0)
	for _, quest := range generation.Catalog.Quests {
		campaign := generation.CampaignByCode[quest.CampaignCode]
		if quest.Enabled && quest.Daily && quest.Easy == easy && campaign != nil && campaignAvailable(*campaign, day) {
			pool = append(pool, quest)
		}
	}
	if len(pool) == 0 {
		return progressionrecord.QuestDefinition{}, false, nil
	}
	sort.Slice(pool, func(left int, right int) bool { return pool[left].ID < pool[right].ID })
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(service.config.DailyPoolSeed))
	_, _ = hash.Write([]byte(day.UTC().Format("2006-01-02")))
	_, _ = hash.Write([]byte(strconv.FormatInt(playerID, 10)))
	return pool[int(hash.Sum64()%uint64(len(pool)))], true, nil
}

// DailyCounts returns currently available easy and hard daily pool sizes.
func (service *Service) DailyCounts(now time.Time) (int32, int32) {
	generation := service.catalog.Current()
	if generation == nil {
		return 0, 0
	}
	var easy int32
	var hard int32
	for _, quest := range generation.Catalog.Quests {
		campaign := generation.CampaignByCode[quest.CampaignCode]
		if !quest.Enabled || !quest.Daily || campaign == nil || !campaignAvailable(*campaign, now) {
			continue
		}
		if quest.Easy {
			easy++
		} else {
			hard++
		}
	}
	return easy, hard
}

// RejectDaily durably discards the current UTC-day offer.
func (service *Service) RejectDaily(ctx context.Context, playerID int64, day time.Time) error {
	return service.store.RejectDailyQuest(ctx, playerID, day)
}
