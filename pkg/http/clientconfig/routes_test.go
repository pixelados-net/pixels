package clientconfig

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestRoutesServeCurrencyConfigAndLocalizedTexts verifies Nitro extension resources.
func TestRoutesServeCurrencyConfigAndLocalizedTexts(t *testing.T) {
	app := fiber.New()
	Register(app, fakeReader{}, i18n.NewCatalog(i18n.Config{DefaultLocale: "es"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {"currency.name.credits": "Créditos", "currency.name.diamonds": "Diamantes"},
	}))

	config := requestBody(t, app, UIConfigPath)
	if !strings.Contains(config, `"system.currency.types":[-1,5]`) {
		t.Fatalf("unexpected config %s", config)
	}
	texts := requestBody(t, app, "/client/texts/es/ExternalTexts.json")
	if !strings.Contains(texts, `"purse.seasonal.currency.5":"Diamantes"`) {
		t.Fatalf("unexpected texts %s", texts)
	}
}

// requestBody requests one client configuration resource.
func requestBody(t *testing.T, app *fiber.App, path string) string {
	t.Helper()
	response, err := app.Test(newRequest(path))
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if response.StatusCode != http.StatusOK || response.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("unexpected response status=%d headers=%v", response.StatusCode, response.Header)
	}

	return string(data)
}

// newRequest creates a client configuration request.
func newRequest(path string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, path, nil)
	return request
}

// fakeReader returns configured client currency types.
type fakeReader struct{}

// Wallet returns no balances.
func (fakeReader) Wallet(context.Context, int64) ([]currencymodel.Balance, error) { return nil, nil }

// Balance returns a zero balance.
func (fakeReader) Balance(context.Context, int64, int32) (int64, error) { return 0, nil }

// Types returns test currency definitions.
func (fakeReader) Types(context.Context) ([]currencymodel.Definition, error) {
	return []currencymodel.Definition{{Type: -1, Key: "credits"}, {Type: 5, Key: "diamonds"}}, nil
}
