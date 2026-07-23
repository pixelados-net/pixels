package forum

import (
	"errors"
	"testing"
	"time"

	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// TestAllowsCoversPolicyMatrix verifies all social forum thresholds.
func TestAllowsCoversPolicyMatrix(t *testing.T) {
	tests := []struct {
		// name identifies the matrix case.
		name string
		// policy stores the forum threshold.
		policy grouprecord.Policy
		// access stores the viewer role and override.
		access Access
		// want stores the expected authorization result.
		want bool
	}{
		{name: "everyone", policy: grouprecord.Everyone, want: true},
		{name: "member", policy: grouprecord.Members, access: Access{Member: true, Role: grouprecord.Member}, want: true},
		{name: "outsider", policy: grouprecord.Members, want: false},
		{name: "admin", policy: grouprecord.Admins, access: Access{Member: true, Role: grouprecord.Admin}, want: true},
		{name: "owner only", policy: grouprecord.Owners, access: Access{Member: true, Role: grouprecord.Admin}, want: false},
		{name: "owner", policy: grouprecord.Owners, access: Access{Member: true, Role: grouprecord.Owner}, want: true},
		{name: "staff", policy: grouprecord.Owners, access: Access{Staff: true}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := allows(test.policy, test.access); got != test.want {
				t.Fatalf("allows=%v want=%v", got, test.want)
			}
		})
	}
}

// TestValidateTextBoundsAndRateLimits verifies bounded plain-text writes.
func TestValidateTextBoundsAndRateLimits(t *testing.T) {
	service := New(groupconfig.Config{ForumSubjectLimit: 5, ForumMessageLimit: 10, ForumPostCooldown: time.Minute}, nil, nil, nil, nil, nil, nil)
	subject, body, err := service.validateText(7, " hi ", " body ")
	if err != nil || subject != "hi" || body != "body" {
		t.Fatalf("subject=%q body=%q err=%v", subject, body, err)
	}
	if _, _, err = service.validateText(7, "hi", "body"); !errors.Is(err, grouprecord.ErrLimit) {
		t.Fatalf("expected rate limit, got %v", err)
	}
	if _, _, err = service.validateText(8, "longer", "body"); !errors.Is(err, grouprecord.ErrInvalid) {
		t.Fatalf("expected invalid subject, got %v", err)
	}
}

// TestCursorsDeduplicateAndExpireViewers verifies bounded forum viewer lifecycle.
func TestCursorsDeduplicateAndExpireViewers(t *testing.T) {
	cursors := NewCursors(groupconfig.Config{ForumCursorTTL: time.Hour})
	cursors.Set("one", Cursor{PlayerID: 7, GroupID: 3, ThreadID: 5})
	cursors.Set("two", Cursor{PlayerID: 7, GroupID: 3, ThreadID: 5})
	cursors.Set("three", Cursor{PlayerID: 8, GroupID: 4})
	viewers := cursors.Viewers(3)
	if len(viewers) != 1 || viewers[0] != 7 {
		t.Fatalf("viewers=%v", viewers)
	}
	cursor, found := cursors.Get("one")
	if !found || cursor.ThreadID != 5 {
		t.Fatalf("cursor=%#v found=%v", cursor, found)
	}
	cursors.Close("one")
	if _, found = cursors.Get("one"); found {
		t.Fatal("closed cursor remained visible")
	}
	expired := NewCursors(groupconfig.Config{ForumCursorTTL: time.Nanosecond})
	expired.Set("old", Cursor{PlayerID: 9, GroupID: 3})
	time.Sleep(time.Millisecond)
	if _, found = expired.Get("old"); found {
		t.Fatal("expired cursor remained visible")
	}
}
