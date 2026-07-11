package votes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// fakeVotes records command mutations.
type fakeVotes struct {
	// roomID stores the cast room.
	roomID int64
	// playerID stores the cast player.
	playerID int64
}

// Cast records one vote.
func (votes *fakeVotes) Cast(_ context.Context, roomID int64, playerID int64) (roomvotes.Mutation, error) {
	votes.roomID = roomID
	votes.playerID = playerID
	return roomvotes.Mutation{Score: 1, Inserted: true}, nil
}

// State returns empty state.
func (*fakeVotes) State(context.Context, int64, int64) (roomvotes.State, error) {
	return roomvotes.State{}, nil
}

// List returns no votes.
func (*fakeVotes) List(context.Context, roomvotes.Query) ([]roomvotes.Vote, error) { return nil, nil }

// TestCastHandlerVotesForCurrentRoom verifies session-derived room targeting.
func TestCastHandlerVotesForCurrentRoom(t *testing.T) {
	handler, connection, votes := castFixture(t)
	err := handler.Handle(context.Background(), command.Envelope[CastCommand]{Command: CastCommand{Handler: connection, Rating: PositiveRating}})
	if err != nil || votes.roomID != 9 || votes.playerID != 2 {
		t.Fatalf("room=%d player=%d err=%v", votes.roomID, votes.playerID, err)
	}
}

// TestCastHandlerRejectsUnsupportedRatings verifies semantic packet validation.
func TestCastHandlerRejectsUnsupportedRatings(t *testing.T) {
	err := (CastHandler{}).Handle(context.Background(), command.Envelope[CastCommand]{Command: CastCommand{Rating: 0}})
	if !errors.Is(err, ErrInvalidRating) {
		t.Fatalf("error=%v", err)
	}
}

// TestCastCommandNameReturnsStableName verifies command routing identity.
func TestCastCommandNameReturnsStableName(t *testing.T) {
	if name := (CastCommand{}).CommandName(); name != CastName {
		t.Fatalf("name=%q", name)
	}
}

// castFixture creates one bound room actor.
func castFixture(t *testing.T) (CastHandler, netconn.Context, *fakeVotes) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 2, Username: "Alice"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 2, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	votes := &fakeVotes{}

	return CastHandler{Players: players, Bindings: bindings, Votes: votes}, netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, votes
}
