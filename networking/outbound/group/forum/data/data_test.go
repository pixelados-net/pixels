package data

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestAppendRecordsWritesAllWireShapes verifies shared forum data encoding.
func TestAppendRecordsWritesAllWireShapes(t *testing.T) {
	now := time.Unix(100, 0)
	if payload, err := AppendSummary(nil, grouprecord.ForumSummary{Group: grouprecord.Group{ID: 1}}, now); err != nil || len(payload) == 0 {
		t.Fatal(err)
	}
	if payload, err := AppendThread(nil, grouprecord.Thread{ID: 1, CreatedAt: now}, now); err != nil || len(payload) == 0 {
		t.Fatal(err)
	}
	if payload, err := AppendPost(nil, grouprecord.Post{ID: 1, CreatedAt: now}, now); err != nil || len(payload) == 0 {
		t.Fatal(err)
	}
}
