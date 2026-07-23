package routes

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestNotifySendsLocalizedBubble verifies player bubble delivery.
func TestNotifySendsLocalizedBubble(t *testing.T) {
	sent := make([]codec.Packet, 0, 1)
	app, players, connections := testApp(t, &sent)
	addPlayer(t, players, connections, "connection-one", &sent)

	body := `{"playerId":7,"kind":"bubble","key":"admin.notification.default","locale":"es"}`
	response := request(t, app, http.MethodPost, "/api/admin/notifications/send", body)
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	if len(sent) != 1 || sent[0].Header != outbubble.Header {
		t.Fatalf("expected one bubble packet, got %#v", sent)
	}
}

// TestNotifyReportsOfflinePlayer verifies missing live players fail.
func TestNotifyReportsOfflinePlayer(t *testing.T) {
	sent := make([]codec.Packet, 0, 1)
	app, _, _ := testApp(t, &sent)

	body := `{"playerId":7,"kind":"bubble","key":"admin.notification.default"}`
	response := request(t, app, http.MethodPost, "/api/admin/notifications/send", body)
	if response.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.StatusCode)
	}
}

// TestNotifyRejectsMissingTarget verifies notification input requires a player id.
func TestNotifyRejectsMissingTarget(t *testing.T) {
	sent := make([]codec.Packet, 0, 1)
	app, _, _ := testApp(t, &sent)
	response := request(t, app, http.MethodPost, "/api/admin/notifications/send", `{"kind":"bubble","key":"admin.notification.default"}`)
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}

// TestNotifySendsLocalizedAlert verifies the explicit alert packet path.
func TestNotifySendsLocalizedAlert(t *testing.T) {
	sent := make([]codec.Packet, 0, 1)
	app, players, connections := testApp(t, &sent)
	addPlayer(t, players, connections, "connection-alert", &sent)

	body := `{"playerId":7,"kind":"alert","key":"admin.notification.default"}`
	response := request(t, app, http.MethodPost, "/api/admin/notifications/send", body)
	if response.StatusCode != fiber.StatusOK || len(sent) != 1 || sent[0].Header != outalert.Header {
		t.Fatalf("expected one alert packet, status=%d packets=%#v", response.StatusCode, sent)
	}
}

// TestNotifyRejectsInvalidRequests verifies malformed notification boundaries.
func TestNotifyRejectsInvalidRequests(t *testing.T) {
	sent := make([]codec.Packet, 0, 1)
	app, players, connections := testApp(t, &sent)
	addPlayer(t, players, connections, "connection-invalid", &sent)

	for _, body := range []string{
		"{",
		`{"playerId":7,"kind":"bubble"}`,
		`{"playerId":7,"kind":"unsupported","key":"admin.notification.default"}`,
	} {
		response := request(t, app, http.MethodPost, "/api/admin/notifications/send", body)
		if response.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected status 400 for %s, got %d", body, response.StatusCode)
		}
	}
}

// testApp creates a route test app.
func testApp(t *testing.T, sent *[]codec.Packet) (*fiber.App, *playerlive.Registry, *netconn.Registry) {
	t.Helper()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	players := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	translations := i18n.NewCatalog(i18n.Config{DefaultLocale: "es"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {"admin.notification.default": "Mensaje del hotel."},
	})
	Register(app, players, connections, translations)

	return app, players, connections
}

// addPlayer registers a live player and connection.
func addPlayer(t *testing.T, players *playerlive.Registry, connections *netconn.Registry, id netconn.ID, sent *[]codec.Packet) {
	t.Helper()

	connection := &testConnection{id: id, sent: sent, done: make(chan struct{})}
	if err := connections.Register(connection); err != nil {
		t.Fatalf("register connection: %v", err)
	}

	peer, err := playerlive.NewSessionPeer(id, "websocket", time.Now())
	if err != nil {
		t.Fatalf("new peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo", Gender: playermodel.GenderMale}, peer)
	if err != nil {
		t.Fatalf("new player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
}

// testConnection captures sent packets.
type testConnection struct {
	// id stores the connection id.
	id netconn.ID
	// sent stores captured outbound packets.
	sent *[]codec.Packet
	// done closes on disconnect.
	done chan struct{}
}

// ID returns the connection identifier.
func (connection *testConnection) ID() netconn.ID { return connection.id }

// Kind returns the connection kind.
func (connection *testConnection) Kind() netconn.Kind { return "websocket" }

// StartedAt returns the connection start time.
func (connection *testConnection) StartedAt() time.Time { return time.Now() }

// AuthenticatedAt returns the authentication time.
func (connection *testConnection) AuthenticatedAt() (time.Time, bool) { return time.Now(), true }

// Authenticate marks the connection as authenticated.
func (connection *testConnection) Authenticate(time.Time) error { return nil }

// State returns the lifecycle state.
func (connection *testConnection) State() netconn.State { return netconn.StateConnected }

// Receive handles an inbound packet.
func (connection *testConnection) Receive(context.Context, codec.Packet) error { return nil }

// Send captures an outbound packet.
func (connection *testConnection) Send(_ context.Context, packet codec.Packet) error {
	*connection.sent = append(*connection.sent, packet)

	return nil
}

// Disconnect disposes the connection.
func (connection *testConnection) Disconnect(context.Context, netconn.Reason) error {
	close(connection.done)

	return nil
}

// Done returns the disposal channel.
func (connection *testConnection) Done() <-chan struct{} { return connection.done }

// request executes one HTTP request.
func request(t *testing.T, app *fiber.App, method string, path string, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}

	return response
}
