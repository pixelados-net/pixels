// Package revoked defines the global sanction revoked event.
package revoked

import (
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies a revoked punishment.
const Name bus.Name = "sanction.revoked"

// Payload describes one revoked punishment projection.
type Payload struct {
	// PunishmentID identifies the record.
	PunishmentID int64
	// ReceiverID identifies the target player.
	ReceiverID int64
	// Kind identifies the effect.
	Kind sanctionrecord.Kind
}
