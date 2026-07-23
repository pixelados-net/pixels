package send

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outtalk "github.com/niflaot/pixels/networking/outbound/chat/talk"
)

// counterForTest stores one configurable flood count.
type counterForTest struct{ count int64 }

// Increment returns the configured count.
func (counter *counterForTest) Increment(context.Context, string, time.Duration) (int64, error) {
	return counter.count, nil
}

// permissionsForTest stores configurable bypass capabilities.
type permissionsForTest struct{ allowed map[permission.Node]bool }

// HasPermission denies one capability.
func (permissions permissionsForTest) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return permissions.allowed[node], nil
}

// globalStoreForTest stores immutable words.
type globalStoreForTest struct{ words []string }

// List returns global words.
func (store globalStoreForTest) List(context.Context) ([]string, error) { return store.words, nil }

// Add accepts one word.
func (globalStoreForTest) Add(context.Context, string) error { return nil }

// Remove accepts one word.
func (globalStoreForTest) Remove(context.Context, string) error { return nil }

// roomFilterForTest applies one deterministic local replacement.
type roomFilterForTest struct{}

// List returns no admin words.
func (roomFilterForTest) List(context.Context, int64) ([]string, error) { return nil, nil }

// Add accepts one room word.
func (roomFilterForTest) Add(context.Context, int64, int64, string) error { return nil }

// Remove accepts one room word.
func (roomFilterForTest) Remove(context.Context, int64, int64, string) error { return nil }

// Contains reports one local word.
func (roomFilterForTest) Contains(_ context.Context, _ int64, text string) (bool, error) {
	return strings.Contains(text, "ugly"), nil
}

// Censor replaces one local word.
func (roomFilterForTest) Censor(_ context.Context, _ int64, text string) (string, bool, error) {
	changed := strings.Contains(text, "ugly")
	return strings.ReplaceAll(text, "ugly", "****"), changed, nil
}

// TestTalkPipelineFiltersAndBroadcasts verifies the full live talk path.
func TestTalkPipelineFiltersAndBroadcasts(t *testing.T) {
	fixture := newFixture(t)
	if err := fixture.source.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive talk: %v", err)
	}
	if len(*fixture.sourcePackets) != 1 || len(*fixture.targetPackets) != 1 {
		t.Fatalf("source=%d target=%d", len(*fixture.sourcePackets), len(*fixture.targetPackets))
	}
	packet := (*fixture.targetPackets)[0]
	values, err := codec.DecodePacketExact(packet, outtalk.Definition)
	if err != nil || packet.Header != outtalk.Header || values[1].String != "*** and ****" {
		t.Fatalf("packet=%#v values=%#v err=%v", packet, values, err)
	}
}

// TestTalkSkipsRecipientsIgnoringSpeaker verifies directional ignore filtering.
func TestTalkSkipsRecipientsIgnoringSpeaker(t *testing.T) {
	fixture := newFixture(t)
	target, found := fixture.players.Find(2)
	if !found {
		t.Fatal("missing target player")
	}
	target.Ignore(1)
	if err := fixture.source.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive talk: %v", err)
	}
	if len(*fixture.sourcePackets) != 1 || len(*fixture.targetPackets) != 0 {
		t.Fatalf("source=%d target=%d", len(*fixture.sourcePackets), len(*fixture.targetPackets))
	}
}

// fixture contains one active two-player room chat setup.
type fixture struct {
	// players stores live players for extending the fixture.
	players *playerlive.Registry
	// bindings stores session bindings for extending the fixture.
	bindings *binding.Registry
	// connections stores sessions for extending the fixture.
	connections *netconn.Registry
	// runtime stores active rooms for extending the fixture.
	runtime *roomlive.Registry
	// source stores the speaking session.
	source netconn.Connection
	// sourcePackets stores packets delivered to the speaker.
	sourcePackets *[]codec.Packet
	// targetPackets stores packets delivered to the listener.
	targetPackets *[]codec.Packet
	// active stores the loaded room.
	active *roomlive.Room
	// counter stores flood state.
	counter *counterForTest
	// request stores the next source chat request.
	request *Request
	// context stores the captured source handler context.
	context *netconn.Context
	// service stores the composed chat behavior.
	service *Service
	// permissions stores mutable test bypass decisions.
	permissions permissionsForTest
}

// newFixture creates one two-player room and chat service.
func newFixture(t *testing.T) fixture {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 99, MaxUsers: 25, ChatDistance: 50, ChatProtection: 1})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	loadWorld(t, active)
	global := chatfilter.New(globalStoreForTest{words: []string{"bad"}})
	if err = global.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh global filter: %v", err)
	}
	counter := &counterForTest{count: 1}
	var service *Service
	request := Request{Kind: KindTalk, Message: "bad and ugly"}
	var sourceContext netconn.Context
	sourcePackets := make([]codec.Packet, 0, 2)
	source := registerSession(t, connections, "source", &sourcePackets, func(connection netconn.Context) error {
		sourceContext = connection
		value := request
		value.Handler = connection
		return service.Handle(context.Background(), value)
	})
	targetPackets := make([]codec.Packet, 0, 1)
	registerSession(t, connections, "target", &targetPackets, nil)
	addPlayer(t, players, bindings, runtime, active, 1, "source", "alice")
	addPlayer(t, players, bindings, runtime, active, 2, "target", "bob")
	permissions := permissionsForTest{allowed: make(map[permission.Node]bool)}
	service = New(chatconfig.Config{}, players, bindings, runtime, connections, permissions, counter, global, roomFilterForTest{}, nil, nil, Nodes{
		FloodImmune: "flood", LengthUnlimited: "length", FilterImmune: "filter",
	})
	t.Cleanup(func() { active.Close() })

	return fixture{
		players: players, bindings: bindings, connections: connections, runtime: runtime,
		source: source, sourcePackets: &sourcePackets, targetPackets: &targetPackets,
		active: active, counter: counter, request: &request, context: &sourceContext, service: service,
		permissions: permissions,
	}
}

// registerSession creates and registers one captured connection.
func registerSession(t *testing.T, registry *netconn.Registry, id netconn.ID, sent *[]codec.Packet, receive func(netconn.Context) error) netconn.Connection {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	if receive != nil {
		_ = inbound.Register(1, func(connection netconn.Context, _ codec.Packet) error { return receive(connection) }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	}
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: id, Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { *sent = append(*sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err = registry.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	return session
}

// addPlayer adds one authenticated player and room occupant.
func addPlayer(t *testing.T, players *playerlive.Registry, bindings *binding.Registry, runtime *roomlive.Registry, active *roomlive.Room, playerID int64, connectionID netconn.ID, username string) {
	t.Helper()
	peer, _ := playerlive.NewSessionPeer(connectionID, "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: playerID, Username: username, Gender: playermodel.GenderMale}, peer)
	_ = players.Add(player)
	_ = bindings.Add(binding.Binding{PlayerID: playerID, ConnectionID: connectionID, ConnectionKind: "websocket"})
	_, _ = runtime.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: playerID, Username: username, ConnectionID: connectionID, ConnectionKind: "websocket"})
	_ = player.EnterRoom(active.ID())
}

// loadWorld loads a small walkable room for unit projection.
func loadWorld(t *testing.T, active *roomlive.Room) {
	t.Helper()
	roomGrid, _ := grid.Parse("000\r000\r000", grid.WithDoor(0, 0))
	err := active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.Point{}}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth})
	if err != nil {
		t.Fatalf("load world: %v", err)
	}
}
