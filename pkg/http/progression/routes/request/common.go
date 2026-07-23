// Package request defines progression administration HTTP payloads.
package request

// Audit attributes one administrative mutation.
type Audit struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason"`
}
