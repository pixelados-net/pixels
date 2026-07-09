package routes

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// routeFixture contains currency route test collaborators.
type routeFixture struct {
	// app serves currency routes.
	app *fiber.App
	// manager records currency calls.
	manager *fakeManager
	// players stores live test players.
	players *playerlive.Registry
	// connections stores test connections.
	connections *netconn.Registry
}

// newRouteFixture creates currency route test collaborators.
func newRouteFixture() routeFixture {
	app := fiber.New()
	manager := &fakeManager{
		wallet: []currencymodel.Balance{{PlayerID: 7, CurrencyType: -1, Amount: 10}},
		types:  []currencymodel.Definition{{Type: -1, Key: "credits", Ledger: true}, {Type: 5, Key: "diamonds"}},
		result: 15,
	}
	players := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	translations := i18n.NewCatalog(i18n.Config{DefaultLocale: "es"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {
			"currency.name.credits":       "Créditos",
			"currency.name.diamonds":      "Diamantes",
			"admin.currency.alert.grant":  "Recibiste {amount} {currency}.",
			"admin.currency.alert.deduct": "Se descontaron {amount} {currency}.",
			"admin.currency.alert.set":    "Tu saldo de {currency} ahora es {balance}.",
		},
	})
	Register(app, Dependencies{
		Finder: existingFinder{}, Players: players, Connections: connections,
		Currencies: manager, Translations: translations, Log: zap.NewNop(),
	})

	return routeFixture{app: app, manager: manager, players: players, connections: connections}
}

// requestRoute executes one JSON route request.
func requestRoute(t *testing.T, app *fiber.App, method string, path string, body string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}

	return response
}

// addLivePlayer registers a test player connection.
func addLivePlayer(t *testing.T, fixture routeFixture) *testConnection {
	t.Helper()
	connection := &testConnection{id: "currency-route", done: make(chan struct{})}
	if err := fixture.connections.Register(connection); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	peer, err := playerlive.NewSessionPeer(connection.id, connection.Kind(), time.Now())
	if err != nil {
		t.Fatalf("new peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo", Gender: playermodel.GenderMale}, peer)
	if err != nil {
		t.Fatalf("new player: %v", err)
	}
	if err := fixture.players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return connection
}

// existingFinder resolves only player seven.
type existingFinder struct{}

// FindByID finds a test player.
func (existingFinder) FindByID(_ context.Context, id int64) (playerservice.Record, bool, error) {
	if id != 7 {
		return playerservice.Record{}, false, nil
	}
	return playerservice.Record{Player: playermodel.Player{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}, Username: "demo",
	}}, true, nil
}

// FindByUsername finds no test player.
func (existingFinder) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// fakeManager records currency route operations.
type fakeManager struct {
	// wallet stores wallet results.
	wallet []currencymodel.Balance
	// types stores configured types.
	types []currencymodel.Definition
	// result stores mutation results.
	result int64
	// err stores mutation failures.
	err error
	// grant stores the last grant.
	grant currencyservice.GrantParams
	// set stores the last set.
	set currencyservice.SetParams
}

// Wallet returns fake balances.
func (manager *fakeManager) Wallet(context.Context, int64) ([]currencymodel.Balance, error) {
	return manager.wallet, nil
}

// Balance returns a fake balance.
func (manager *fakeManager) Balance(context.Context, int64, int32) (int64, error) {
	return manager.result, nil
}

// Types returns fake definitions.
func (manager *fakeManager) Types(context.Context) ([]currencymodel.Definition, error) {
	return manager.types, nil
}

// Grant records a signed grant.
func (manager *fakeManager) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	manager.grant = params
	return manager.result, manager.err
}

// Set records an absolute set.
func (manager *fakeManager) Set(_ context.Context, params currencyservice.SetParams) (int64, error) {
	manager.set = params
	return manager.result, manager.err
}

// testConnection records currency route packets.
type testConnection struct {
	// id identifies the connection.
	id netconn.ID
	// sent stores outbound packets.
	sent []codec.Packet
	// done closes on disposal.
	done chan struct{}
}

// ID returns the connection id.
func (connection *testConnection) ID() netconn.ID { return connection.id }

// Kind returns the connection kind.
func (connection *testConnection) Kind() netconn.Kind { return "websocket" }

// StartedAt returns the connection start time.
func (connection *testConnection) StartedAt() time.Time { return time.Now() }

// AuthenticatedAt returns the authentication time.
func (connection *testConnection) AuthenticatedAt() (time.Time, bool) { return time.Now(), true }

// Authenticate marks the connection authenticated.
func (connection *testConnection) Authenticate(time.Time) error { return nil }

// State returns the connection state.
func (connection *testConnection) State() netconn.State { return netconn.StateConnected }

// Receive accepts an inbound packet.
func (connection *testConnection) Receive(context.Context, codec.Packet) error { return nil }

// Send records one outbound packet.
func (connection *testConnection) Send(_ context.Context, packet codec.Packet) error {
	connection.sent = append(connection.sent, packet)
	return nil
}

// Disconnect closes the connection.
func (connection *testConnection) Disconnect(context.Context, netconn.Reason) error {
	close(connection.done)
	return nil
}

// Done returns the disposal signal.
func (connection *testConnection) Done() <-chan struct{} { return connection.done }
