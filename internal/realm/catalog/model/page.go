// Package model contains durable catalog records.
package model

import sharedmodel "github.com/niflaot/pixels/pkg/model"

const (
	// DefaultLayout identifies the first supported catalog page layout.
	DefaultLayout = "default_3x3"
)

// Page contains one persistent catalog page.
type Page struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// ParentID identifies the optional parent page.
	ParentID *int64

	// Name stores the stable localization slug.
	Name string

	// Layout identifies the client page layout.
	Layout string

	// IconColor stores the client icon color.
	IconColor int32

	// IconImage stores the client icon image.
	IconImage int32

	// MinRank stores the minimum visible player rank.
	MinRank int32

	// OrderNum stores sibling display order.
	OrderNum int32

	// Visible reports whether the page appears in the catalog tree.
	Visible bool

	// Enabled reports whether the page can be opened.
	Enabled bool

	// ClubOnly reports whether the page requires club membership.
	ClubOnly bool
}

// Accessible reports whether a player can see and open the page.
func (page Page) Accessible(rank int32, hasClub bool) bool {
	return page.Visible && page.Enabled && rank >= page.MinRank && (!page.ClubOnly || hasClub)
}
