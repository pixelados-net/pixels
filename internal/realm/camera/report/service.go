// Package report validates photo evidence and delegates to moderation.
package report

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// Furniture reads photo items and definitions.
type Furniture interface {
	// FindItemByID finds one furniture instance.
	FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error)
	// FindDefinitionByID finds one furniture definition.
	FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error)
}

// Moderator accepts validated photo report parameters.
type Moderator interface {
	// Report creates one moderated call for help.
	Report(context.Context, moderationrecord.ReportParams) (moderationcore.ReportResult, error)
}

// Service resolves authoritative photo evidence for moderation.
type Service struct {
	// furniture reads item ownership and definitions.
	furniture Furniture
	// moderation owns issue creation and throttling.
	moderation Moderator
	// metrics records bounded report outcomes.
	metrics *camerametrics.Metrics
}

// New creates a camera photo report service.
func New(furniture Furniture, moderation Moderator, metrics *camerametrics.Metrics) *Service {
	return &Service{furniture: furniture, moderation: moderation, metrics: metrics}
}

// Submit validates one placed photo and creates its moderation report.
func (service *Service) Submit(ctx context.Context, reporterID int64, itemID int64, roomID int64, topicID int64, message string) (moderationcore.ReportResult, error) {
	item, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil {
		return moderationcore.ReportResult{}, err
	}
	if !found || item.RoomID == nil || *item.RoomID != roomID {
		return moderationcore.ReportResult{}, camerarecord.ErrInvalidPhoto
	}
	definition, found, err := service.furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return moderationcore.ReportResult{}, err
	}
	if !found || definition.InteractionType != "external_image" {
		return moderationcore.ReportResult{}, camerarecord.ErrInvalidPhoto
	}
	ownerID := item.OwnerPlayerID
	photoID := item.ID
	incidentRoomID := *item.RoomID
	message = strings.TrimSpace(message)
	if message == "" {
		message = fmt.Sprintf("photo:%d", photoID)
	}
	result, err := service.moderation.Report(ctx, moderationrecord.ReportParams{ReporterPlayerID: reporterID, ReportedPlayerID: &ownerID, RoomID: &incidentRoomID, PhotoItemID: &photoID, TopicID: topicID, Kind: "cfh_photo", Message: message})
	if err == nil {
		service.metrics.Reported()
	}
	return result, err
}

// SubmitSelfie resolves a legacy string item reference and submits it.
func (service *Service) SubmitSelfie(ctx context.Context, reporterID int64, extraData string, roomID int64, topicID int64, message string) (moderationcore.ReportResult, error) {
	itemID, err := strconv.ParseInt(strings.TrimSpace(extraData), 10, 64)
	if err != nil || itemID <= 0 {
		return moderationcore.ReportResult{}, camerarecord.ErrInvalidPhoto
	}
	return service.Submit(ctx, reporterID, itemID, roomID, topicID, message)
}
