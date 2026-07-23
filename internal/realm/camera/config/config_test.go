package config

import (
	"testing"
	"time"
)

// TestLoadDefaultsAndOverrides verifies bounded runtime policy parsing.
func TestLoadDefaultsAndOverrides(t *testing.T) {
	t.Setenv("PIXELS_CAMERA_ENABLED", "false")
	t.Setenv("PIXELS_CAMERA_CAPTURE_COOLDOWN", "7s")
	t.Setenv("PIXELS_CAMERA_MAX_PHOTO_BYTES", "123")
	t.Setenv("PIXELS_CAMERA_MAX_THUMBNAIL_BYTES", "45")
	t.Setenv("PIXELS_CAMERA_PENDING_RETENTION", "2h")
	t.Setenv("PIXELS_CAMERA_SUPERSEDED_RETENTION", "30m")
	t.Setenv("PIXELS_CAMERA_CLEANUP_INTERVAL", "1m")
	t.Setenv("PIXELS_CAMERA_CLEANUP_RETRY", "2m")
	t.Setenv("PIXELS_CAMERA_CLEANUP_BATCH_SIZE", "12")
	config := Load()
	if config.Enabled || config.CaptureCooldown != 7*time.Second || config.MaxPhotoBytes != 123 || config.MaxThumbnailBytes != 45 || config.PendingRetention != 2*time.Hour || config.SupersededRetention != 30*time.Minute || config.CleanupInterval != time.Minute || config.CleanupRetry != 2*time.Minute || config.CleanupBatchSize != 12 {
		t.Fatalf("unexpected config: %+v", config)
	}
}

// TestLoadRejectsInvalidValues verifies safe fallback policy.
func TestLoadRejectsInvalidValues(t *testing.T) {
	t.Setenv("PIXELS_CAMERA_ENABLED", "invalid")
	t.Setenv("PIXELS_CAMERA_CAPTURE_COOLDOWN", "-1s")
	t.Setenv("PIXELS_CAMERA_MAX_PHOTO_BYTES", "0")
	t.Setenv("PIXELS_CAMERA_CLEANUP_INTERVAL", "0s")
	t.Setenv("PIXELS_CAMERA_CLEANUP_BATCH_SIZE", "0")
	config := Load()
	if !config.Enabled || config.CaptureCooldown != 3*time.Second || config.MaxPhotoBytes != 2<<20 || config.CleanupInterval != 5*time.Minute || config.CleanupBatchSize != 100 {
		t.Fatalf("unexpected fallbacks: %+v", config)
	}
}
