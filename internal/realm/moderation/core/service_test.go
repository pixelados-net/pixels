package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// moderationStoreForTest supplies deterministic topics and atomic issue claims.
type moderationStoreForTest struct {
	mutex   sync.Mutex
	topics  []moderationrecord.Topic
	issues  []moderationrecord.Issue
	created int
}

// Topics returns configured topics.
func (store *moderationStoreForTest) Topics(context.Context, bool) ([]moderationrecord.Topic, error) {
	return append([]moderationrecord.Topic(nil), store.topics...), nil
}

// Topic returns one configured topic.
func (store *moderationStoreForTest) Topic(_ context.Context, id int64) (moderationrecord.Topic, bool, error) {
	for _, value := range store.topics {
		if value.ID == id {
			return value, true, nil
		}
	}
	return moderationrecord.Topic{}, false, nil
}

// CreateTopic appends one topic.
func (store *moderationStoreForTest) CreateTopic(_ context.Context, value moderationrecord.Topic) (moderationrecord.Topic, error) {
	store.topics = append(store.topics, value)
	return value, nil
}

// UpdateTopic replaces one topic.
func (*moderationStoreForTest) UpdateTopic(_ context.Context, value moderationrecord.Topic) (moderationrecord.Topic, bool, error) {
	return value, true, nil
}

// CreateIssue stores detached report evidence.
func (store *moderationStoreForTest) CreateIssue(_ context.Context, params moderationrecord.ReportParams) (moderationrecord.Issue, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.created++
	issue := moderationrecord.Issue{ID: int64(store.created), ReporterPlayerID: params.ReporterPlayerID, ReportedPlayerID: params.ReportedPlayerID, TopicID: params.TopicID, Message: params.Message, State: "open", Chatlog: append([]moderationrecord.ChatEntry(nil), params.Chatlog...), CreatedAt: time.Now()}
	store.issues = append(store.issues, issue)
	return issue, nil
}

// Issue returns one issue.
func (store *moderationStoreForTest) Issue(_ context.Context, id int64, _ bool) (moderationrecord.Issue, bool, error) {
	for _, value := range store.issues {
		if value.ID == id {
			return value, true, nil
		}
	}
	return moderationrecord.Issue{}, false, nil
}

// Issues returns configured issues.
func (store *moderationStoreForTest) Issues(context.Context, string, int32) ([]moderationrecord.Issue, error) {
	return append([]moderationrecord.Issue(nil), store.issues...), nil
}

// Pending returns no issues.
func (*moderationStoreForTest) Pending(context.Context, int64) ([]moderationrecord.Issue, error) {
	return nil, nil
}

// DeletePending returns no issue ids.
func (*moderationStoreForTest) DeletePending(context.Context, int64) ([]int64, error) {
	return nil, nil
}

// Pick atomically claims one issue.
func (store *moderationStoreForTest) Pick(_ context.Context, id int64, moderatorID int64) (moderationrecord.Issue, bool, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	for index := range store.issues {
		if store.issues[index].ID == id && store.issues[index].State == "open" {
			store.issues[index].State = "picked"
			store.issues[index].PickedByPlayerID = &moderatorID
			return store.issues[index], true, nil
		}
	}
	return moderationrecord.Issue{}, false, nil
}

// Release returns one moderator-owned test issue to the queue.
func (store *moderationStoreForTest) Release(_ context.Context, id int64, moderatorID int64) (moderationrecord.Issue, bool, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	for index := range store.issues {
		picker := store.issues[index].PickedByPlayerID
		if store.issues[index].ID == id && store.issues[index].State == "picked" && picker != nil && *picker == moderatorID {
			store.issues[index].State = "open"
			store.issues[index].PickedByPlayerID = nil
			return store.issues[index], true, nil
		}
	}
	return moderationrecord.Issue{}, false, nil
}

// Close resolves one available test issue.
func (store *moderationStoreForTest) Close(_ context.Context, id int64, moderatorID int64, resolution int32) (moderationrecord.Issue, bool, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	for index := range store.issues {
		picker := store.issues[index].PickedByPlayerID
		owned := picker == nil || *picker == moderatorID
		if store.issues[index].ID == id && (store.issues[index].State == "open" || store.issues[index].State == "picked") && owned {
			store.issues[index].State = "resolved"
			store.issues[index].Resolution = &resolution
			return store.issues[index], true, nil
		}
	}
	return moderationrecord.Issue{}, false, nil
}

// Preferences returns default geometry.
func (*moderationStoreForTest) Preferences(context.Context, int64) (moderationrecord.Preferences, error) {
	return moderationrecord.Preferences{}, nil
}

// Presets returns no presets.
func (*moderationStoreForTest) Presets(context.Context, bool) ([]moderationrecord.Preset, error) {
	return nil, nil
}

// CreatePreset accepts one preset.
func (*moderationStoreForTest) CreatePreset(_ context.Context, value moderationrecord.Preset) (moderationrecord.Preset, error) {
	return value, nil
}

// UpdatePreset accepts one preset.
func (*moderationStoreForTest) UpdatePreset(_ context.Context, value moderationrecord.Preset) (moderationrecord.Preset, bool, error) {
	return value, true, nil
}

// Visits returns no room visits.
func (*moderationStoreForTest) Visits(context.Context, int64, int32) ([]moderationrecord.RoomVisit, error) {
	return nil, nil
}

// InsertFeedback accepts guide feedback.
func (*moderationStoreForTest) InsertFeedback(context.Context, int64, int64, bool) error { return nil }

// CreateGuardianTicket returns one test ticket id.
func (*moderationStoreForTest) CreateGuardianTicket(context.Context, int64, int64, time.Time) (int64, error) {
	return 1, nil
}

// SaveGuardianVote accepts one test verdict.
func (*moderationStoreForTest) SaveGuardianVote(context.Context, int64, int64, int32) error {
	return nil
}

// CloseGuardianTicket accepts one test result.
func (*moderationStoreForTest) CloseGuardianTicket(context.Context, int64, int32) error { return nil }

// moderationPermissionsForTest allows all staff operations.
type moderationPermissionsForTest struct{}

// HasPermission permits every node.
func (moderationPermissionsForTest) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return true, nil
}

// TestReportAutoReplyDoesNotEnqueue verifies topic actions avoid queue writes.
func TestReportAutoReplyDoesNotEnqueue(t *testing.T) {
	reply := "moderation.reply.help"
	store := &moderationStoreForTest{topics: []moderationrecord.Topic{{ID: 1, Action: "auto_reply", AutoReplyKey: &reply, Enabled: true}}}
	service := New(moderationconfig.Config{Enabled: true, ContextWindow: 10}, store, nil, nil, nil, moderationPermissionsForTest{}, nil)
	if err := service.RefreshTopics(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err := service.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: 1, TopicID: 1, Message: "help"})
	if err != nil || result.ReplyKey != reply || store.created != 0 {
		t.Fatalf("result=%+v err=%v created=%d", result, err, store.created)
	}
}

// TestReportFreezesDetachedEvidence verifies caller mutations cannot alter an issue.
func TestReportFreezesDetachedEvidence(t *testing.T) {
	store := &moderationStoreForTest{topics: []moderationrecord.Topic{{ID: 1, Action: "queue", Enabled: true}}}
	service := New(moderationconfig.Config{Enabled: true, ContextWindow: 10}, store, nil, nil, nil, moderationPermissionsForTest{}, nil)
	service.now = func() time.Time { return time.Unix(10, 0) }
	if err := service.RefreshTopics(context.Background()); err != nil {
		t.Fatal(err)
	}
	evidence := []moderationrecord.ChatEntry{{Message: "original"}}
	result, err := service.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: 1, TopicID: 1, Message: " report ", Chatlog: evidence})
	evidence[0].Message = "mutated"
	if err != nil || result.Issue == nil || result.Issue.Chatlog[0].Message != "original" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

// TestPickHasExactlyOneConcurrentWinner verifies atomic queue ownership semantics.
func TestPickHasExactlyOneConcurrentWinner(t *testing.T) {
	store := &moderationStoreForTest{issues: []moderationrecord.Issue{{ID: 1, State: "open"}}}
	service := New(moderationconfig.Config{Enabled: true}, store, nil, nil, nil, moderationPermissionsForTest{}, nil)
	results := make(chan error, 2)
	for moderatorID := int64(1); moderatorID <= 2; moderatorID++ {
		go func(id int64) { _, err := service.Pick(context.Background(), id, 1); results <- err }(moderatorID)
	}
	winners := 0
	for range 2 {
		if <-results == nil {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("winners=%d", winners)
	}
}
