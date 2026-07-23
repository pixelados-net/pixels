package figure

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// Catalog stores immutable avatar set and color entitlement rules.
type Catalog struct {
	// sets stores rules by two-byte type and set identifier.
	sets map[setKey]setRule
	// colors stores rules by palette and color identifier.
	colors map[colorKey]colorRule
}

// setKey identifies one typed avatar set.
type setKey struct {
	// kind stores the two-byte figure type.
	kind [2]byte
	// id stores the figure set identifier.
	id int32
}

// setRule contains immutable figure-set policy.
type setRule struct {
	// paletteID identifies the accepted color palette.
	paletteID int32
	// club stores the minimum club tier.
	club int16
	// gender stores M, F, or U.
	gender byte
	// selectable reports ordinary editor availability.
	selectable bool
	// sellable reports that ownership is required.
	sellable bool
}

// colorKey identifies one palette color.
type colorKey struct {
	// paletteID identifies the palette.
	paletteID int32
	// id identifies the color.
	id int32
}

// colorRule contains immutable palette policy.
type colorRule struct {
	// club stores the minimum club tier.
	club int16
	// selectable reports editor availability.
	selectable bool
}

// NewCatalog loads and validates the configured figuredata JSON or XML once.
func NewCatalog(config Config) (*Catalog, error) {
	data, name, err := loadFigureData(config)
	if err != nil {
		return nil, fmt.Errorf("load figure data: %w", err)
	}
	catalog := &Catalog{sets: make(map[setKey]setRule), colors: make(map[colorKey]colorRule)}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json":
		err = catalog.decodeJSON(json.NewDecoder(bytes.NewReader(data)))
	case ".xml":
		err = catalog.decode(xml.NewDecoder(bytes.NewReader(data)))
	default:
		err = fmt.Errorf("unsupported figure data extension %q", filepath.Ext(name))
	}
	if err != nil {
		return nil, fmt.Errorf("decode figure data: %w", err)
	}
	if len(catalog.sets) == 0 || len(catalog.colors) == 0 {
		return nil, fmt.Errorf("figure data is empty")
	}
	return catalog, nil
}

// Allowed validates figure existence and account entitlement without allocations.
func (catalog *Catalog) Allowed(value string, gender playermodel.Gender, club playermodel.ClubLevel, unlocked []int32) bool {
	parts, count, valid := parse(value)
	if !valid || catalog == nil {
		return false
	}
	for index := 0; index < count; index++ {
		part := parts[index]
		rule, found := catalog.sets[setKey{kind: part.kind, id: part.setID}]
		owned := contains(unlocked, part.setID)
		if !found || rule.gender != 'U' && rule.gender != genderByte(gender) || int16(club) < rule.club || (!rule.selectable || rule.sellable) && !owned {
			return false
		}
		for colorIndex := 0; colorIndex < part.colorCount; colorIndex++ {
			color, colorFound := catalog.colors[colorKey{paletteID: rule.paletteID, id: part.colors[colorIndex]}]
			if !colorFound || !color.selectable || int16(club) < color.club {
				return false
			}
		}
	}
	return true
}

// contains reports whether one bounded unlock snapshot contains a set.
func contains(values []int32, expected int32) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

// genderByte returns the figuredata gender code.
func genderByte(gender playermodel.Gender) byte {
	value := strings.ToUpper(string(gender))
	if len(value) != 1 {
		return 0
	}
	return value[0]
}

// attribute returns one XML attribute value.
func attribute(attributes []xml.Attr, name string) string {
	for _, current := range attributes {
		if current.Name.Local == name {
			return current.Value
		}
	}
	return ""
}

// integerAttribute parses one bounded XML integer attribute.
func integerAttribute(attributes []xml.Attr, name string) int32 {
	value, _ := strconv.ParseInt(attribute(attributes, name), 10, 32)
	return int32(value)
}
