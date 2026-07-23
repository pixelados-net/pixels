// Package model contains durable catalog records.
package model

import (
	"time"

	"github.com/niflaot/pixels/internal/permission"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

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

	// RequiredNode stores the optional permission needed to access the page.
	RequiredNode *permission.Node

	// OrderNum stores sibling display order.
	OrderNum int32

	// Visible reports whether the page appears in the catalog tree.
	Visible bool

	// Enabled reports whether the page can be opened.
	Enabled bool

	// ClubOnly reports whether the page requires club membership.
	ClubOnly bool

	// NewAdditions reports whether new offers contribute to the novelty badge.
	NewAdditions bool

	// ExpiresAt stores the optional client-facing page expiry.
	ExpiresAt *time.Time

	// ExcludedFromKickback excludes purchases from HC kickback totals.
	ExcludedFromKickback bool
}

// Accessible reports whether visibility and club requirements allow page access.
func (page Page) Accessible(hasClub bool) bool {
	return page.Visible && page.Enabled && (!page.ClubOnly || hasClub)
}
