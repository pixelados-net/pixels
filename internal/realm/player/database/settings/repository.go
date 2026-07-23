// Package settings implements PostgreSQL persistence for player settings.
package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	playersettings "github.com/niflaot/pixels/internal/realm/player/settings"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// findSQL reads settings while creating a default row when needed.
	findSQL = `insert into player_settings(player_id) values($1) on conflict(player_id) do update set player_id=excluded.player_id returning player_id,volume_system,volume_furniture,volume_trax,old_chat,camera_follow_blocked,safety_locked,version`
	// volumeSQL replaces volume settings and returns the new snapshot.
	volumeSQL = `insert into player_settings(player_id,volume_system,volume_furniture,volume_trax) values($1,$2,$3,$4) on conflict(player_id) do update set volume_system=excluded.volume_system,volume_furniture=excluded.volume_furniture,volume_trax=excluded.volume_trax,updated_at=now(),version=player_settings.version+1 returning player_id,volume_system,volume_furniture,volume_trax,old_chat,camera_follow_blocked,safety_locked,version`
	// oldChatSQL replaces old-chat selection and returns the new snapshot.
	oldChatSQL = `insert into player_settings(player_id,old_chat) values($1,$2) on conflict(player_id) do update set old_chat=excluded.old_chat,updated_at=now(),version=player_settings.version+1 returning player_id,volume_system,volume_furniture,volume_trax,old_chat,camera_follow_blocked,safety_locked,version`
	// cameraFollowSQL replaces camera-follow privacy and returns the new snapshot.
	cameraFollowSQL = `insert into player_settings(player_id,camera_follow_blocked) values($1,$2) on conflict(player_id) do update set camera_follow_blocked=excluded.camera_follow_blocked,updated_at=now(),version=player_settings.version+1 returning player_id,volume_system,volume_furniture,volume_trax,old_chat,camera_follow_blocked,safety_locked,version`
	// homeRoomSQL replaces the profile home-room reference after room validation.
	homeRoomSQL = `update player_profiles set home_room_id=$2,updated_at=now(),version=version+1 where player_id=$1 and ($2::bigint is null or exists(select 1 from rooms where id=$2 and deleted_at is null and not is_bundle_template)) returning player_id`
)

// Repository persists user settings.
type Repository struct {
	// pool runs PostgreSQL queries and participates in transaction scopes.
	pool *postgres.Pool
}

// New creates a settings repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Find returns persisted settings, creating defaults atomically.
func (repository *Repository) Find(ctx context.Context, playerID int64) (playersettings.Record, error) {
	return repository.scan(postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, findSQL, playerID))
}

// SetVolume replaces volume fields.
func (repository *Repository) SetVolume(ctx context.Context, playerID int64, system int32, furniture int32, trax int32) (playersettings.Record, error) {
	return repository.scan(postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, volumeSQL, playerID, system, furniture, trax))
}

// SetOldChat replaces old-chat selection.
func (repository *Repository) SetOldChat(ctx context.Context, playerID int64, oldChat bool) (playersettings.Record, error) {
	return repository.scan(postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, oldChatSQL, playerID, oldChat))
}

// SetCameraFollowBlocked replaces camera-follow privacy.
func (repository *Repository) SetCameraFollowBlocked(ctx context.Context, playerID int64, blocked bool) (playersettings.Record, error) {
	return repository.scan(postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, cameraFollowSQL, playerID, blocked))
}

// SetHomeRoom replaces or clears a home-room reference.
func (repository *Repository) SetHomeRoom(ctx context.Context, playerID int64, roomID *int64) error {
	var updatedPlayerID int64
	err := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, homeRoomSQL, playerID, roomID).Scan(&updatedPlayerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return playersettings.ErrInvalidHomeRoom
		}
		return fmt.Errorf("set player home room: %w", err)
	}
	return nil
}

// UpdateAdmin applies one optimistic protected settings mutation.
func (repository *Repository) UpdateAdmin(ctx context.Context, playerID int64, expectedVersion int64, patch playersettings.AdminPatch) (playersettings.Record, error) {
	current, err := repository.Find(ctx, playerID)
	if err != nil {
		return playersettings.Record{}, err
	}
	if current.Version != expectedVersion {
		return playersettings.Record{}, playersettings.ErrSettingsConflict
	}
	if patch.VolumeSystem != nil {
		current.VolumeSystem = *patch.VolumeSystem
	}
	if patch.VolumeFurniture != nil {
		current.VolumeFurniture = *patch.VolumeFurniture
	}
	if patch.VolumeTrax != nil {
		current.VolumeTrax = *patch.VolumeTrax
	}
	if patch.OldChat != nil {
		current.OldChat = *patch.OldChat
	}
	if patch.CameraFollowBlocked != nil {
		current.CameraFollowBlocked = *patch.CameraFollowBlocked
	}
	if patch.SafetyLocked != nil {
		current.SafetyLocked = *patch.SafetyLocked
	}
	row := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, `update player_settings set volume_system=$3,volume_furniture=$4,volume_trax=$5,old_chat=$6,camera_follow_blocked=$7,safety_locked=$8,updated_at=now(),version=version+1 where player_id=$1 and version=$2 returning player_id,volume_system,volume_furniture,volume_trax,old_chat,camera_follow_blocked,safety_locked,version`, playerID, expectedVersion, current.VolumeSystem, current.VolumeFurniture, current.VolumeTrax, current.OldChat, current.CameraFollowBlocked, current.SafetyLocked)
	updated, err := repository.scan(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return playersettings.Record{}, playersettings.ErrSettingsConflict
		}
		return playersettings.Record{}, err
	}
	return updated, nil
}

// scan decodes one settings row.
func (repository *Repository) scan(row pgx.Row) (playersettings.Record, error) {
	var record playersettings.Record
	err := row.Scan(&record.PlayerID, &record.VolumeSystem, &record.VolumeFurniture, &record.VolumeTrax, &record.OldChat, &record.CameraFollowBlocked, &record.SafetyLocked, &record.Version)
	if err != nil {
		return playersettings.Record{}, fmt.Errorf("scan player settings: %w", err)
	}
	return record, nil
}
