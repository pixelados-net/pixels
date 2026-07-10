package routes

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencybroadcast "github.com/niflaot/pixels/internal/realm/inventory/currency/broadcast"
	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outnotification "github.com/niflaot/pixels/networking/outbound/user/currency/notification"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// TestGrantFlowsThroughEventBroadcaster verifies HTTP to service to live packet projection.
func TestGrantFlowsThroughEventBroadcaster(t *testing.T) {
	localBus := bus.New()
	catalog, err := currency.NewCatalog([]currencymodel.Definition{{Type: 5, Key: "diamonds"}}, nil)
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	manager := currencyservice.New(integrationStore{}, catalog, localBus, zap.NewNop())
	players := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	broadcaster := currencybroadcast.New(players, connections)
	subscription, err := localBus.Subscribe(currencychanged.Name, bus.PriorityLow, broadcaster.Handle)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(subscription.Unsubscribe)

	app := fiber.New()
	Register(app, Dependencies{
		Finder: existingFinder{}, Players: players, Connections: connections,
		Currencies: manager, Translations: i18n.NewCatalog(i18n.Config{}, nil), Log: zap.NewNop(),
	})
	fixture := routeFixture{app: app, players: players, connections: connections}
	connection := addLivePlayer(t, fixture)

	response := requestRoute(t, app, http.MethodPost, "/api/admin/currencies/grant", `{"playerId":7,"currencyType":5,"amount":5}`)
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status %d", response.StatusCode)
	}
	if len(connection.sent) != 1 || connection.sent[0].Header != outnotification.Header {
		t.Fatalf("unexpected projected packets %#v", connection.sent)
	}
}

// integrationStore commits deterministic route integration mutations.
type integrationStore struct{}

// FindBalance finds no existing balance.
func (integrationStore) FindBalance(context.Context, int64, int32) (currencymodel.Balance, bool, error) {
	return currencymodel.Balance{}, false, nil
}

// ListBalances lists no existing balances.
func (integrationStore) ListBalances(context.Context, int64) ([]currencymodel.Balance, error) {
	return nil, nil
}

// Grant commits one deterministic delta.
func (integrationStore) Grant(_ context.Context, mutation currencyrepo.Mutation) (currencyrepo.Result, error) {
	return currencyrepo.Result{
		Balance: currencymodel.Balance{PlayerID: mutation.PlayerID, CurrencyType: mutation.CurrencyType, Amount: mutation.Amount},
		Delta:   mutation.Amount,
	}, nil
}

// Set commits one deterministic absolute balance.
func (integrationStore) Set(_ context.Context, mutation currencyrepo.Mutation) (currencyrepo.Result, error) {
	return currencyrepo.Result{
		Balance: currencymodel.Balance{PlayerID: mutation.PlayerID, CurrencyType: mutation.CurrencyType, Amount: mutation.Amount},
		Delta:   mutation.Amount,
	}, nil
}
