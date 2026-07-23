// Package applied defines the global sanction applied event.
package applied

import (
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies an applied punishment.
const Name bus.Name = "sanction.applied"

// Payload describes one applied punishment projection.
type Payload struct {
	// PunishmentID identifies the record.
	PunishmentID int64
	// ReceiverID identifies the target player.
	ReceiverID int64
	// Kind identifies the effect.
	Kind sanctionrecord.Kind
	// ExpiresAt optionally bounds the effect.
	ExpiresAt *time.Time
	// Source identifies the issuer workflow.
	Source string
}
