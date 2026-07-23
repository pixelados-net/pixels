// Package model contains room audit records.
package model

import "time"

// RightsAction names one audited rights mutation.
type RightsAction string

const (
	// RightsGranted records a grant.
	RightsGranted RightsAction = "granted"
	// RightsRevoked records an explicit revoke.
	RightsRevoked RightsAction = "revoked"
	// RightsRevokedAll records a revoke-all mutation.
	RightsRevokedAll RightsAction = "revoked_all"
	// RightsRelinquished records self-revocation.
	RightsRelinquished RightsAction = "relinquished"
)

// RightsAudit stores one append-only room rights action.
type RightsAudit struct {
	// ID identifies the audit row.
	ID int64 `json:"id"`
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the affected player.
	PlayerID int64 `json:"playerId"`
	// ActorKind identifies the source family.
	ActorKind string `json:"actorKind"`
	// ActorID optionally identifies the source player.
	ActorID *int64 `json:"actorId,omitempty"`
	// Action identifies the mutation.
	Action RightsAction `json:"action"`
	// CreatedAt stores when the action occurred.
	CreatedAt time.Time `json:"createdAt"`
}
