package core

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// deniedPermissions rejects every moderation capability.
type deniedPermissions struct{}

// HasPermission rejects the requested node.
func (deniedPermissions) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return false, nil
}

// TestReportValidationAndTopicActions verifies disabled, malformed, missing, and ignored reports.
func TestReportValidationAndTopicActions(t *testing.T) {
	ignoredStore := &moderationStoreForTest{topics: []moderationrecord.Topic{{ID: 1, Action: "ignore", Enabled: true}}}
	service := New(moderationconfig.Config{Enabled: false}, ignoredStore, nil, nil, nil, moderationPermissionsForTest{}, nil)
	if _, err := service.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: 1, TopicID: 1}); !errors.Is(err, ErrDisabled) {
		t.Fatalf("disabled err=%v", err)
	}
	service.config.Enabled = true
	if _, err := service.Report(context.Background(), moderationrecord.ReportParams{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid err=%v", err)
	}
	if err := service.RefreshTopics(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: 1, TopicID: 2}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing err=%v", err)
	}
	result, err := service.Report(context.Background(), moderationrecord.ReportParams{ReporterPlayerID: 1, TopicID: 1, Message: "ignored"})
	if err != nil || !result.Ignored || ignoredStore.created != 0 {
		t.Fatalf("result=%+v err=%v created=%d", result, err, ignoredStore.created)
	}
}

// TestTopicsAreDetached verifies callers cannot mutate the atomic cache generation.
func TestTopicsAreDetached(t *testing.T) {
	store := &moderationStoreForTest{topics: []moderationrecord.Topic{{ID: 1, Category: "help", Action: "queue", Enabled: true}}}
	service := New(moderationconfig.Config{Enabled: true}, store, nil, nil, nil, moderationPermissionsForTest{}, nil)
	if err := service.RefreshTopics(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := service.Topics()
	first[0].Category = "mutated"
	if service.Topics()[0].Category != "help" {
		t.Fatal("topic cache leaked mutable storage")
	}
}

// TestIssueLifecycle verifies ownership rules across pick, release, and close.
func TestIssueLifecycle(t *testing.T) {
	store := &moderationStoreForTest{issues: []moderationrecord.Issue{{ID: 1, State: "open"}}}
	service := New(moderationconfig.Config{Enabled: true}, store, nil, nil, nil, moderationPermissionsForTest{}, nil)
	if _, err := service.Pick(context.Background(), 7, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Pick(context.Background(), 8, 1); !errors.Is(err, ErrPickFailed) {
		t.Fatalf("second pick err=%v", err)
	}
	if _, err := service.Release(context.Background(), 8, 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign release err=%v", err)
	}
	if _, err := service.Release(context.Background(), 7, 1); err != nil {
		t.Fatal(err)
	}
	issue, err := service.Close(context.Background(), 9, 1, 3)
	if err != nil || issue.State != "resolved" || issue.Resolution == nil || *issue.Resolution != 3 {
		t.Fatalf("issue=%+v err=%v", issue, err)
	}
	if _, err = service.Close(context.Background(), 9, 1, 3); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second close err=%v", err)
	}
}

// TestIssueAuthorization verifies queue mutations are permission-gated.
func TestIssueAuthorization(t *testing.T) {
	store := &moderationStoreForTest{issues: []moderationrecord.Issue{{ID: 1, State: "open"}}}
	service := New(moderationconfig.Config{Enabled: true}, store, nil, nil, nil, deniedPermissions{}, nil)
	if _, err := service.Pick(context.Background(), 7, 1); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err=%v", err)
	}
}
