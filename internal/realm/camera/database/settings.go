package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
)

// settingsSQL stores the canonical camera settings projection.
const settingsSQL = `select enabled, credits_price, points_price, points_type, publish_points_price, publish_points_type, publish_cooldown_seconds, updated_at, version from camera_settings where id=1`

// Settings returns the singleton operational camera policy.
func (repository *Repository) Settings(ctx context.Context) (camerarecord.Settings, error) {
	return scanSettings(repository.executorFor(ctx).QueryRow(ctx, settingsSQL))
}

// UpdateSettings applies one optimistic settings replacement.
func (repository *Repository) UpdateSettings(ctx context.Context, settings camerarecord.Settings, expectedVersion int64) (camerarecord.Settings, bool, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, `update camera_settings set enabled=$1, credits_price=$2, points_price=$3, points_type=$4, publish_points_price=$5, publish_points_type=$6, publish_cooldown_seconds=$7, updated_at=now(), version=version+1 where id=1 and version=$8 returning enabled, credits_price, points_price, points_type, publish_points_price, publish_points_type, publish_cooldown_seconds, updated_at, version`, settings.Enabled, settings.CreditsPrice, settings.PointsPrice, settings.PointsType, settings.PublishPointsPrice, settings.PublishPointsType, int64(settings.PublishCooldown/time.Second), expectedVersion)
	updated, err := scanSettings(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return camerarecord.Settings{}, false, nil
	}
	if err != nil {
		return camerarecord.Settings{}, false, err
	}
	return updated, true, nil
}

// settingsScanner scans one camera settings row.
type settingsScanner interface {
	// Scan decodes one database row.
	Scan(...any) error
}

// scanSettings scans one camera settings row.
func scanSettings(scanner settingsScanner) (camerarecord.Settings, error) {
	var settings camerarecord.Settings
	var cooldownSeconds int64
	err := scanner.Scan(&settings.Enabled, &settings.CreditsPrice, &settings.PointsPrice, &settings.PointsType, &settings.PublishPointsPrice, &settings.PublishPointsType, &cooldownSeconds, &settings.UpdatedAt, &settings.Version)
	settings.PublishCooldown = time.Duration(cooldownSeconds) * time.Second
	return settings, err
}
