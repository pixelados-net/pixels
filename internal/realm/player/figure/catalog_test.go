package figure

import (
	"os"
	"path/filepath"
	"testing"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// TestCatalogAllowed verifies existence, gender, club, color, and clothing ownership gates.
func TestCatalogAllowed(t *testing.T) {
	catalog := testCatalog(t)
	tests := []struct {
		name     string
		figure   string
		gender   playermodel.Gender
		club     playermodel.ClubLevel
		unlocked []int32
		allowed  bool
	}{
		{name: "ordinary", figure: "hd-180-1", gender: playermodel.GenderMale, allowed: true},
		{name: "missing", figure: "hd-999-1", gender: playermodel.GenderMale},
		{name: "gender", figure: "hr-200-1", gender: playermodel.GenderFemale},
		{name: "club missing", figure: "ch-300-1", gender: playermodel.GenderMale},
		{name: "club", figure: "ch-300-1", gender: playermodel.GenderMale, club: playermodel.ClubLevelHC, allowed: true},
		{name: "sellable missing", figure: "ha-400-1", gender: playermodel.GenderMale},
		{name: "sellable owned", figure: "ha-400-1", gender: playermodel.GenderMale, unlocked: []int32{400}, allowed: true},
		{name: "color", figure: "hd-180-9", gender: playermodel.GenderMale},
	}
	for _, test := range tests {
		if actual := catalog.Allowed(test.figure, test.gender, test.club, test.unlocked); actual != test.allowed {
			t.Fatalf("%s allowed=%v expected %v", test.name, actual, test.allowed)
		}
	}
}

// TestJSONCatalogAllowed verifies Nitro Renderer JSON entitlement loading.
func TestJSONCatalogAllowed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "FigureData.json")
	data := []byte(`{"palettes":[{"id":1,"colors":[{"id":1,"club":0,"selectable":true}]}],"setTypes":[{"type":"hd","paletteId":1,"sets":[{"id":180,"gender":"U","club":0,"selectable":true,"sellable":false},{"id":181,"gender":"U","club":0,"selectable":true,"sellable":true}]}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write JSON figure catalog: %v", err)
	}
	catalog, err := NewCatalog(Config{Path: path})
	if err != nil {
		t.Fatalf("load JSON figure catalog: %v", err)
	}
	if !catalog.Allowed("hd-180-1", playermodel.GenderMale, playermodel.ClubLevelNone, nil) {
		t.Fatal("ordinary JSON figure rejected")
	}
	if catalog.Allowed("hd-181-1", playermodel.GenderMale, playermodel.ClubLevelNone, nil) {
		t.Fatal("locked JSON figure accepted")
	}
	if !catalog.Allowed("hd-181-1", playermodel.GenderMale, playermodel.ClubLevelNone, []int32{181}) {
		t.Fatal("unlocked JSON figure rejected")
	}
}

// BenchmarkCatalogAllowed measures immutable entitlement validation.
func BenchmarkCatalogAllowed(benchmark *testing.B) {
	catalog := testCatalog(benchmark)
	benchmark.ReportAllocs()
	for range benchmark.N {
		if !catalog.Allowed("hd-180-1.ha-400-1", playermodel.GenderMale, playermodel.ClubLevelNone, []int32{400}) {
			benchmark.Fatal("valid entitled figure rejected")
		}
	}
}

// testCatalog loads a minimal authoritative fixture.
func testCatalog(testingObject testing.TB) *Catalog {
	testingObject.Helper()
	path := filepath.Join(testingObject.TempDir(), "figuredata.xml")
	data := []byte(`<figuredata><colors><palette id="1"><color id="1" club="0" selectable="1"/></palette></colors><sets><settype type="hd" paletteid="1"><set id="180" gender="U" club="0" selectable="1"/></settype><settype type="hr" paletteid="1"><set id="200" gender="M" club="0" selectable="1"/></settype><settype type="ch" paletteid="1"><set id="300" gender="U" club="1" selectable="1"/></settype><settype type="ha" paletteid="1"><set id="400" gender="U" club="0" selectable="1" sellable="1"/></settype></sets></figuredata>`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		testingObject.Fatalf("write figure catalog: %v", err)
	}
	catalog, err := NewCatalog(Config{Path: path})
	if err != nil {
		testingObject.Fatalf("load figure catalog: %v", err)
	}
	return catalog
}
