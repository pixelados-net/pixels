package poll

import (
	"context"
	"errors"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomgrid "github.com/niflaot/pixels/internal/realm/room/world/grid"
	roompath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
)

// pollFixture builds one occupied live room and captures broadcasts.
func pollFixture(t testing.TB) (*Service, *[]codec.Packet) {
	t.Helper()
	registry := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	room, err := registry.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 42, MaxUsers: 10})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := roomgrid.Parse("00", roomgrid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	door, valid := roomgrid.NewPoint(0, 0)
	if !valid {
		t.Fatal("invalid test door")
	}
	if err = room.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: roompath.Position{Point: door}}); err != nil {
		t.Fatal(err)
	}
	for _, playerID := range []int64{42, 43} {
		occupant := roomlive.Occupant{PlayerID: playerID, ConnectionID: netconn.ID("test"), ConnectionKind: netconn.Kind("test")}
		if _, err = registry.Join(context.Background(), 9, occupant); err != nil {
			t.Fatal(err)
		}
	}
	packets := make([]codec.Packet, 0)
	service := New(registry, func(_ context.Context, _ *roomlive.Room, packet codec.Packet, _ int64) error {
		packets = append(packets, packet)
		return nil
	})
	t.Cleanup(func() {
		service.Close()
		_, _, _ = registry.Close(context.Background(), 9)
	})
	return service, &packets
}

// databasePollStore stores deterministic DB poll test state.
type databasePollStore struct {
	// definitions stores enabled cached polls.
	definitions []Definition
	// completed stores per-player completion.
	completed map[int64]bool
	// roomQueries counts forbidden entry hot-path queries.
	roomQueries int
}

// Polls returns enabled definitions.
func (store *databasePollStore) Polls(context.Context) ([]Definition, error) {
	return append([]Definition(nil), store.definitions...), nil
}

// Poll returns one definition.
func (store *databasePollStore) Poll(_ context.Context, id int32) (Definition, bool, error) {
	for _, definition := range store.definitions {
		if definition.ID == id {
			return definition, true, nil
		}
	}
	return Definition{}, false, nil
}

// PollForRoom records an unexpected hot-path query.
func (store *databasePollStore) PollForRoom(context.Context, int64) (Definition, bool, error) {
	store.roomQueries++
	return Definition{}, false, nil
}

// Completed returns deterministic completion state.
func (store *databasePollStore) Completed(_ context.Context, playerID int64, _ int32) (bool, error) {
	return store.completed[playerID], nil
}

// SaveAnswer completes one poll exactly once.
func (store *databasePollStore) SaveAnswer(_ context.Context, playerID int64, _ int32, _ int32, _ []string) (bool, string, error) {
	if store.completed[playerID] {
		return false, "", ErrInvalidAnswer
	}
	store.completed[playerID] = true
	return true, "GAMES_TEST", nil
}

// RejectPoll records completion.
func (store *databasePollStore) RejectPoll(_ context.Context, playerID int64, _ int32) error {
	store.completed[playerID] = true
	return nil
}

// badgeCapture records idempotent reward calls.
type badgeCapture struct{ calls int }

// GrantBadge records one grant.
func (capture *badgeCapture) GrantBadge(context.Context, int64, string, string) (bool, error) {
	capture.calls++
	return true, nil
}

// TestPollLifecycle verifies launch, occupant vote, duplicate rejection, and finish.
func TestPollLifecycle(t *testing.T) {
	service, packets := pollFixture(t)
	id, err := service.Start(context.Background(), 9, 42, "¿Sí o no?", time.Minute)
	if err != nil || id != 1 || len(*packets) != 1 {
		t.Fatalf("id=%d packets=%d err=%v", id, len(*packets), err)
	}
	if _, found, err := service.Current(43, id); err != nil || !found {
		t.Fatalf("current found=%v err=%v", found, err)
	}
	if err = service.Answer(context.Background(), 43, id, -1, []string{"1"}); err != nil {
		t.Fatal(err)
	}
	values, _, decodeErr := codec.DecodePacket((*packets)[1], codec.Definition{codec.Int32Field})
	if decodeErr != nil || values[0].Int32 != 43 {
		t.Fatalf("answered user id=%d err=%v", values[0].Int32, decodeErr)
	}
	if err = service.Answer(context.Background(), 43, id, -1, []string{"1"}); !errors.Is(err, ErrInvalidAnswer) {
		t.Fatalf("duplicate error %v", err)
	}
	if err = service.Finish(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if _, found, err := service.Current(43, id); err != nil || found || len(*packets) != 3 {
		t.Fatalf("current found=%v packets=%d err=%v", found, len(*packets), err)
	}
	if err = service.Finish(context.Background(), id); err != nil {
		t.Fatal(err)
	}
}

// TestPollRequiresRoomRights verifies an occupant cannot launch a room poll.
func TestPollRequiresRoomRights(t *testing.T) {
	service, _ := pollFixture(t)
	if _, err := service.Start(context.Background(), 9, 43, "no", time.Minute); !errors.Is(err, ErrForbidden) {
		t.Fatalf("error %v", err)
	}
}

// TestPollBroadcastFailureRollsBackRuntime verifies failed launch has no ghost poll.
func TestPollBroadcastFailureRollsBackRuntime(t *testing.T) {
	service, _ := pollFixture(t)
	service.broadcast = func(context.Context, *roomlive.Room, codec.Packet, int64) error { return errors.New("send") }
	id, err := service.Start(context.Background(), 9, 42, "test", time.Minute)
	if err == nil || id != 0 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if _, found, currentErr := service.Current(42, 1); currentErr != nil || found {
		t.Fatalf("ghost poll found=%v err=%v", found, currentErr)
	}
}

// TestDatabasePollCacheOfferAndUniqueReward verifies O(1) room offers and one completion badge.
func TestDatabasePollCacheOfferAndUniqueReward(t *testing.T) {
	service, _ := pollFixture(t)
	store := &databasePollStore{definitions: []Definition{{ID: 7, RoomID: 9, Enabled: true, Headline: "Poll", Summary: "Answer", StartMessage: "Start", ThanksMessage: "Thanks", RewardBadge: "GAMES_TEST", Questions: []outcontents.Question{{ID: 8, SortOrder: 1, Type: 1, Text: "Ready?", Choices: []outcontents.Choice{{Value: "yes", Text: "Yes"}}}}}}, completed: make(map[int64]bool)}
	badges := &badgeCapture{}
	service.WithDatabase(store, badges)
	if err := service.ReloadDatabase(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, offered, err := service.OfferForRoom(context.Background(), 43, 9); err != nil || !offered || store.roomQueries != 0 {
		t.Fatalf("offer=%v room_queries=%d err=%v", offered, store.roomQueries, err)
	}
	if err := service.AnswerDatabase(context.Background(), 43, 7, 8, []string{"yes"}); err != nil || badges.calls != 1 {
		t.Fatalf("badge calls=%d err=%v", badges.calls, err)
	}
	if err := service.AnswerDatabase(context.Background(), 43, 7, 8, []string{"yes"}); !errors.Is(err, ErrInvalidAnswer) || badges.calls != 1 {
		t.Fatalf("duplicate badge calls=%d err=%v", badges.calls, err)
	}
	if _, offered, err := service.OfferForRoom(context.Background(), 43, 9); err != nil || offered {
		t.Fatalf("repeat offer=%v err=%v", offered, err)
	}
}

// TestDatabasePollRejectsAnswersOutsideTheDefinition verifies server-authored choices.
func TestDatabasePollRejectsAnswersOutsideTheDefinition(t *testing.T) {
	service, _ := pollFixture(t)
	store := &databasePollStore{definitions: []Definition{{ID: 7, Enabled: true, Questions: []outcontents.Question{{ID: 8, Type: 1, Text: "Ready?", Choices: []outcontents.Choice{{Value: "yes", Text: "Yes"}}}, {ID: 9, Type: 2, Text: "Pick", Choices: []outcontents.Choice{{Value: "a", Text: "A"}, {Value: "b", Text: "B"}}}, {ID: 10, Type: 0, Text: "Say"}}}}, completed: make(map[int64]bool)}
	service.WithDatabase(store, nil)
	if err := service.ReloadDatabase(context.Background()); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		// questionID identifies the submitted question.
		questionID int32
		// values stores the rejected client values.
		values []string
	}{{8, []string{"no"}}, {8, []string{"yes", "yes"}}, {9, []string{"a", "a"}}, {9, []string{"c"}}, {10, []string{""}}, {99, []string{"yes"}}}
	for _, test := range tests {
		if err := service.AnswerDatabase(context.Background(), 43, 7, test.questionID, test.values); !errors.Is(err, ErrInvalidAnswer) {
			t.Fatalf("question=%d values=%v error=%v", test.questionID, test.values, err)
		}
	}
}
