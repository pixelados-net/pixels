package figure

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
)

// jsonFigureData contains the avatar entitlement rules used by Nitro Renderer.
type jsonFigureData struct {
	// Palettes contains selectable avatar colors.
	Palettes []jsonPalette `json:"palettes"`
	// SetTypes contains typed avatar-part sets.
	SetTypes []jsonSetType `json:"setTypes"`
}

// jsonPalette contains one color collection.
type jsonPalette struct {
	// ID identifies the palette.
	ID int32 `json:"id"`
	// Colors contains the palette rules.
	Colors []jsonColor `json:"colors"`
}

// jsonColor contains one avatar color rule.
type jsonColor struct {
	// ID identifies the color.
	ID int32 `json:"id"`
	// Club stores the minimum club tier.
	Club int16 `json:"club"`
	// Selectable reports editor availability.
	Selectable bool `json:"selectable"`
}

// jsonSetType contains one typed collection of avatar sets.
type jsonSetType struct {
	// Type stores the figure part type.
	Type string `json:"type"`
	// PaletteID identifies the accepted color palette.
	PaletteID int32 `json:"paletteId"`
	// Sets contains the avatar set rules.
	Sets []jsonSet `json:"sets"`
}

// jsonSet contains one avatar set entitlement rule.
type jsonSet struct {
	// ID identifies the set.
	ID int32 `json:"id"`
	// Gender stores M, F, or U.
	Gender string `json:"gender"`
	// Club stores the minimum club tier.
	Club int16 `json:"club"`
	// Selectable reports editor availability.
	Selectable bool `json:"selectable"`
	// Sellable reports that ownership is required.
	Sellable bool `json:"sellable"`
}

// decodeJSON loads Nitro Renderer figure data into immutable lookup maps.
func (catalog *Catalog) decodeJSON(decoder *json.Decoder) error {
	var data jsonFigureData
	if err := decoder.Decode(&data); err != nil {
		return err
	}
	for _, palette := range data.Palettes {
		for _, color := range palette.Colors {
			if palette.ID <= 0 || color.ID <= 0 {
				continue
			}
			catalog.colors[colorKey{paletteID: palette.ID, id: color.ID}] = colorRule{club: color.Club, selectable: color.Selectable}
		}
	}
	for _, setType := range data.SetTypes {
		kind := strings.ToLower(setType.Type)
		if len(kind) != 2 {
			continue
		}
		keyKind := [2]byte{kind[0], kind[1]}
		for _, set := range setType.Sets {
			gender := strings.ToUpper(set.Gender)
			if set.ID <= 0 || len(gender) != 1 {
				continue
			}
			catalog.sets[setKey{kind: keyKind, id: set.ID}] = setRule{paletteID: setType.PaletteID, club: set.Club, gender: gender[0], selectable: set.Selectable, sellable: set.Sellable}
		}
	}
	return nil
}

// decode streams figuredata without retaining the XML tree.
func (catalog *Catalog) decode(decoder *xml.Decoder) error {
	var paletteID int32
	var setType [2]byte
	var setPaletteID int32
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "palette":
				paletteID = integerAttribute(value.Attr, "id")
			case "color":
				catalog.addColor(paletteID, value.Attr)
			case "settype":
				kind := strings.ToLower(attribute(value.Attr, "type"))
				if len(kind) == 2 {
					setType = [2]byte{kind[0], kind[1]}
					setPaletteID = integerAttribute(value.Attr, "paletteid")
				}
			case "set":
				catalog.addSet(setType, setPaletteID, value.Attr)
			}
		case xml.EndElement:
			if value.Name.Local == "palette" {
				paletteID = 0
			}
			if value.Name.Local == "settype" {
				setType, setPaletteID = [2]byte{}, 0
			}
		}
	}
}

// addColor stores one valid palette color rule.
func (catalog *Catalog) addColor(paletteID int32, attributes []xml.Attr) {
	id := integerAttribute(attributes, "id")
	if paletteID <= 0 || id <= 0 {
		return
	}
	catalog.colors[colorKey{paletteID: paletteID, id: id}] = colorRule{club: int16(integerAttribute(attributes, "club")), selectable: attribute(attributes, "selectable") != "0"}
}

// addSet stores one valid typed set rule.
func (catalog *Catalog) addSet(kind [2]byte, paletteID int32, attributes []xml.Attr) {
	id := integerAttribute(attributes, "id")
	gender := strings.ToUpper(attribute(attributes, "gender"))
	if kind == [2]byte{} || id <= 0 || len(gender) != 1 {
		return
	}
	catalog.sets[setKey{kind: kind, id: id}] = setRule{paletteID: paletteID, club: int16(integerAttribute(attributes, "club")), gender: gender[0], selectable: attribute(attributes, "selectable") != "0", sellable: attribute(attributes, "sellable") == "1"}
}
