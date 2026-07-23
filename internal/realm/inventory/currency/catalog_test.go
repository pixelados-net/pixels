package currency

import (
	"errors"
	"testing"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestLoadCatalogValidatesProjectDefinitions verifies the committed catalog and translations.
func TestLoadCatalogValidatesProjectDefinitions(t *testing.T) {
	catalog, err := LoadCatalog(Config{Types: "-1:credits,0:duckets,5:diamonds", LedgerTypes: []int32{-1}})
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	translations, err := i18n.LoadCatalog(i18n.Config{Path: "../../../../i18n/translations.json"}, nil)
	if err != nil {
		t.Fatalf("load translations: %v", err)
	}

	for _, definition := range catalog.Types() {
		key := i18n.Key("currency.name." + definition.Key)
		if translated := translations.Default(key); translated == string(key) {
			t.Fatalf("missing translation %q", key)
		}
	}
}

// TestLoadCatalogRejectsMalformedEnvironment verifies invalid entries fail startup.
func TestLoadCatalogRejectsMalformedEnvironment(t *testing.T) {
	for _, value := range []string{"", "credits", "abc:credits", "-1:"} {
		if _, err := LoadCatalog(Config{Types: value}); !errors.Is(err, ErrInvalidCatalog) {
			t.Fatalf("expected invalid catalog for %q, got %v", value, err)
		}
	}
}

// TestNewCatalogRejectsInvalidDefinitions verifies catalog invariants.
func TestNewCatalogRejectsInvalidDefinitions(t *testing.T) {
	_, err := NewCatalog(nil, nil)
	if !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("expected invalid empty catalog, got %v", err)
	}

	_, err = NewCatalog(projectDefinitionsForTest(), nil)
	if !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("expected duplicate type error, got %v", err)
	}
}

// TestCatalogReturnsCopies verifies callers cannot mutate catalog ordering data.
func TestCatalogReturnsCopies(t *testing.T) {
	catalog, err := NewCatalog(projectDefinitionsForTest()[:1], []int32{-1})
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}
	types := catalog.Types()
	types[0].Key = "changed"
	definition, found := catalog.Type(-1)
	if !found || definition.Key != "credits" || !definition.Ledger {
		t.Fatalf("unexpected immutable definition %#v", definition)
	}
}

// projectDefinitionsForTest returns duplicate definitions for validation tests.
func projectDefinitionsForTest() []currencymodel.Definition {
	return []currencymodel.Definition{{Type: -1, Key: "credits"}, {Type: -1, Key: "duplicate"}}
}
