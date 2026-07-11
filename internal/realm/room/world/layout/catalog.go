package layout

// Catalog stores available room layouts by model name.
type Catalog struct {
	// layouts stores layouts keyed by normalized name.
	layouts map[string]Layout
}

// NewCatalog creates a layout catalog.
func NewCatalog(layouts []Layout) (*Catalog, error) {
	catalog := &Catalog{layouts: make(map[string]Layout, len(layouts))}
	for _, roomLayout := range layouts {
		roomLayout.Name = NormalizeName(roomLayout.Name)
		if !roomLayout.Valid() {
			return nil, ErrInvalidLayout
		}
		catalog.layouts[roomLayout.Name] = roomLayout
	}

	return catalog, nil
}

// Find returns a layout by model name.
func (catalog *Catalog) Find(name string) (Layout, bool) {
	if catalog == nil {
		return Layout{}, false
	}

	roomLayout, found := catalog.layouts[NormalizeName(name)]

	return roomLayout, found
}

// MustFind returns a layout by model name or an error.
func (catalog *Catalog) MustFind(name string) (Layout, error) {
	roomLayout, found := catalog.Find(name)
	if !found {
		return Layout{}, ErrLayoutNotFound
	}

	return roomLayout, nil
}

// List returns all registered layouts.
func (catalog *Catalog) List() []Layout {
	if catalog == nil {
		return nil
	}

	layouts := make([]Layout, 0, len(catalog.layouts))
	for _, roomLayout := range catalog.layouts {
		layouts = append(layouts, roomLayout)
	}

	return layouts
}
