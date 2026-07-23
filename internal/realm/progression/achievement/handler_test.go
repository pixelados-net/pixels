package achievement

import (
	"context"
	"testing"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlimits "github.com/niflaot/pixels/networking/inbound/progression/achievement/limits"
	inlist "github.com/niflaot/pixels/networking/inbound/user/achievement/list"
	outlimits "github.com/niflaot/pixels/networking/outbound/progression/achievement/limits"
	outlist "github.com/niflaot/pixels/networking/outbound/user/achievement/list"
)

// achievementCatalogSource returns one fixed achievement catalog.
type achievementCatalogSource struct{ value progressionrecord.Catalog }

// Catalog returns the fixed achievement catalog.
func (source achievementCatalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return source.value, nil
}

// achievementStore implements the focused player progress read.
type achievementStore struct {
	progressionrecord.Store
	values []progressionrecord.PlayerAchievement
}

// PlayerAchievements returns the fixed player progress.
func (store achievementStore) PlayerAchievements(context.Context, int64) ([]progressionrecord.PlayerAchievement, error) {
	return store.values, nil
}

// TestAchievementHandlersProjectCatalogAndLimits verifies both Nitro achievement requests.
func TestAchievementHandlersProjectCatalogAndLimits(t *testing.T) {
	definitions := []progressionrecord.AchievementDefinition{
		{ID: 1, Name: "RoomEntry", Category: "explore", Visible: true, Enabled: true, Levels: []progressionrecord.AchievementLevel{{Level: 1, ProgressNeeded: 5}}},
		{ID: 2, Name: "Hidden", Category: "identity", Visible: false, Enabled: true, Levels: []progressionrecord.AchievementLevel{{Level: 1, ProgressNeeded: 1}}},
	}
	catalog := progressionengine.NewCatalog(achievementCatalogSource{value: progressionrecord.Catalog{Achievements: definitions}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "achievement", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	connection, sent := achievementConnection(t)
	handler := Handler{Catalog: catalog, Store: achievementStore{values: []progressionrecord.PlayerAchievement{{PlayerID: 7, DefinitionID: 1, Progress: 3}}}, Bindings: bindings}
	if err := handler.list(connection, codec.Packet{Header: inlist.Header}); err != nil {
		t.Fatal(err)
	}
	if err := handler.limits(connection, codec.Packet{Header: inlimits.Header}); err != nil {
		t.Fatal(err)
	}
	if len(*sent) != 3 || (*sent)[0].Header != outlimits.Header || (*sent)[1].Header != outlist.Header || (*sent)[2].Header != outlimits.Header {
		t.Fatalf("packets=%#v", *sent)
	}
	values, rest, err := codec.DecodePacket((*sent)[0], codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || len(rest) != 0 || values[0].Int32 != 1 || values[1].String != "RoomEntry" || values[2].Int32 != 1 || values[3].Int32 != 1 || values[4].Int32 != 5 {
		t.Fatalf("limits values=%#v rest=%d err=%v", values, len(rest), err)
	}
}

// TestAchievementHandlersRejectMalformedAndMissingState verifies safe adapter boundaries.
func TestAchievementHandlersRejectMalformedAndMissingState(t *testing.T) {
	Register(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	Register(registry, Handler{})
	if registry.Len() != 2 {
		t.Fatalf("handlers=%d want=2", registry.Len())
	}
	connection, _ := achievementConnection(t)
	if err := (Handler{}).list(connection, codec.Packet{Header: inlist.Header, Payload: []byte{1}}); err == nil {
		t.Fatal("malformed list accepted")
	}
	if err := (Handler{}).list(connection, codec.Packet{Header: inlist.Header}); err != nil {
		t.Fatal(err)
	}
	if err := (Handler{}).limits(connection, codec.Packet{Header: inlimits.Header}); err != nil {
		t.Fatal(err)
	}
}

// achievementConnection captures one connection context and outbound packets.
func achievementConnection(t testing.TB) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { connection = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "achievement", Kind: "test", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	return connection, &sent
}
