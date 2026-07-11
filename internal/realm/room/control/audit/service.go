package audit

import (
	"context"

	auditmodel "github.com/niflaot/pixels/internal/realm/room/control/audit/model"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
)

const (
	// defaultLimit stores the default audit page size.
	defaultLimit = 50
	// maxLimit stores the maximum audit page size.
	maxLimit = 200
)

// Query filters room audit records.
type Query struct {
	// RoomID optionally scopes one room.
	RoomID *int64
	// TargetPlayerID optionally scopes one affected player.
	TargetPlayerID *int64
	// ActorPlayerID optionally scopes one acting player.
	ActorPlayerID *int64
	// ActionTypes optionally restricts moderation action kinds.
	ActionTypes []moderationmodel.Action
	// Before optionally excludes ids at or above the cursor.
	Before *int64
	// Limit caps returned records.
	Limit int
}

// StoreQuery contains persistence-level audit filters.
type StoreQuery Query

// Store persists and reads room audit records.
type Store interface {
	// InsertRights appends one rights audit row.
	InsertRights(context.Context, auditmodel.RightsAudit) error
	// InsertModeration appends one moderation audit row.
	InsertModeration(context.Context, auditmodel.ModerationAction) error
	// RightsHistory lists matching rights history.
	RightsHistory(context.Context, StoreQuery) ([]auditmodel.RightsAudit, error)
	// ModerationHistory lists matching moderation history.
	ModerationHistory(context.Context, StoreQuery) ([]auditmodel.ModerationAction, error)
}

// Manager reads room rights and moderation audit history.
type Manager interface {
	// ModerationHistory lists matching moderation actions newest first.
	ModerationHistory(ctx context.Context, query Query) ([]auditmodel.ModerationAction, error)
	// RightsHistory lists matching rights actions newest first.
	RightsHistory(ctx context.Context, query Query) ([]auditmodel.RightsAudit, error)
}

// Service validates and reads room audit history.
type Service struct {
	// store persists and reads audit records.
	store Store
}

// New creates a room audit service.
func New(store Store) *Service {
	return &Service{store: store}
}

// ModerationHistory lists matching moderation actions newest first.
func (service *Service) ModerationHistory(ctx context.Context, query Query) ([]auditmodel.ModerationAction, error) {
	query, err := normalizeQuery(query)
	if err != nil {
		return nil, err
	}

	return service.store.ModerationHistory(ctx, repositoryQuery(query))
}

// RightsHistory lists matching rights actions newest first.
func (service *Service) RightsHistory(ctx context.Context, query Query) ([]auditmodel.RightsAudit, error) {
	query.ActionTypes = nil
	query, err := normalizeQuery(query)
	if err != nil {
		return nil, err
	}

	return service.store.RightsHistory(ctx, repositoryQuery(query))
}

// normalizeQuery validates ids, actions, cursor, and page size.
func normalizeQuery(query Query) (Query, error) {
	if invalidOptionalID(query.RoomID) || invalidOptionalID(query.TargetPlayerID) || invalidOptionalID(query.ActorPlayerID) || invalidOptionalID(query.Before) {
		return Query{}, ErrInvalidQuery
	}
	for _, action := range query.ActionTypes {
		if !action.Valid() {
			return Query{}, ErrInvalidQuery
		}
	}
	if query.Limit <= 0 {
		query.Limit = defaultLimit
	}
	if query.Limit > maxLimit {
		query.Limit = maxLimit
	}

	return query, nil
}

// invalidOptionalID reports a present non-positive id.
func invalidOptionalID(value *int64) bool {
	return value != nil && *value <= 0
}

// repositoryQuery maps service filters to repository filters.
func repositoryQuery(query Query) StoreQuery {
	return StoreQuery(query)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
