// Package rightsrevoked defines the room rights revoked event.
package rightsrevoked

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room rights revocation.
	Name bus.Name = "room.rights_revoked"
	// ActionExplicit identifies one explicit revocation.
	ActionExplicit Action = "revoked"
	// ActionAll identifies a revoke-all operation.
	ActionAll Action = "revoked_all"
	// ActionRelinquished identifies self-revocation.
	ActionRelinquished Action = "relinquished"
)

// Action names a room rights revocation kind.
type Action string

// Payload describes a committed room rights revocation.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the former rights holder.
	PlayerID int64
	// ActorID identifies the revoker.
	ActorID int64
	// Action identifies how rights were removed.
	Action Action
}
