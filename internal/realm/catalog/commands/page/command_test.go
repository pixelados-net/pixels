package page

import (
	"context"
	"errors"
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
	outpage "github.com/niflaot/pixels/networking/outbound/catalog/page"
	"github.com/niflaot/pixels/pkg/i18n"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap/zapcore"
)

// TestHandleSendsCatalogPage verifies offer and furniture projection.
func TestHandleSendsCatalogPage(t *testing.T) {
	connection, sent := pageConnection(t)
	players, bindings := pagePlayer(t, connection)
	reader := pageReader{
		page: catalogmodel.Page{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, Name: "groups", Layout: "guild_frontpage"},
		items: []catalogmodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 5}}, PageID: 2, DefinitionID: 9,
			Name: "chair", CostCredits: 2, PointsType: -1, Amount: 1, Enabled: true}},
		definition: furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, SpriteID: 39, Kind: furnituremodel.KindFloor},
	}
	handler := Handler{Players: players, Bindings: bindings, Catalog: reader, Translations: i18n.NewCatalog(i18n.Config{}, nil)}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: -1, Mode: "NORMAL"}})
	if err != nil || len(*sent) != 1 || (*sent)[0].Header != outpage.Header {
		t.Fatalf("unexpected packets %#v error %v", *sent, err)
	}
	values, _, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}, (*sent)[0].Payload)
	if err != nil || values[2].String != "guild_frontpage" {
		t.Fatalf("unexpected catalog page layout %#v error %v", values, err)
	}
	player, _ := players.Find(7)
	viewer, _ := player.Catalog()
	pageID, found := viewer.CurrentPage()
	if !found || pageID != 2 {
		t.Fatalf("unexpected current page %d found=%t", pageID, found)
	}
}

// TestHandleSendsPureEffectOffer verifies pages do not resolve furniture id zero.
func TestHandleSendsPureEffectOffer(t *testing.T) {
	connection, sent := pageConnection(t)
	players, bindings := pagePlayer(t, connection)
	effectID := int32(101)
	reader := pageReader{
		page: catalogmodel.Page{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, Name: "avatar_effects", Layout: "default_3x3"},
		items: []catalogmodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 5}}, PageID: 2,
			Name: "effect_confetti", CostCredits: 2, PointsType: -1, GrantsEffectID: &effectID, Enabled: true}},
	}
	handler := Handler{Players: players, Bindings: bindings, Catalog: reader, Translations: i18n.NewCatalog(i18n.Config{}, nil)}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: -1, Mode: "NORMAL"}})
	if err != nil || len(*sent) != 1 || (*sent)[0].Header != outpage.Header {
		t.Fatalf("unexpected packets %#v error %v", *sent, err)
	}
}

// TestCommandMetadataAndMissingDefinition verifies metadata and mapping failures.
func TestCommandMetadataAndMissingDefinition(t *testing.T) {
	connection, _ := pageConnection(t)
	players, bindings := pagePlayer(t, connection)
	input := Command{Connection: connection, PageID: 2, OfferID: -1, Mode: "NORMAL"}
	encoder := zapcore.NewMapObjectEncoder()
	if input.CommandName() != Name || input.MarshalLogObject(encoder) != nil || encoder.Fields["page_id"] != int64(2) {
		t.Fatalf("unexpected command metadata %#v", encoder.Fields)
	}
	reader := pageReader{page: catalogmodel.Page{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}},
		items: []catalogmodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 5}}, DefinitionID: 99}}, found: false}
	handler := Handler{Players: players, Bindings: bindings, Catalog: reader}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); err == nil {
		t.Fatal("expected missing definition error")
	}
	expected := errors.New("page unavailable")
	reader.err = expected
	handler.Catalog = reader
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); !errors.Is(err, expected) {
		t.Fatalf("expected page error, got %v", err)
	}
}

// pageReader supplies catalog page and offer fixtures.
type pageReader struct {
	// page stores a catalog page fixture.
	page catalogmodel.Page
	// items stores offer fixtures.
	items []catalogmodel.Item
	// definition stores furniture metadata.
	definition furnituremodel.Definition
	// found reports whether furniture metadata exists.
	found bool
	// err stores an optional page read failure.
	err error
}

// Pages supplies unused page tree behavior.
func (pageReader) Pages(context.Context, int64, bool) ([]catalogmodel.Page, error) { return nil, nil }

// Page returns catalog page fixtures.
func (reader pageReader) Page(context.Context, int64, int64, bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	return reader.page, reader.items, reader.err
}

// Definition returns furniture metadata.
func (reader pageReader) Definition(context.Context, int64) (furnituremodel.Definition, bool, error) {
	found := reader.found
	if reader.definition.ID > 0 {
		found = true
	}
	return reader.definition, found, nil
}

// SanitizeList supplies unused sanitize behavior.
func (pageReader) SanitizeList(context.Context) ([]furnituremodel.Definition, error) { return nil, nil }

// pageReaderAssertion verifies pageReader implements Reader.
var pageReaderAssertion catalogservice.Reader = pageReader{}

// pagePlayer creates a bound live player fixture.
func pagePlayer(t *testing.T, connection netconn.Context) (*playerlive.Registry, *binding.Registry) {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	peer, _ := playerlive.NewSessionPeer(connection.ConnectionID, connection.ConnectionKind, time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	_ = players.Add(player)
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})

	return players, bindings
}

// pageConnection captures a functional connection context and sent packets.
func pageConnection(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	var captured netconn.Context
	inbound := netconn.NewHandlerRegistry()
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { captured = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "catalog", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})

	return captured, &sent
}
