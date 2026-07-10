package buy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outsoldout "github.com/niflaot/pixels/networking/outbound/catalog/limited/soldout"
	outfailed "github.com/niflaot/pixels/networking/outbound/catalog/purchase/failed"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outunavailable "github.com/niflaot/pixels/networking/outbound/catalog/purchase/unavailable"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestHandleCompletesPurchaseAndRefreshesInventory verifies successful purchase responses.
func TestHandleCompletesPurchaseAndRefreshesInventory(t *testing.T) {
	handler, connection, sent, manager := buyFixture(t)
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: 5, Amount: 1}})
	if err != nil || manager.purchases != 1 || len(*sent) != 3 ||
		(*sent)[0].Header != outunseen.Header || (*sent)[1].Header != outok.Header || (*sent)[2].Header != outrefresh.Header {
		t.Fatalf("unexpected packets %#v purchases=%d error %v", *sent, manager.purchases, err)
	}
}

// TestCommandMetadataAndUnexpectedFailure verifies command metadata and server failure mapping.
func TestCommandMetadataAndUnexpectedFailure(t *testing.T) {
	handler, connection, sent, manager := buyFixture(t)
	input := Command{Connection: connection, PageID: 2, OfferID: 5, Amount: 1}
	encoder := zapcore.NewMapObjectEncoder()
	if input.CommandName() != Name || input.MarshalLogObject(encoder) != nil || encoder.Fields["offer_id"] != int64(5) {
		t.Fatalf("unexpected command metadata %#v", encoder.Fields)
	}
	manager.err = errors.New("database unavailable")
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: input}); err != nil || len(*sent) != 1 || (*sent)[0].Header != outfailed.Header {
		t.Fatalf("unexpected packets %#v error %v", *sent, err)
	}
}

// TestHandleMapsExpectedPurchaseFailures verifies non-fatal protocol outcomes.
func TestHandleMapsExpectedPurchaseFailures(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		header uint16
	}{
		{name: "insufficient", err: currencyservice.ErrInsufficientBalance, header: outunavailable.Header},
		{name: "sold out", err: catalogservice.ErrLimitedSoldOut, header: outsoldout.Header},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler, connection, sent, manager := buyFixture(t)
			manager.err = test.err
			err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: 5, Amount: 1}})
			if err != nil || len(*sent) != 1 || (*sent)[0].Header != test.header {
				t.Fatalf("unexpected packets %#v error %v", *sent, err)
			}
		})
	}
}

// TestHandleRejectsBulkPurchase verifies deferred bulk purchase policy.
func TestHandleRejectsBulkPurchase(t *testing.T) {
	handler, connection, sent, _ := buyFixture(t)
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: 5, Amount: 2}})
	if err != nil || len(*sent) != 1 || (*sent)[0].Header != outunavailable.Header {
		t.Fatalf("unexpected packets %#v error %v", *sent, err)
	}
}

// TestHandleRejectsMismatchedPageAndMissingDefinition verifies pre-purchase guards.
func TestHandleRejectsMismatchedPageAndMissingDefinition(t *testing.T) {
	handler, connection, sent, manager := buyFixture(t)
	manager.pageItems = []catalogmodel.Item{}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: 5, Amount: 1}})
	if err != nil || manager.purchases != 0 || len(*sent) != 1 || (*sent)[0].Header != outunavailable.Header {
		t.Fatalf("unexpected mismatched page packets %#v error %v", *sent, err)
	}
	handler, connection, sent, manager = buyFixture(t)
	manager.definitionFound = false
	err = handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection, PageID: 2, OfferID: 5, Amount: 1}})
	if err != nil || len(*sent) != 1 || (*sent)[0].Header != outfailed.Header {
		t.Fatalf("unexpected missing definition packets %#v error %v", *sent, err)
	}
}

// buyManager supplies catalog purchase fixtures.
type buyManager struct {
	// item stores the offered catalog item.
	item catalogmodel.Item
	// definition stores furniture metadata.
	definition furnituremodel.Definition
	// err stores an optional purchase failure.
	err error
	// purchases counts purchase calls.
	purchases int
	// pageItems overrides page offer fixtures when non-nil.
	pageItems []catalogmodel.Item
	// definitionFound overrides furniture metadata availability.
	definitionFound bool
}

// Pages supplies unused catalog tree behavior.
func (buyManager) Pages(context.Context, int32, bool) ([]catalogmodel.Page, error) { return nil, nil }

// Page returns the offered item.
func (manager *buyManager) Page(context.Context, int64, int32, bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	if manager.pageItems != nil {
		return catalogmodel.Page{}, manager.pageItems, nil
	}
	return catalogmodel.Page{}, []catalogmodel.Item{manager.item}, nil
}

// Definition returns furniture metadata.
func (manager *buyManager) Definition(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return manager.definition, manager.definitionFound, nil
}

// SanitizeList supplies unused sanitize behavior.
func (buyManager) SanitizeList(context.Context) ([]furnituremodel.Definition, error) { return nil, nil }

// Purchase returns a configured purchase outcome.
func (manager *buyManager) Purchase(context.Context, catalogservice.PurchaseParams) (catalogservice.PurchaseResult, error) {
	manager.purchases++
	if manager.err != nil {
		return catalogservice.PurchaseResult{}, manager.err
	}
	return catalogservice.PurchaseResult{
		Item: manager.item,
		GrantedItems: []furnituremodel.Item{{Base: sharedmodel.Base{
			Identity: sharedmodel.Identity{ID: 41},
		}}},
	}, nil
}

// Refresh supplies unused cache behavior.
func (buyManager) Refresh(context.Context) error { return nil }

// buyManagerAssertion verifies buyManager implements Manager.
var buyManagerAssertion catalogservice.Manager = (*buyManager)(nil)

// buyFixture creates a bound purchase command fixture.
func buyFixture(t *testing.T) (Handler, netconn.Context, *[]codec.Packet, *buyManager) {
	t.Helper()
	connection, sent := buyConnection(t)
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	peer, _ := playerlive.NewSessionPeer(connection.ConnectionID, connection.ConnectionKind, time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	_ = players.Add(player)
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})
	item := catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 5}}, PageID: 2, DefinitionID: 9,
		Name: "chair", CostCredits: 2, PointsType: -1, Amount: 1, Enabled: true}
	manager := &buyManager{item: item, definitionFound: true, definition: furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, SpriteID: 39, Kind: furnituremodel.KindFloor}}

	return Handler{Players: players, Bindings: bindings, Catalog: manager, Log: zap.NewNop()}, connection, sent, manager
}

// buyConnection captures a functional connection context and sent packets.
func buyConnection(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	var captured netconn.Context
	inbound := netconn.NewHandlerRegistry()
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { captured = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "catalog", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}

	return captured, &sent
}
