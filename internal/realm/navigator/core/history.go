package core

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/realm/navigator/record"
)

// RecordVisit records one admitted room visit.
func (service *Service) RecordVisit(ctx context.Context, playerID int64, roomID int64) error {
	if err := validatePlayerRoom(playerID, roomID); err != nil {
		return err
	}
	store, ok := service.store.(record.VisitStore)
	if !ok {
		return fmt.Errorf("navigator store does not support visit history")
	}
	return store.RecordVisit(ctx, playerID, roomID)
}

// RecordVisits persists one validated bounded visit batch.
func (service *Service) RecordVisits(ctx context.Context, visits []record.Visit) error {
	if len(visits) == 0 {
		return nil
	}
	for _, visit := range visits {
		if err := validatePlayerRoom(visit.PlayerID, visit.RoomID); err != nil || visit.VisitedAt.IsZero() {
			if err != nil {
				return err
			}
			return ErrInvalidPreference
		}
	}
	store, ok := service.store.(record.VisitBatchStore)
	if !ok {
		return fmt.Errorf("navigator store does not support batched visit history")
	}
	return store.RecordVisits(ctx, visits)
}

// ListRecentRoomIDs lists bounded recent room history.
func (service *Service) ListRecentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}
	store, ok := service.store.(record.VisitStore)
	if !ok {
		return nil, fmt.Errorf("navigator store does not support visit history")
	}
	return store.ListRecentRoomIDs(ctx, playerID, normalizeHistoryLimit(limit))
}

// ListFrequentRoomIDs lists bounded frequent room history.
func (service *Service) ListFrequentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}
	store, ok := service.store.(record.VisitStore)
	if !ok {
		return nil, fmt.Errorf("navigator store does not support visit history")
	}
	return store.ListFrequentRoomIDs(ctx, playerID, normalizeHistoryLimit(limit))
}

// DeleteVisitHistory removes one player's complete navigator visit history.
func (service *Service) DeleteVisitHistory(ctx context.Context, playerID int64) (int64, error) {
	if playerID <= 0 {
		return 0, ErrInvalidPlayerID
	}
	store, ok := service.store.(record.VisitAdminStore)
	if !ok {
		return 0, fmt.Errorf("navigator store does not support visit history administration")
	}
	return store.DeleteVisitHistory(ctx, playerID)
}

// normalizeHistoryLimit bounds history result sizes.
func normalizeHistoryLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}
