package record

import "context"

// Store persists trade settlement audits and owns transactions.
type Store interface {
	WithinTransaction(context.Context, func(context.Context) error) error
	InsertAudit(context.Context, Audit) error
	ListAudits(context.Context, int64, int32) ([]Audit, error)
}
