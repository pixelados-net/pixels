package record

import "context"

// AuditAttribution stores trusted administrative mutation attribution.
type AuditAttribution struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64
	// Reason stores the required human-readable justification.
	Reason string
}

// auditContextKey isolates social-group audit attribution in contexts.
type auditContextKey struct{}

// WithAudit attaches trusted administrative attribution to a mutation context.
func WithAudit(ctx context.Context, actorPlayerID int64, reason string) context.Context {
	return context.WithValue(ctx, auditContextKey{}, AuditAttribution{ActorPlayerID: actorPlayerID, Reason: reason})
}

// AuditFromContext returns trusted administrative attribution when present.
func AuditFromContext(ctx context.Context) (AuditAttribution, bool) {
	attribution, found := ctx.Value(auditContextKey{}).(AuditAttribution)
	return attribution, found && attribution.ActorPlayerID > 0
}
