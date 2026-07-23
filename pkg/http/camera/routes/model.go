// Package routes exposes protected camera administration.
package routes

import (
	"time"

	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
)

// AuditRequest stores required administrative attribution.
type AuditRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason stores human-readable attribution.
	Reason string `json:"reason"`
}

// SettingsRequest replaces camera prices and policy optimistically.
type SettingsRequest struct {
	AuditRequest
	// Enabled controls camera operations.
	Enabled bool `json:"enabled"`
	// CreditsPrice stores the purchase credit price.
	CreditsPrice int64 `json:"creditsPrice"`
	// PointsPrice stores the purchase points price.
	PointsPrice int64 `json:"pointsPrice"`
	// PointsType identifies the purchase points currency.
	PointsType int32 `json:"pointsType"`
	// PublishPointsPrice stores the publication price.
	PublishPointsPrice int64 `json:"publishPointsPrice"`
	// PublishPointsType identifies the publication currency.
	PublishPointsType int32 `json:"publishPointsType"`
	// PublishCooldownSeconds stores the publication cooldown.
	PublishCooldownSeconds int64 `json:"publishCooldownSeconds"`
	// Version stores the expected optimistic version.
	Version int64 `json:"version"`
}

// SettingsResponse exposes current camera prices and policy.
type SettingsResponse struct {
	// Enabled controls camera operations.
	Enabled bool `json:"enabled"`
	// CreditsPrice stores the purchase credit price.
	CreditsPrice int64 `json:"creditsPrice"`
	// PointsPrice stores the purchase points price.
	PointsPrice int64 `json:"pointsPrice"`
	// PointsType identifies the purchase points currency.
	PointsType int32 `json:"pointsType"`
	// PublishPointsPrice stores the publication price.
	PublishPointsPrice int64 `json:"publishPointsPrice"`
	// PublishPointsType identifies the publication currency.
	PublishPointsType int32 `json:"publishPointsType"`
	// PublishCooldownSeconds stores the publication cooldown.
	PublishCooldownSeconds int64 `json:"publishCooldownSeconds"`
	// UpdatedAt stores the last mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version stores optimistic state.
	Version int64 `json:"version"`
}

// settingsResponse maps one domain settings record.
func settingsResponse(settings camerarecord.Settings) SettingsResponse {
	return SettingsResponse{Enabled: settings.Enabled, CreditsPrice: settings.CreditsPrice, PointsPrice: settings.PointsPrice, PointsType: settings.PointsType, PublishPointsPrice: settings.PublishPointsPrice, PublishPointsType: settings.PublishPointsType, PublishCooldownSeconds: int64(settings.PublishCooldown / time.Second), UpdatedAt: settings.UpdatedAt, Version: settings.Version}
}

// settingsRecord maps one replacement request.
func settingsRecord(request SettingsRequest) camerarecord.Settings {
	return camerarecord.Settings{Enabled: request.Enabled, CreditsPrice: request.CreditsPrice, PointsPrice: request.PointsPrice, PointsType: request.PointsType, PublishPointsPrice: request.PublishPointsPrice, PublishPointsType: request.PublishPointsType, PublishCooldown: time.Duration(request.PublishCooldownSeconds) * time.Second}
}
