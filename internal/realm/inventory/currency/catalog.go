package currency

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

var (
	// ErrInvalidCatalog reports invalid currency catalog data.
	ErrInvalidCatalog = errors.New("invalid currency catalog")

	// ErrUnknownType reports an unconfigured currency type.
	ErrUnknownType = errors.New("unknown currency type")
)

// Catalog stores immutable configured currency definitions.
type Catalog struct {
	// definitions stores definitions in configured display order.
	definitions []currencymodel.Definition

	// byType stores definitions by protocol type.
	byType map[int32]currencymodel.Definition
}

// LoadCatalog reads and validates the configured currency catalog.
func LoadCatalog(config Config) (*Catalog, error) {
	data, err := os.ReadFile(config.CatalogPath)
	if err != nil {
		return nil, fmt.Errorf("read currency catalog %q: %w", config.CatalogPath, err)
	}

	var definitions []currencymodel.Definition
	if err := json.Unmarshal(data, &definitions); err != nil {
		return nil, fmt.Errorf("decode currency catalog %q: %w", config.CatalogPath, err)
	}

	return NewCatalog(definitions, config.LedgerTypes)
}

// NewCatalog validates definitions and applies ledger configuration.
func NewCatalog(definitions []currencymodel.Definition, ledgerTypes []int32) (*Catalog, error) {
	ledger := typeSet(ledgerTypes)
	byType := make(map[int32]currencymodel.Definition, len(definitions))
	copied := make([]currencymodel.Definition, len(definitions))
	for index, definition := range definitions {
		definition.Key = strings.TrimSpace(definition.Key)
		if definition.Key == "" {
			return nil, fmt.Errorf("%w: currency type %d has no key", ErrInvalidCatalog, definition.Type)
		}
		if _, exists := byType[definition.Type]; exists {
			return nil, fmt.Errorf("%w: duplicate currency type %d", ErrInvalidCatalog, definition.Type)
		}

		definition.Ledger = ledger[definition.Type]
		copied[index] = definition
		byType[definition.Type] = definition
	}

	if len(copied) == 0 {
		return nil, fmt.Errorf("%w: catalog is empty", ErrInvalidCatalog)
	}

	return &Catalog{definitions: copied, byType: byType}, nil
}

// Types returns a stable copy of configured currency definitions.
func (catalog *Catalog) Types() []currencymodel.Definition {
	return append([]currencymodel.Definition(nil), catalog.definitions...)
}

// Type returns one configured currency definition.
func (catalog *Catalog) Type(currencyType int32) (currencymodel.Definition, bool) {
	definition, found := catalog.byType[currencyType]

	return definition, found
}

// typeSet creates a currency type membership set.
func typeSet(types []int32) map[int32]bool {
	set := make(map[int32]bool, len(types))
	for _, currencyType := range types {
		set[currencyType] = true
	}

	return set
}
