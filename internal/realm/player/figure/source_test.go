package figure

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

const (
	// testFigureDataJSON contains one valid Nitro figure-data document.
	testFigureDataJSON = `{"palettes":[{"id":1,"colors":[{"id":1,"club":0,"selectable":true}]}],"setTypes":[{"type":"hd","paletteId":1,"sets":[{"id":180,"gender":"U","club":0,"selectable":true,"sellable":false}]}]}`
	// testFigureDataXML contains one valid legacy figure-data document.
	testFigureDataXML = `<figuredata><colors><palette id="1"><color id="1" club="0" selectable="1"/></palette></colors><sets><settype type="hd" paletteid="1"><set id="180" gender="U" club="0" selectable="1"/></settype></sets></figuredata>`
)

// sourceErrorCase describes one rejected remote source.
type sourceErrorCase struct {
	// name identifies the test case.
	name string
	// config contains the rejected source.
	config Config
	// errorMatch contains the expected diagnostic fragment.
	errorMatch string
}

// TestRemoteCatalogAllowed verifies Nitro FigureData JSON loading over HTTP.
func TestRemoteCatalogAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/FigureData.json" {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(testFigureDataJSON))
	}))
	defer server.Close()
	catalog, err := NewCatalog(Config{URL: server.URL + "/FigureData.json", Timeout: time.Second, MaxBytes: 1024})
	if err != nil {
		t.Fatalf("load remote figure catalog: %v", err)
	}
	if !catalog.Allowed("hd-180-1", playermodel.GenderMale, playermodel.ClubLevelNone, nil) {
		t.Fatal("remote JSON figure rejected")
	}
}

// TestRemoteCatalogIntegration verifies an explicitly requested live FigureData endpoint.
func TestRemoteCatalogIntegration(t *testing.T) {
	sourceURL := os.Getenv("PIXELS_TEST_FIGURE_DATA_URL")
	if sourceURL == "" {
		t.Skip("PIXELS_TEST_FIGURE_DATA_URL is not set")
	}
	catalog, err := NewCatalog(Config{URL: sourceURL})
	if err != nil {
		t.Fatalf("load live figure catalog: %v", err)
	}
	if len(catalog.sets) == 0 || len(catalog.colors) == 0 {
		t.Fatal("live figure catalog is empty")
	}
	t.Logf("loaded %d figure sets and %d colors", len(catalog.sets), len(catalog.colors))
}

// TestRemoteCatalogErrors verifies status, size, scheme, and extension failures.
func TestRemoteCatalogErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/missing.json":
			http.Error(writer, "missing", http.StatusNotFound)
		case "/large.json":
			_, _ = writer.Write([]byte(testFigureDataJSON))
		default:
			_, _ = writer.Write([]byte(testFigureDataJSON))
		}
	}))
	defer server.Close()
	tests := []sourceErrorCase{
		{name: "status", config: Config{URL: server.URL + "/missing.json"}, errorMatch: "unexpected status 404"},
		{name: "size", config: Config{URL: server.URL + "/large.json", MaxBytes: 8}, errorMatch: "exceeds 8 bytes"},
		{name: "scheme", config: Config{URL: "file:///FigureData.json"}, errorMatch: "unsupported figure data URL scheme"},
		{name: "extension", config: Config{URL: server.URL + "/FigureData"}, errorMatch: "unsupported figure data extension"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCatalog(test.config)
			if err == nil || !strings.Contains(err.Error(), test.errorMatch) {
				t.Fatalf("error=%v expected to contain %q", err, test.errorMatch)
			}
		})
	}
}

// TestLocalCatalogOverridesURL verifies explicit local sources do not require the network.
func TestLocalCatalogOverridesURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "figuredata.xml")
	if err := os.WriteFile(path, []byte(testFigureDataXML), 0o600); err != nil {
		t.Fatalf("write local figure catalog: %v", err)
	}
	catalog, err := NewCatalog(Config{URL: "https://invalid.invalid/FigureData.json", Path: path})
	if err != nil {
		t.Fatalf("load local figure catalog: %v", err)
	}
	if !catalog.Allowed("hd-180-1", playermodel.GenderMale, playermodel.ClubLevelNone, nil) {
		t.Fatal("local XML figure rejected")
	}
}

// TestLoadConfigDefaults verifies the production FigureData endpoint and safety limits.
func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("PIXELS_FIGURE_DATA_URL", "")
	t.Setenv("PIXELS_FIGURE_DATA_PATH", "")
	t.Setenv("PIXELS_FIGURE_DATA_TIMEOUT", "")
	t.Setenv("PIXELS_FIGURE_DATA_MAX_BYTES", "")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load figure config: %v", err)
	}
	if config.URL != "https://storageapi.pixelados.net/assets-prod/gamedata/FigureData.json" {
		t.Fatalf("unexpected URL %q", config.URL)
	}
	if config.Path != "" || config.Timeout != 15*time.Second || config.MaxBytes != 16*1024*1024 {
		t.Fatalf("unexpected defaults: %+v", config)
	}
}
