package routes

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/networking/codec"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
)

// TestReadRoutesReturnWalletAndTypes verifies currency administration reads.
func TestReadRoutesReturnWalletAndTypes(t *testing.T) {
	fixture := newRouteFixture()
	wallet := requestRoute(t, fixture.app, http.MethodGet, "/api/admin/players/7/currencies", "")
	if wallet.StatusCode != fiber.StatusOK || !bodyContains(t, wallet, `"currencyType":-1`) {
		t.Fatalf("unexpected wallet status %d", wallet.StatusCode)
	}
	types := requestRoute(t, fixture.app, http.MethodGet, "/api/admin/currencies/types", "")
	if types.StatusCode != fiber.StatusOK || !bodyContains(t, types, `"key":"diamonds"`) {
		t.Fatalf("unexpected types status %d", types.StatusCode)
	}
}

// TestGrantDefaultsAlertToFalse verifies admin mutations do not alert implicitly.
func TestGrantDefaultsAlertToFalse(t *testing.T) {
	fixture := newRouteFixture()
	connection := addLivePlayer(t, fixture)
	response := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/grant", `{"amount":5}`)

	if response.StatusCode != fiber.StatusOK || len(connection.sent) != 0 {
		t.Fatalf("unexpected status=%d packets=%#v", response.StatusCode, connection.sent)
	}
	if fixture.manager.grant.Amount != 5 || fixture.manager.grant.ActorKind != currencyservice.ActorAdmin {
		t.Fatalf("unexpected grant %#v", fixture.manager.grant)
	}
	if !bodyContains(t, response, `"alertRequested":false`) {
		t.Fatal("expected alert to default false")
	}
}

// TestGrantSendsLocalizedAlertWhenRequested verifies explicit localized delivery.
func TestGrantSendsLocalizedAlertWhenRequested(t *testing.T) {
	fixture := newRouteFixture()
	connection := addLivePlayer(t, fixture)
	response := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/grant", `{"amount":5,"alert":true,"locale":"es"}`)

	if response.StatusCode != fiber.StatusOK || len(connection.sent) != 1 || connection.sent[0].Header != outalert.Header {
		t.Fatalf("unexpected status=%d packets=%#v", response.StatusCode, connection.sent)
	}
	values, err := codec.DecodePacketExact(connection.sent[0], outalert.Definition)
	if err != nil || values[0].String != "Recibiste 5 Diamantes." {
		t.Fatalf("unexpected localized alert values=%#v err=%v", values, err)
	}
	if !bodyContains(t, response, `"alertSent":true`) {
		t.Fatal("expected delivered alert")
	}
}

// TestMutationErrorsUseMeaningfulStatuses verifies validation and domain error mapping.
func TestMutationErrorsUseMeaningfulStatuses(t *testing.T) {
	fixture := newRouteFixture()
	fixture.manager.err = currencyservice.ErrInsufficientBalance
	conflict := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/deduct", `{"amount":20}`)
	if conflict.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected conflict, got %d", conflict.StatusCode)
	}
	invalid := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/grant", `{"amount":0}`)
	if invalid.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", invalid.StatusCode)
	}
	missing := requestRoute(t, fixture.app, http.MethodGet, "/api/admin/players/8/currencies", "")
	if missing.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected not found, got %d", missing.StatusCode)
	}
}

// TestDeductAndSetMapAdministrativeIntent verifies action-specific service parameters.
func TestDeductAndSetMapAdministrativeIntent(t *testing.T) {
	fixture := newRouteFixture()
	deduct := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/deduct", `{"amount":3,"reason":"moderation"}`)
	if deduct.StatusCode != fiber.StatusOK || fixture.manager.grant.Amount != -3 || fixture.manager.grant.Reason != "moderation" {
		t.Fatalf("unexpected deduct status=%d params=%#v", deduct.StatusCode, fixture.manager.grant)
	}

	set := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/set", `{"amount":0}`)
	if set.StatusCode != fiber.StatusOK || fixture.manager.set.Amount != 0 || fixture.manager.set.Reason != "admin_api_set" {
		t.Fatalf("unexpected set status=%d params=%#v", set.StatusCode, fixture.manager.set)
	}
}

// TestAlertRequestedForOfflinePlayerCommitsWithoutDelivery verifies optional side-effect semantics.
func TestAlertRequestedForOfflinePlayerCommitsWithoutDelivery(t *testing.T) {
	fixture := newRouteFixture()
	response := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/grant", `{"amount":5,"alert":true}`)
	if response.StatusCode != fiber.StatusOK || !bodyContains(t, response, `"alertSent":false`) {
		t.Fatalf("unexpected offline alert status %d", response.StatusCode)
	}
}

// TestMutationInputRejectsMalformedRequests verifies route parsing boundaries.
func TestMutationInputRejectsMalformedRequests(t *testing.T) {
	fixture := newRouteFixture()
	cases := []string{
		"/api/admin/players/invalid/currencies/5/grant",
		"/api/admin/players/7/currencies/invalid/grant",
	}
	for _, path := range cases {
		response := requestRoute(t, fixture.app, http.MethodPost, path, `{"amount":1}`)
		if response.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected bad request for %s, got %d", path, response.StatusCode)
		}
	}
	malformed := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/grant", "{")
	if malformed.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected malformed body error, got %d", malformed.StatusCode)
	}
	negativeSet := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/set", `{"amount":-1}`)
	if negativeSet.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected negative set error, got %d", negativeSet.StatusCode)
	}
}

// TestMutationMapsRemainingDomainErrors verifies stable validation responses.
func TestMutationMapsRemainingDomainErrors(t *testing.T) {
	fixture := newRouteFixture()
	cases := []error{
		currencyservice.ErrInvalidCurrencyType,
		currencyservice.ErrInvalidAmount,
		currencyservice.ErrInvalidActor,
	}
	for _, expected := range cases {
		fixture.manager.err = expected
		response := requestRoute(t, fixture.app, http.MethodPost, "/api/admin/players/7/currencies/5/grant", `{"amount":1}`)
		if response.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected bad request for %v, got %d", expected, response.StatusCode)
		}
	}
}

// bodyContains reads and searches one response body.
func bodyContains(t *testing.T, response *http.Response, value string) bool {
	t.Helper()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return strings.Contains(string(data), value)
}
