package clientconfig

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestRoutesServeCurrencyConfigAndLocalizedTexts verifies Nitro extension resources.
func TestRoutesServeCurrencyConfigAndLocalizedTexts(t *testing.T) {
	app := fiber.New()
	Register(app, fakeReader{}, fakeLayouts{}, i18n.NewCatalog(i18n.Config{DefaultLocale: "es"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {
			"currency.name.credits":               "Créditos",
			"currency.name.diamonds":              "Diamantes",
			"catalog.page.groups":                 "Grupos",
			"catalog.page.guild_custom_furni":     "Muebles de grupo",
			"catalog.start.guild.purchase.button": "Crear un grupo",
			"navigator.searchcode.title.groups":   "Sedes de grupos populares",
		},
	}))

	config := requestBody(t, app, UIConfigPath)
	if !strings.Contains(config, `"system.currency.types":[-1,5]`) {
		t.Fatalf("unexpected config %s", config)
	}
	if !strings.Contains(config, `"navigator.room.models":[{"clubLevel":0,"tileSize":104,"name":"a"}]`) {
		t.Fatalf("unexpected room models %s", config)
	}
	if !strings.Contains(config, `"crafting.recycler.batch.size":8`) {
		t.Fatalf("unexpected recycler config %s", config)
	}
	texts := requestBody(t, app, "/client/texts/es/ExternalTexts.json")
	if !strings.Contains(texts, `"purse.seasonal.currency.5":"Diamantes"`) {
		t.Fatalf("unexpected texts %s", texts)
	}
	if !strings.Contains(texts, `"catalog.page.groups":"Grupos"`) || !strings.Contains(texts, `"catalog.page.guild_custom_furni":"Muebles de grupo"`) || !strings.Contains(texts, `"catalog.start.guild.purchase.button":"Crear un grupo"`) || !strings.Contains(texts, `"navigator.searchcode.title.groups":"Sedes de grupos populares"`) {
		t.Fatalf("missing group catalog texts %s", texts)
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

// fakeLayouts provides Nitro room models.
type fakeLayouts struct{}

// Create creates no layout.
func (fakeLayouts) Create(context.Context, roomlayout.SaveParams) (roomlayout.Layout, error) {
	return roomlayout.Layout{}, nil
}

// Update updates no layout.
func (fakeLayouts) Update(context.Context, int64, roomlayout.SaveParams) (roomlayout.Layout, error) {
	return roomlayout.Layout{}, nil
}

// FindByID finds no layout.
func (fakeLayouts) FindByID(context.Context, int64) (roomlayout.Layout, bool, error) {
	return roomlayout.Layout{}, false, nil
}

// FindByName finds no layout.
func (fakeLayouts) FindByName(context.Context, string) (roomlayout.Layout, bool, error) {
	return roomlayout.Layout{}, false, nil
}

// List lists enabled and disabled layouts.
func (fakeLayouts) List(context.Context) ([]roomlayout.Layout, error) {
	return []roomlayout.Layout{{Name: "model_a", TileSize: 104, ClubLevel: 2, Enabled: true}, {Name: "model_b", TileSize: 94}}, nil
}

// Catalog returns no layout catalog.
func (fakeLayouts) Catalog(context.Context) (*roomlayout.Catalog, error) { return nil, nil }
