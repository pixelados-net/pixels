package projection

import (
	"context"
	"errors"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// activeReaderForTest supplies one aggregate sanction projection.
type activeReaderForTest struct {
	// state stores the returned aggregate projection.
	state sanctionrecord.ActiveState
	// err optionally fails the read.
	err error
}

// Active returns configured test state.
func (reader *activeReaderForTest) Active(context.Context, int64) (sanctionrecord.ActiveState, error) {
	return reader.state, reader.err
}

// tradeManagerForTest records compatibility-column changes.
type tradeManagerForTest struct {
	// allow stores the latest compatibility state.
	allow bool
	// err optionally fails persistence.
	err error
}

// SetAllowTrade records the requested state.
func (manager *tradeManagerForTest) SetAllowTrade(_ context.Context, _ int64, boolValue bool) error {
	manager.allow = boolValue
	return manager.err
}

// livePlayerForTest creates one registry-backed authenticated player.
func livePlayerForTest(t *testing.T, id int64) (*playerlive.Registry, *playerlive.Player) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("test", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: id, Username: "test", AllowTrade: true}, peer)
	if err != nil {
		t.Fatal(err)
	}
	registry := playerlive.NewRegistry()
	if err = registry.Add(player); err != nil {
		t.Fatal(err)
	}
	return registry, player
}

// TestMuteRefreshesAggregateProjection verifies overlapping state is copied, not toggled.
func TestMuteRefreshesAggregateProjection(t *testing.T) {
	registry, player := livePlayerForTest(t, 7)
	expires := time.Now().Add(time.Hour)
	reader := &activeReaderForTest{state: sanctionrecord.ActiveState{MuteUntil: &expires}}
	effect := &Mute{sanctions: reader, players: registry}
	punishment := sanctionrecord.Punishment{ReceiverPlayerID: 7}
	if err := effect.Apply(context.Background(), punishment); err != nil {
		t.Fatal(err)
	}
	if player.Snapshot().Sanctions.MuteUntil != expires {
		t.Fatalf("sanctions=%+v", player.Snapshot().Sanctions)
	}
	reader.state = sanctionrecord.ActiveState{MutedPermanently: true}
	if err := effect.Revoke(context.Background(), punishment); err != nil {
		t.Fatal(err)
	}
	if !player.Snapshot().Sanctions.MutePermanent || !player.Snapshot().Sanctions.MuteUntil.IsZero() {
		t.Fatalf("sanctions=%+v", player.Snapshot().Sanctions)
	}
}

// TestTradeLockRefreshesDurableAndLiveProjection verifies one aggregate writer.
func TestTradeLockRefreshesDurableAndLiveProjection(t *testing.T) {
	registry, player := livePlayerForTest(t, 7)
	reader := &activeReaderForTest{state: sanctionrecord.ActiveState{TradeLockedPermanently: true}}
	writer := &tradeManagerForTest{allow: true}
	effect := &TradeLock{sanctions: reader, players: writer, live: registry}
	punishment := sanctionrecord.Punishment{ReceiverPlayerID: 7}
	if err := effect.Apply(context.Background(), punishment); err != nil {
		t.Fatal(err)
	}
	if writer.allow || player.Snapshot().AllowTrade || !player.Snapshot().Sanctions.TradeLockPermanent {
		t.Fatalf("allow=%v snapshot=%+v", writer.allow, player.Snapshot())
	}
	reader.state = sanctionrecord.ActiveState{}
	if err := effect.Revoke(context.Background(), punishment); err != nil {
		t.Fatal(err)
	}
	if !writer.allow || !player.Snapshot().AllowTrade || player.Snapshot().Sanctions.TradeLockPermanent {
		t.Fatalf("allow=%v snapshot=%+v", writer.allow, player.Snapshot())
	}
}

// TestProjectionErrorsDoNotMutateLiveState verifies failed aggregate reads stop effects.
func TestProjectionErrorsDoNotMutateLiveState(t *testing.T) {
	registry, player := livePlayerForTest(t, 7)
	expected := errors.New("read failed")
	effect := &Mute{sanctions: &activeReaderForTest{err: expected}, players: registry}
	if err := effect.Apply(context.Background(), sanctionrecord.Punishment{ReceiverPlayerID: 7}); !errors.Is(err, expected) {
		t.Fatalf("err=%v", err)
	}
	if player.Snapshot().Sanctions.MutePermanent {
		t.Fatal("failed read mutated live state")
	}
}

// TestProjectionConstructorsExposeKinds verifies behavior registry identities.
func TestProjectionConstructorsExposeKinds(t *testing.T) {
	registry := playerlive.NewRegistry()
	writer := &tradeManagerForTest{}
	if NewMute(nil, registry).Kind() != sanctionrecord.KindMute || NewTradeLock(nil, writer, registry).Kind() != sanctionrecord.KindTradeLock {
		t.Fatal("projection effect kind mismatch")
	}
}
