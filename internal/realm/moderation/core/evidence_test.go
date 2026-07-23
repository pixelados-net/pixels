package core

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	historyrepo "github.com/niflaot/pixels/internal/realm/chat/history/repository"
	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// historyStoreForTest embeds unused history behavior and returns fixed evidence.
type historyStoreForTest struct {
	// Store supplies unused history persistence methods.
	historyrepo.Store
	// entries stores server-authoritative evidence.
	entries []historymodel.Entry
}

// History returns configured server evidence.
func (store historyStoreForTest) History(context.Context, historymodel.Query) ([]historymodel.Entry, error) {
	return append([]historymodel.Entry(nil), store.entries...), nil
}

// escalationStoreForTest embeds unused sanction persistence and records one insert.
type escalationStoreForTest struct {
	// Store supplies unused sanction persistence methods.
	sanctionrecord.Store
	// inserted stores the common-engine request.
	inserted *sanctionrecord.ApplyParams
}

// Ladder returns one warning level.
func (*escalationStoreForTest) Ladder(context.Context) ([]sanctionrecord.LadderEntry, error) {
	return []sanctionrecord.LadderEntry{{Level: 1, Kind: sanctionrecord.KindWarn, ProbationDays: 1}}, nil
}

// LastEscalation reports no prior level.
func (*escalationStoreForTest) LastEscalation(context.Context, int64) (sanctionrecord.Punishment, int32, bool, error) {
	return sanctionrecord.Punishment{}, 0, false, nil
}

// Insert records one common-engine request.
func (store *escalationStoreForTest) Insert(_ context.Context, params sanctionrecord.ApplyParams) (sanctionrecord.Punishment, error) {
	store.inserted = &params
	return sanctionrecord.Punishment{ID: 1, ReceiverPlayerID: params.ReceiverPlayerID, Kind: params.Kind, Reason: params.Reason, Source: params.Source, IssuedAt: time.Now()}, nil
}

// escalationPermissions permits automation while preserving target non-immunity.
type escalationPermissions struct{}

// HasPermission returns false only for target immunity.
func (escalationPermissions) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return node != sanctioncore.ImmuneNode, nil
}

// TestReportFreezesServerAndProtocolEvidence verifies authoritative history is copied in order.
func TestReportFreezesServerAndProtocolEvidence(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	historyService := history.NewService(historyStoreForTest{entries: []historymodel.Entry{{PlayerID: 2, Kind: "talk", Message: "new", CreatedAt: now}, {PlayerID: 2, Kind: "talk", Message: "old", CreatedAt: now.Add(-time.Second)}}})
	store := &moderationStoreForTest{topics: []moderationrecord.Topic{{ID: 1, Action: "queue", Enabled: true}}}
	service := New(moderationconfig.Config{Enabled: true, ContextWindow: 2}, store, historyService, nil, nil, moderationPermissionsForTest{}, bus.New())
	service.now = func() time.Time { return now }
	if err := service.RefreshTopics(context.Background()); err != nil {
		t.Fatal(err)
	}
	target := int64(2)
	result, err := service.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: 1, ReportedPlayerID: &target, TopicID: 1, Message: "report", Chatlog: []moderationrecord.ChatEntry{{PatternID: "IM", Message: " extra "}}})
	if err != nil || result.Issue == nil || len(result.Issue.Chatlog) != 3 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if result.Issue.Chatlog[0].Message != "old" || result.Issue.Chatlog[1].Message != "new" || result.Issue.Chatlog[2].Message != "extra" {
		t.Fatalf("evidence=%+v", result.Issue.Chatlog)
	}
}

// TestCloseWithDefaultUsesCommonSanctionEngine verifies issue escalation wiring.
func TestCloseWithDefaultUsesCommonSanctionEngine(t *testing.T) {
	target := int64(2)
	issueStore := &moderationStoreForTest{issues: []moderationrecord.Issue{{ID: 1, ReporterPlayerID: 1, ReportedPlayerID: &target, TopicID: 1, Message: "reason", State: "open"}}}
	service := New(moderationconfig.Config{Enabled: true}, issueStore, nil, nil, nil, moderationPermissionsForTest{}, bus.New())
	sanctionStore := &escalationStoreForTest{}
	sanctions := sanctioncore.New(sanctionStore, escalationPermissions{}, nil, nil)
	issue, err := service.CloseWithDefault(context.Background(), sanctions, 7, 1, 3)
	if err != nil || issue.State != "resolved" || sanctionStore.inserted == nil || sanctionStore.inserted.ReceiverPlayerID != target || sanctionStore.inserted.IssueID == nil || *sanctionStore.inserted.IssueID != 1 {
		t.Fatalf("issue=%+v inserted=%+v err=%v", issue, sanctionStore.inserted, err)
	}
}

// TestSimpleServiceAccessors verifies bounded sanitation and empty-cache behavior.
func TestSimpleServiceAccessors(t *testing.T) {
	store := &moderationStoreForTest{}
	service := New(moderationconfig.Config{Enabled: true}, store, nil, nil, nil, moderationPermissionsForTest{}, nil)
	if !service.Enabled() || service.Store() != store || service.Topics() != nil {
		t.Fatal("service accessor mismatch")
	}
	if value := service.Sanitize("  " + strings.Repeat("x", 600) + "  "); len(value) != 500 {
		t.Fatalf("sanitized length=%d", len(value))
	}
}
