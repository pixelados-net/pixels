package votes

import (
	"context"
	"errors"
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outscore "github.com/niflaot/pixels/networking/outbound/room/score"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// fakeStore stores deterministic vote behavior for tests.
type fakeStore struct {
	// mutation stores the next cast result.
	mutation Mutation
	// voted reports one persisted vote.
	voted bool
	// voters stores batched active voters.
	voters map[int64]struct{}
	// casts stores mutation call count.
	casts int
	// items stores listed votes.
	items []Vote
}

// Cast returns the configured mutation.
func (store *fakeStore) Cast(context.Context, int64, int64) (Mutation, error) {
	store.casts++
	return store.mutation, nil
}

// HasVote returns configured eligibility.
func (store *fakeStore) HasVote(context.Context, int64, int64) (bool, error) { return store.voted, nil }

// Existing returns configured active voters.
func (store *fakeStore) Existing(context.Context, int64, []int64) (map[int64]struct{}, error) {
	return store.voters, nil
}

// List returns no configured votes.
func (store *fakeStore) List(context.Context, Query) ([]Vote, error) { return store.items, nil }

// fakeRooms resolves one room.
type fakeRooms struct {
	// room stores the resolved room.
	room roommodel.Room
}

// FindByID resolves the configured room.
func (rooms fakeRooms) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return rooms.room, rooms.room.ID > 0, nil
}

// capturePublisher records published events.
type capturePublisher struct {
	// events stores published events.
	events []bus.Event
}

// Publish records one event.
func (publisher *capturePublisher) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestCastIsIdempotentAndRejectsOwnerVotes verifies vote mutation rules.
func TestCastIsIdempotentAndRejectsOwnerVotes(t *testing.T) {
	store := &fakeStore{mutation: Mutation{Score: 8, Inserted: true}}
	events := &capturePublisher{}
	service := New(store, fakeRooms{room: voteRoom(7, 1, 7)}, nil, nil, events)
	result, err := service.Cast(context.Background(), 7, 2)
	if err != nil || !result.Inserted || result.Score != 8 || store.casts != 1 || len(events.events) != 1 {
		t.Fatalf("cast result=%+v casts=%d events=%d err=%v", result, store.casts, len(events.events), err)
	}
	store.mutation = Mutation{Score: 8}
	result, err = service.Cast(context.Background(), 7, 2)
	if err != nil || result.Inserted || result.Score != 8 || len(events.events) != 1 {
		t.Fatalf("duplicate result=%+v events=%d err=%v", result, len(events.events), err)
	}
	result, err = service.Cast(context.Background(), 7, 1)
	if err != nil || result.Inserted || result.Score != 7 || store.casts != 2 {
		t.Fatalf("owner result=%+v casts=%d err=%v", result, store.casts, err)
	}
}

// TestStateReportsOwnerAndVoterEligibility verifies permanent eligibility rules.
func TestStateReportsOwnerAndVoterEligibility(t *testing.T) {
	store := &fakeStore{voted: true}
	service := New(store, fakeRooms{room: voteRoom(7, 1, 4)}, nil, nil, nil)
	owner, err := service.State(context.Background(), 7, 1)
	if err != nil || owner.CanVote || owner.Voted || owner.Score != 4 {
		t.Fatalf("owner state=%+v err=%v", owner, err)
	}
	voter, err := service.State(context.Background(), 7, 2)
	if err != nil || voter.CanVote || !voter.Voted || voter.Score != 4 {
		t.Fatalf("voter state=%+v err=%v", voter, err)
	}
}

// TestStateAndListValidateInputs verifies read validation and new-voter eligibility.
func TestStateAndListValidateInputs(t *testing.T) {
	store := &fakeStore{items: []Vote{{RoomID: 7, PlayerID: 2}}}
	service := New(store, fakeRooms{room: voteRoom(7, 1, 4)}, nil, nil, nil)
	state, err := service.State(context.Background(), 7, 2)
	if err != nil || !state.CanVote || state.Voted {
		t.Fatalf("state=%+v err=%v", state, err)
	}
	items, err := service.List(context.Background(), Query{RoomID: 7})
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if _, err := service.State(context.Background(), 0, 2); !errors.Is(err, ErrInvalidRoomID) {
		t.Fatalf("invalid room error=%v", err)
	}
	if _, err := service.State(context.Background(), 7, 0); !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("invalid player error=%v", err)
	}
	if _, err := service.List(context.Background(), Query{}); !errors.Is(err, ErrInvalidRoomID) {
		t.Fatalf("invalid list error=%v", err)
	}
	invalidPlayer := int64(0)
	if _, err := (Query{RoomID: 7, PlayerID: &invalidPlayer}).Normalize(); !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("invalid query player error=%v", err)
	}
	normalized, err := (Query{RoomID: 7, Limit: MaxLimit + 1}).Normalize()
	if err != nil || normalized.Limit != MaxLimit {
		t.Fatalf("normalized=%+v err=%v", normalized, err)
	}
}

// TestBroadcastProjectsIndividualEligibility verifies one score with per-player controls.
func TestBroadcastProjectsIndividualEligibility(t *testing.T) {
	runtime := roomlive.NewRegistry(nil)
	_, err := runtime.Activate(roomlive.Snapshot{ID: 7, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	connections := netconn.NewRegistry()
	ownerPackets := registerVoteConnection(t, connections, "owner")
	voterPackets := registerVoteConnection(t, connections, "voter")
	actorPackets := registerVoteConnection(t, connections, "actor")
	eligiblePackets := registerVoteConnection(t, connections, "eligible")
	joinVoteOccupant(t, runtime, 1, "owner")
	joinVoteOccupant(t, runtime, 2, "voter")
	joinVoteOccupant(t, runtime, 3, "actor")
	joinVoteOccupant(t, runtime, 4, "eligible")
	store := &fakeStore{mutation: Mutation{Score: 6, Inserted: true}, voters: map[int64]struct{}{2: {}, 3: {}}}
	service := New(store, fakeRooms{room: voteRoom(7, 1, 5)}, runtime, connections, nil)
	if _, err := service.Cast(context.Background(), 7, 3); err != nil {
		t.Fatalf("cast vote: %v", err)
	}
	assertVotePacket(t, *ownerPackets, false)
	assertVotePacket(t, *voterPackets, false)
	assertVotePacket(t, *actorPackets, false)
	assertVotePacket(t, *eligiblePackets, true)
}

// voteRoom creates one durable room fixture.
func voteRoom(roomID int64, ownerID int64, score int) roommodel.Room {
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: roomID}}, OwnerPlayerID: ownerID, Score: score}
}

// registerVoteConnection registers one packet-capturing connection.
func registerVoteConnection(t *testing.T, registry *netconn.Registry, id netconn.ID) *[]codec.Packet {
	t.Helper()
	packets := make([]codec.Packet, 0, 1)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: id, Kind: "websocket", Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error {
		packets = append(packets, packet)
		return nil
	}, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}
	if err := registry.Register(session); err != nil {
		t.Fatalf("register connection: %v", err)
	}

	return &packets
}

// joinVoteOccupant joins one test room occupant.
func joinVoteOccupant(t *testing.T, runtime *roomlive.Registry, playerID int64, connectionID netconn.ID) {
	t.Helper()
	_, err := runtime.Join(context.Background(), 7, roomlive.Occupant{PlayerID: playerID, Username: string(connectionID), ConnectionID: connectionID, ConnectionKind: "websocket"})
	if err != nil {
		t.Fatalf("join occupant: %v", err)
	}
}

// assertVotePacket verifies projected score eligibility.
func assertVotePacket(t *testing.T, packets []codec.Packet, canVote bool) {
	t.Helper()
	if len(packets) != 1 || packets[0].Header != outscore.Header {
		t.Fatalf("unexpected packets: %+v", packets)
	}
	values, err := codec.DecodePacketExact(packets[0], outscore.Definition)
	if err != nil || values[0].Int32 != 6 || values[1].Boolean != canVote {
		t.Fatalf("score packet values=%+v err=%v", values, err)
	}
}
