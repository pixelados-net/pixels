package pages

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outpages "github.com/niflaot/pixels/networking/outbound/catalog/pages"
	"github.com/niflaot/pixels/pkg/i18n"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap/zapcore"
)

// TestHandleSendsVisibleCatalogTree verifies catalog pages command behavior.
func TestHandleSendsVisibleCatalogTree(t *testing.T) {
	connection, sent := pagesConnection(t)
	players, bindings := pagesPlayer(t, connection)
	reader := pagesReader{pages: []catalogmodel.Page{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Name: "furniture", Visible: true, Enabled: true}}}
	handler := Handler{Players: players, Bindings: bindings, Catalog: reader,
		Translations: i18n.NewCatalog(i18n.Config{}, map[i18n.Locale]map[i18n.Key]string{"en": {"catalog.page.furniture": "Furniture"}})}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, Mode: "NORMAL"}})
	if err != nil || len(*sent) != 1 || (*sent)[0].Header != outpages.Header {
		t.Fatalf("unexpected packets %#v error %v", *sent, err)
	}
	player, _ := players.Find(7)
	viewer, found := player.Catalog()
	if !found || viewer.Mode() != "NORMAL" {
		t.Fatal("expected embedded catalog viewer")
	}
}

// TestCommandMetadataAndFailures verifies logging metadata and read failures.
func TestCommandMetadataAndFailures(t *testing.T) {
	connection, _ := pagesConnection(t)
	players, bindings := pagesPlayer(t, connection)
	input := Command{Connection: connection, Mode: "NORMAL"}
	encoder := zapcore.NewMapObjectEncoder()
	if input.CommandName() != Name || input.MarshalLogObject(encoder) != nil || encoder.Fields["mode"] != "NORMAL" {
		t.Fatalf("unexpected command metadata %#v", encoder.Fields)
	}
	expected := errors.New("catalog unavailable")
	handler := Handler{Players: players, Bindings: bindings, Catalog: pagesReader{err: expected}}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); !errors.Is(err, expected) {
		t.Fatalf("expected read error, got %v", err)
	}
	handler.Catalog = pagesReader{pages: []catalogmodel.Page{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: math.MaxInt64}}}}}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); err == nil {
		t.Fatal("expected protocol range error")
	}
}

// pagesReader supplies catalog page fixtures.
type pagesReader struct {
	// pages stores visible page fixtures.
	pages []catalogmodel.Page
	// err stores an optional page read failure.
	err error
}

// Pages returns visible page fixtures.
func (reader pagesReader) Pages(context.Context, int64, bool) ([]catalogmodel.Page, error) {
	return reader.pages, reader.err
}

// Page supplies unused page behavior.
func (pagesReader) Page(context.Context, int64, int64, bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	return catalogmodel.Page{}, nil, nil
}

// Definition supplies unused definition behavior.
func (pagesReader) Definition(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{}, false, nil
}

// SanitizeList supplies unused sanitize behavior.
func (pagesReader) SanitizeList(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}

// readerAssertion verifies pagesReader implements Reader.
var readerAssertion catalogservice.Reader = pagesReader{}

// pagesPlayer creates a bound live player fixture.
func pagesPlayer(t *testing.T, connection netconn.Context) (*playerlive.Registry, *binding.Registry) {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	peer, _ := playerlive.NewSessionPeer(connection.ConnectionID, connection.ConnectionKind, time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	_ = players.Add(player)
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})

	return players, bindings
}

// pagesConnection captures a functional connection context and sent packets.
func pagesConnection(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	var captured netconn.Context
	inbound := netconn.NewHandlerRegistry()
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { captured = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "catalog", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})

	return captured, &sent
}
