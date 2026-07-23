package lovelock

import (
	"context"
	"testing"

	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
)

// storeForTest records guarded lovelock mutations.
type storeForTest struct {
	// finish stores whether confirmation was requested.
	finish bool
	// cancel stores whether cancellation was requested.
	cancel bool
}

// Start reports an occupied handshake for confirmation tests.
func (store *storeForTest) Start(context.Context, int64, int64) (Pending, bool, error) {
	return Pending{}, false, nil
}

// Invite reports an occupied handshake for confirmation tests.
func (store *storeForTest) Invite(context.Context, int64, int64) (Pending, bool, error) {
	return Pending{}, false, nil
}

// Finish records a successful confirmation.
func (store *storeForTest) Finish(context.Context, int64, int64) (bool, error) {
	store.finish = true
	return true, nil
}

// Cancel records a successful cancellation.
func (store *storeForTest) Cancel(context.Context, int64, int64) (bool, error) {
	store.cancel = true
	return true, nil
}

// TestConfirmRoutesAcceptAndCancel verifies both client decision branches.
func TestConfirmRoutesAcceptAndCancel(t *testing.T) {
	store := &storeForTest{}
	service := New(store, nil)
	request := essential.Request{PlayerID: 8}
	request.Item.ID = 9
	if err := service.Confirm(context.Background(), request, true); err != nil || !store.finish {
		t.Fatalf("finish=%t err=%v", store.finish, err)
	}
	if err := service.Confirm(context.Background(), request, false); err != nil || !store.cancel {
		t.Fatalf("cancel=%t err=%v", store.cancel, err)
	}
}

// BenchmarkConfirm measures the guarded confirmation dispatch without room delivery.
func BenchmarkConfirm(b *testing.B) {
	service := New(&storeForTest{}, nil)
	request := essential.Request{PlayerID: 8}
	request.Item.ID = 9
	for range b.N {
		_ = service.Confirm(context.Background(), request, true)
	}
}
