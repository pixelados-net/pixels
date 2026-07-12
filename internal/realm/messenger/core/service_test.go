package core

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"github.com/niflaot/pixels/pkg/redis"
)

// TestSendRequestValidatesAndPersists verifies request validation and success.
func TestSendRequestValidatesAndPersists(t *testing.T) {
	service, store := serviceFixture(t)
	if _, err := service.SendRequest(context.Background(), 1, ""); err != ErrInvalidUsername {
		t.Fatalf("expected invalid username, got %v", err)
	}
	result, err := service.SendRequest(context.Background(), 1, "alice")
	if err != nil || !result.Sent || len(store.requests) != 1 {
		t.Fatalf("unexpected result=%#v requests=%#v err=%v", result, store.requests, err)
	}
	duplicate, err := service.SendRequest(context.Background(), 1, "alice")
	if err != nil || duplicate.Sent || len(store.requests) != 1 {
		t.Fatalf("unexpected duplicate=%#v requests=%#v err=%v", duplicate, store.requests, err)
	}
}

// TestAcceptAndRemoveMaintainSymmetry verifies both friendship directions.
func TestAcceptAndRemoveMaintainSymmetry(t *testing.T) {
	service, store := serviceFixture(t)
	store.requests = append(store.requests, messengermodel.Request{FromPlayerID: 2, ToPlayerID: 1})
	result, err := service.Accept(context.Background(), 1, 2)
	if err != nil || !result.Accepted || !store.friends[[2]int64{1, 2}] || !store.friends[[2]int64{2, 1}] {
		t.Fatalf("unexpected accept=%#v friends=%#v err=%v", result, store.friends, err)
	}
	removed, err := service.Remove(context.Background(), 1, []int64{2})
	if err != nil || len(removed) != 1 || store.friends[[2]int64{1, 2}] || store.friends[[2]int64{2, 1}] {
		t.Fatalf("unexpected removed=%#v friends=%#v err=%v", removed, store.friends, err)
	}
}

// TestSearchUsesThrottleAndSharedCache verifies Redis search behavior.
func TestSearchUsesThrottleAndSharedCache(t *testing.T) {
	service, store := serviceFixture(t)
	first, err := service.Search(context.Background(), 1, " Al ice ")
	if err != nil || len(first.Results) != 1 || store.searches != 1 {
		t.Fatalf("unexpected first=%#v searches=%d err=%v", first, store.searches, err)
	}
	throttled, err := service.Search(context.Background(), 1, "alice")
	if err != nil || !throttled.Throttled || store.searches != 1 {
		t.Fatalf("unexpected throttle=%#v searches=%d err=%v", throttled, store.searches, err)
	}
	cached, err := service.Search(context.Background(), 2, "alice")
	if err != nil || len(cached.Results) != 1 || store.searches != 1 {
		t.Fatalf("unexpected cache=%#v searches=%d err=%v", cached, store.searches, err)
	}
}

// TestCardsProfileAndRelationsProjectDurableAndLiveState verifies read projections.
func TestCardsProfileAndRelationsProjectDurableAndLiveState(t *testing.T) {
	service, store := serviceFixture(t)
	store.friends[[2]int64{1, 2}] = true
	store.friends[[2]int64{2, 1}] = true
	store.relations[[2]int64{1, 2}] = messengermodel.RelationHeart
	addLivePlayer(t, service.live, 2, "alice", 9)
	cards, err := service.Cards(context.Background(), 1)
	if err != nil || len(cards) != 1 || !cards[0].Online || !cards[0].FollowingAllowed || cards[0].Relation != messengermodel.RelationHeart {
		t.Fatalf("unexpected cards=%#v err=%v", cards, err)
	}
	profile, err := service.Profile(context.Background(), 1, 2)
	if err != nil || !profile.IsFriend || !profile.InRoom || profile.FriendCount != 1 {
		t.Fatalf("unexpected profile=%#v err=%v", profile, err)
	}
	if err = service.SetRelation(context.Background(), 1, 2, messengermodel.RelationSmile); err != nil || store.relations[[2]int64{1, 2}] != messengermodel.RelationSmile {
		t.Fatalf("unexpected relation=%d err=%v", store.relations[[2]int64{1, 2}], err)
	}
	if err = service.SetRelation(context.Background(), 1, 2, 9); err != ErrInvalidRelation {
		t.Fatalf("expected invalid relation, got %v", err)
	}
}

// TestPrivacyAndPrivateChatApplyRules verifies privacy and live private messages.
func TestPrivacyAndPrivateChatApplyRules(t *testing.T) {
	service, store := serviceFixture(t)
	store.friends[[2]int64{1, 2}] = true
	updated, err := service.SetRoomInvites(context.Background(), 2, true)
	if err != nil || !updated.Profile.BlockRoomInvites {
		t.Fatalf("unexpected privacy=%#v err=%v", updated.Profile, err)
	}
	message, err := service.SendPrivate(context.Background(), 1, 2, " hello ")
	if err != nil || message.Deliver || message.Message != "hello" {
		t.Fatalf("unexpected offline message=%#v err=%v", message, err)
	}
	throttled, err := service.SendPrivate(context.Background(), 1, 2, "again")
	if err != nil || !throttled.Throttled {
		t.Fatalf("unexpected throttle=%#v err=%v", throttled, err)
	}
}

// TestFollowAndInviteRespectPresenceAndPrivacy verifies social room behavior.
func TestFollowAndInviteRespectPresenceAndPrivacy(t *testing.T) {
	service, store := serviceFixture(t)
	store.friends[[2]int64{1, 2}] = true
	actor := addLivePlayer(t, service.live, 1, "demo", 4)
	_ = actor
	addLivePlayer(t, service.live, 2, "alice", 9)
	follow, err := service.Follow(context.Background(), 1, 2)
	if err != nil || follow.RoomID != 9 {
		t.Fatalf("unexpected follow=%#v err=%v", follow, err)
	}
	invite, err := service.Invite(context.Background(), 1, []int64{2}, " join ")
	if err != nil || invite.RoomID != 4 || len(invite.Delivered) != 1 {
		t.Fatalf("unexpected invite=%#v err=%v", invite, err)
	}
	players := service.players.(*fakePlayers)
	record := players.records[2]
	record.Profile.BlockRoomInvites = true
	players.records[2] = record
	invite, err = service.Invite(context.Background(), 1, []int64{2}, "join")
	if err != nil || len(invite.Blocked) != 1 {
		t.Fatalf("unexpected blocked invite=%#v err=%v", invite, err)
	}
}

// BenchmarkPresenceUpdates measures viewer-specific card projection.
func BenchmarkPresenceUpdates(b *testing.B) {
	service, store := serviceFixture(b)
	for id := int64(2); id < 102; id++ {
		store.records[id] = playerRecord(id, "friend")
		store.friendships[1] = append(store.friendships[1], messengermodel.Friendship{PlayerID: id, FriendPlayerID: 1})
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := service.PresenceUpdates(context.Background(), 1, func(int64) bool { return true }); err != nil {
			b.Fatal(err)
		}
	}
}

// testTB captures test and benchmark cleanup behavior.
type testTB interface {
	Helper()
	Cleanup(func())
	Fatalf(string, ...any)
}

// serviceFixture creates messenger behavior backed by deterministic fakes.
func serviceFixture(test testTB) (*Service, *fakeStore) {
	test.Helper()
	server, err := miniredis.Run()
	if err != nil {
		test.Fatalf("start redis: %v", err)
	}
	client := redis.New(redis.Config{Address: server.Addr()})
	test.Cleanup(func() {
		_ = client.Close()
		server.Close()
	})
	store := newFakeStore()
	players := &fakePlayers{records: map[int64]playerservice.Record{1: playerRecord(1, "demo"), 2: playerRecord(2, "alice")}}
	service := New(Options{MaxFriends: 200, MaxFriendsClub: 500, MaxSearchResults: 50, SearchCacheTTL: time.Minute, SearchThrottle: time.Second, ChatThrottle: time.Millisecond}, store, players, playerlive.NewRegistry(), roomlive.NewRegistry(nil), nil, client, nil, Nodes{}, nil)
	return service, store
}

// playerRecord creates one durable player fixture.
func playerRecord(id int64, username string) playerservice.Record {
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}, Username: username}, Profile: playermodel.Profile{PlayerID: id, Gender: playermodel.GenderMale, Look: "hd-180-1", Motto: "hello"}}
}

// addLivePlayer registers one player and optional room presence.
func addLivePlayer(t *testing.T, registry *playerlive.Registry, id int64, username string, roomID int64) *playerlive.Player {
	t.Helper()
	peer, err := playerlive.NewSessionPeer(netconn.ID(username), netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: id, Username: username}, peer)
	if err != nil {
		t.Fatal(err)
	}
	if roomID > 0 {
		if err = player.EnterRoom(roomID); err != nil {
			t.Fatal(err)
		}
	}
	if err = registry.Add(player); err != nil {
		t.Fatal(err)
	}
	return player
}

// fakePlayers stores player records by id.
type fakePlayers struct {
	// records stores player fixtures.
	records map[int64]playerservice.Record
}

// Create supplies unused player creation behavior.
func (players *fakePlayers) Create(context.Context, playerservice.CreateParams) (playerservice.Record, error) {
	return playerservice.Record{}, nil
}

// FindByID returns one player fixture.
func (players *fakePlayers) FindByID(_ context.Context, id int64) (playerservice.Record, bool, error) {
	record, found := players.records[id]
	return record, found, nil
}

// FindByUsername returns one case-insensitive player fixture.
func (players *fakePlayers) FindByUsername(_ context.Context, username string) (playerservice.Record, bool, error) {
	for _, record := range players.records {
		if record.Player.Username == username {
			return record, true, nil
		}
	}
	return playerservice.Record{}, false, nil
}

// UpdatePrivacy replaces one player privacy fixture.
func (players *fakePlayers) UpdatePrivacy(_ context.Context, playerID int64, params playerservice.PrivacyParams) (playerservice.Record, error) {
	record := players.records[playerID]
	record.Profile.BlockFriendRequests = params.BlockFriendRequests
	record.Profile.BlockRoomInvites = params.BlockRoomInvites
	record.Profile.BlockFollowing = params.BlockFollowing
	players.records[playerID] = record
	return record, nil
}
