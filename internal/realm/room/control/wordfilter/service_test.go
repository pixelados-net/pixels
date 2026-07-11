package wordfilter

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// storeForTest stores words in memory.
type storeForTest struct {
	// words stores normalized words.
	words []string
	// lists counts persistence reads.
	lists int
}

// TestCensor verifies whole-word replacement and unchanged fast paths.
func TestCensor(t *testing.T) {
	service := New(&storeForTest{words: []string{"bad", "niño"}}, roomsForTest{}, nil)
	result, changed, err := service.Censor(context.Background(), 9, "bad badge niño")
	if err != nil || !changed || result != "*** badge ****" {
		t.Fatalf("result=%q changed=%v err=%v", result, changed, err)
	}
	result, changed, err = service.Censor(context.Background(), 9, "good")
	if err != nil || changed || result != "good" {
		t.Fatalf("result=%q changed=%v err=%v", result, changed, err)
	}
}

// List lists in-memory words.
func (store *storeForTest) List(context.Context, int64) ([]string, error) {
	store.lists++
	return append([]string(nil), store.words...), nil
}

// Add inserts one word when absent.
func (store *storeForTest) Add(_ context.Context, _ int64, word string) error {
	for _, current := range store.words {
		if current == word {
			return nil
		}
	}
	store.words = append(store.words, word)

	return nil
}

// Remove removes one word.
func (store *storeForTest) Remove(_ context.Context, _ int64, word string) error {
	for index, current := range store.words {
		if current == word {
			store.words = append(store.words[:index], store.words[index+1:]...)
			break
		}
	}

	return nil
}

// roomsForTest resolves one room.
type roomsForTest struct{ room roommodel.Room }

// FindByID finds one room.
func (rooms roomsForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return rooms.room, true, nil
}

// permissionsForTest allows configured nodes.
type permissionsForTest map[permission.Node]bool

// HasPermission reports one permission decision.
func (permissions permissionsForTest) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return permissions["own"], nil
}

// TestServiceMutatesCachesAndMatchesWholeWords verifies filter lifecycle and matching.
func TestServiceMutatesCachesAndMatchesWholeWords(t *testing.T) {
	store := &storeForTest{}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}
	authorizer := roomsettings.New(permissionsForTest{"own": true}, roomsettings.Nodes{OwnManage: "own", AnyManage: "any"})
	service := New(store, roomsForTest{room: room}, authorizer)
	if err := service.Add(context.Background(), 9, 1, " Demo "); err != nil {
		t.Fatalf("add: %v", err)
	}
	for _, testCase := range []struct {
		text    string
		blocked bool
	}{{"a DEMO message", true}, {"demonstration", false}, {"démo", false}} {
		blocked, err := service.Contains(context.Background(), 9, testCase.text)
		if err != nil || blocked != testCase.blocked {
			t.Fatalf("text=%q blocked=%v err=%v", testCase.text, blocked, err)
		}
	}
	if _, err := service.List(context.Background(), 9); err != nil {
		t.Fatalf("list: %v", err)
	}
	if store.lists != 1 {
		t.Fatalf("expected one cached persistence read, got %d", store.lists)
	}
	if err := service.Remove(context.Background(), 9, 1, "demo"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	blocked, err := service.Contains(context.Background(), 9, "demo")
	if err != nil || blocked {
		t.Fatalf("removed word blocked=%v err=%v", blocked, err)
	}
}

// TestServiceBoundsCachedRoomSnapshots verifies long-lived processes cap retained filters.
func TestServiceBoundsCachedRoomSnapshots(t *testing.T) {
	service := New(&storeForTest{words: []string{"spam"}}, roomsForTest{}, nil)
	for roomID := int64(1); roomID <= MaxCachedRooms+1; roomID++ {
		if _, err := service.Contains(context.Background(), roomID, "ordinary text"); err != nil {
			t.Fatalf("room %d: %v", roomID, err)
		}
	}
	if len(service.cache) > MaxCachedRooms {
		t.Fatalf("cache grew to %d entries", len(service.cache))
	}
}

// BenchmarkContainsCached measures the allocation-sensitive cached chat lookup path.
func BenchmarkContainsCached(b *testing.B) {
	store := &storeForTest{words: []string{"spam", "scam", "blocked"}}
	service := New(store, roomsForTest{}, nil)
	_, _ = service.Contains(context.Background(), 9, "warm cache")
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		blocked, err := service.Contains(ctx, 9, "ordinary room conversation without matches")
		if err != nil || blocked {
			b.Fatal("unexpected filter match")
		}
	}
}
