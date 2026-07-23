package request

import (
	"strings"
	"time"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// Promo creates or replaces one promotional badge definition.
type Promo struct {
	Audit
	// Code identifies a promotion during creation.
	Code string `json:"code"`
	// BadgeCode identifies the awarded badge.
	BadgeCode string `json:"badgeCode"`
	// StartsAt stores the optional opening instant.
	StartsAt *time.Time `json:"startsAt"`
	// EndsAt stores the optional closing instant.
	EndsAt *time.Time `json:"endsAt"`
	// MaxClaims stores zero for unlimited claims.
	MaxClaims int64 `json:"maxClaims"`
	// Enabled optionally controls availability and defaults true.
	Enabled *bool `json:"enabled"`
}

// Value maps one full promotion request to a domain record.
func (request Promo) Value(code string) progressionrecord.PromoBadge {
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	return progressionrecord.PromoBadge{Code: strings.ToUpper(strings.TrimSpace(code)), BadgeCode: strings.ToUpper(strings.TrimSpace(request.BadgeCode)), StartsAt: request.StartsAt, EndsAt: request.EndsAt, MaxClaims: request.MaxClaims, Enabled: enabled}
}

// ValidPromo validates one promotion window and cap.
func ValidPromo(value progressionrecord.PromoBadge) bool {
	return value.Code != "" && value.BadgeCode != "" && value.MaxClaims >= 0 && (value.StartsAt == nil || value.EndsAt == nil || value.StartsAt.Before(*value.EndsAt))
}

// PromoClaim forces or normally evaluates one manual claim.
type PromoClaim struct {
	Audit
	// Force bypasses the availability window and global claim cap.
	Force bool `json:"force"`
}
