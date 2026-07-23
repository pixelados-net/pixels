package core

import "context"

// AddFavorite adds a favorite room for a player.
func (service *Service) AddFavorite(ctx context.Context, playerID int64, roomID int64, limit int32, unlimited bool) error {
	if err := validatePlayerRoom(playerID, roomID); err != nil {
		return err
	}
	if limit <= 0 {
		return ErrInvalidPreference
	}

	return service.store.AddFavorite(ctx, playerID, roomID, limit, unlimited)
}

// RemoveFavorite removes a favorite room for a player.
func (service *Service) RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error {
	if err := validatePlayerRoom(playerID, roomID); err != nil {
		return err
	}

	return service.store.RemoveFavorite(ctx, playerID, roomID)
}

// ListFavoriteRoomIDs lists favorite room ids for a player.
func (service *Service) ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}

	return service.store.ListFavoriteRoomIDs(ctx, playerID)
}

// validatePlayerRoom validates player and room ids.
func validatePlayerRoom(playerID int64, roomID int64) error {
	if playerID <= 0 {
		return ErrInvalidPlayerID
	}
	if roomID <= 0 {
		return ErrInvalidRoomID
	}

	return nil
}
