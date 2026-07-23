package report

import (
	"context"
	"errors"
	"testing"

	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestSubmitUsesAuthoritativePhotoOwner verifies client target ids are unnecessary.
func TestSubmitUsesAuthoritativePhotoOwner(t *testing.T) {
	roomID := int64(40)
	moderator := &reportModerator{}
	metrics := camerametrics.New()
	service := New(reportFurniture{item: furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, DefinitionID: 4, OwnerPlayerID: 12, RoomID: &roomID}}, moderator, metrics)
	if _, err := service.Submit(context.Background(), 7, 9, 40, 3, "evidence"); err != nil {
		t.Fatalf("submit report: %v", err)
	}
	if moderator.params.ReportedPlayerID == nil || *moderator.params.ReportedPlayerID != 12 || moderator.params.PhotoItemID == nil || *moderator.params.PhotoItemID != 9 || moderator.params.Kind != "cfh_photo" {
		t.Fatalf("unexpected report parameters: %+v", moderator.params)
	}
	if metrics.Snapshot().Reports != 1 {
		t.Fatalf("report metric missing: %+v", metrics.Snapshot())
	}
}

// TestSubmitRejectsUnplacedAndWrongDefinition verifies evidence integrity.
func TestSubmitRejectsUnplacedAndWrongDefinition(t *testing.T) {
	service := New(reportFurniture{item: furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, DefinitionID: 4}}, &reportModerator{}, camerametrics.New())
	if _, err := service.Submit(context.Background(), 7, 9, 40, 3, ""); !errors.Is(err, camerarecord.ErrInvalidPhoto) {
		t.Fatalf("expected unplaced rejection, got %v", err)
	}
	if _, err := service.SubmitSelfie(context.Background(), 7, "invalid", 40, 3, ""); !errors.Is(err, camerarecord.ErrInvalidPhoto) {
		t.Fatalf("expected selfie reference rejection, got %v", err)
	}
}

// reportFurniture stores one placed photo fixture.
type reportFurniture struct {
	// item stores the reported furniture fixture.
	item furnituremodel.Item
}

// FindItemByID returns the configured item.
func (furniture reportFurniture) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return furniture.item, true, nil
}

// FindDefinitionByID returns an external-image definition.
func (reportFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{InteractionType: "external_image"}, true, nil
}

// reportModerator stores submitted moderation parameters.
type reportModerator struct {
	// params stores the delegated report parameters.
	params moderationrecord.ReportParams
}

// Report stores one report request.
func (moderator *reportModerator) Report(_ context.Context, params moderationrecord.ReportParams) (moderationcore.ReportResult, error) {
	moderator.params = params
	return moderationcore.ReportResult{}, nil
}
