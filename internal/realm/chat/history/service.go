package history

import (
	"context"

	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	historyrepo "github.com/niflaot/pixels/internal/realm/chat/history/repository"
)

// Service queries bounded chat history.
type Service struct {
	// store reads partitioned history.
	store historyrepo.Store
}

// NewService creates chat history query behavior.
func NewService(store historyrepo.Store) *Service { return &Service{store: store} }

// History returns one bounded keyset page.
func (service *Service) History(ctx context.Context, query historymodel.Query) ([]historymodel.Entry, error) {
	return service.store.History(ctx, query.Normalize())
}
