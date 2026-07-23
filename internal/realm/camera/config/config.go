// Package config loads runtime camera upload policy.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config stores camera upload limits and throttling policy.
type Config struct {
	// Enabled controls camera handlers.
	Enabled bool
	// CaptureCooldown limits consecutive photo uploads.
	CaptureCooldown time.Duration
	// MaxPhotoBytes limits photo payloads.
	MaxPhotoBytes int
	// MaxThumbnailBytes limits thumbnail payloads.
	MaxThumbnailBytes int
	// PendingRetention limits how long an unused active photo is retained.
	PendingRetention time.Duration
	// SupersededRetention delays deletion of replaced unreferenced photos.
	SupersededRetention time.Duration
	// CleanupInterval controls aggregate cleanup cadence.
	CleanupInterval time.Duration
	// CleanupRetry delays another deletion attempt after storage failure.
	CleanupRetry time.Duration
	// CleanupBatchSize bounds one reconciliation query.
	CleanupBatchSize int
}

// Load reads camera environment settings with safe defaults.
func Load() Config {
	return Config{
		Enabled:             envBool("PIXELS_CAMERA_ENABLED", true),
		CaptureCooldown:     envDuration("PIXELS_CAMERA_CAPTURE_COOLDOWN", 3*time.Second),
		MaxPhotoBytes:       envInt("PIXELS_CAMERA_MAX_PHOTO_BYTES", 2<<20),
		MaxThumbnailBytes:   envInt("PIXELS_CAMERA_MAX_THUMBNAIL_BYTES", 1<<20),
		PendingRetention:    envPositiveDuration("PIXELS_CAMERA_PENDING_RETENTION", 24*time.Hour),
		SupersededRetention: envPositiveDuration("PIXELS_CAMERA_SUPERSEDED_RETENTION", time.Hour),
		CleanupInterval:     envPositiveDuration("PIXELS_CAMERA_CLEANUP_INTERVAL", 5*time.Minute),
		CleanupRetry:        envPositiveDuration("PIXELS_CAMERA_CLEANUP_RETRY", 5*time.Minute),
		CleanupBatchSize:    envInt("PIXELS_CAMERA_CLEANUP_BATCH_SIZE", 100),
	}
}

// envBool reads one optional boolean environment value.
func envBool(name string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	if err != nil {
		return fallback
	}
	return value
}

// envDuration reads one optional positive duration environment value.
func envDuration(name string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(name))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

// envPositiveDuration reads one optional strictly positive duration value.
func envPositiveDuration(name string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// envInt reads one optional positive integer environment value.
func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
