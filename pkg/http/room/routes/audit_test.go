package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	roomaudit "github.com/niflaot/pixels/internal/realm/room/control/audit"
	auditmodel "github.com/niflaot/pixels/internal/realm/room/control/audit/model"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
)

// auditManagerForTest captures audit queries.
type auditManagerForTest struct {
	// query stores the latest query.
	query roomaudit.Query
}

// ModerationHistory returns one moderation row.
func (manager *auditManagerForTest) ModerationHistory(_ context.Context, query roomaudit.Query) ([]auditmodel.ModerationAction, error) {
	manager.query = query
	for _, action := range query.ActionTypes {
		if !action.Valid() {
			return nil, roomaudit.ErrInvalidQuery
		}
	}
	return []auditmodel.ModerationAction{{ID: 3, RoomID: 9, TargetPlayerID: 2, Action: moderationmodel.ActionKick}}, nil
}

// RightsHistory returns one rights row.
func (manager *auditManagerForTest) RightsHistory(_ context.Context, query roomaudit.Query) ([]auditmodel.RightsAudit, error) {
	manager.query = query
	return []auditmodel.RightsAudit{{ID: 4, RoomID: 9, PlayerID: 2}}, nil
}

// moderationReaderForTest returns active sanctions.
type moderationReaderForTest struct{}

// IsBanned reports no active ban.
func (moderationReaderForTest) IsBanned(context.Context, int64, int64) (bool, error) {
	return false, nil
}

// IsMuted reports no active mute.
func (moderationReaderForTest) IsMuted(context.Context, int64, int64) (bool, error) {
	return false, nil
}

// ListBans returns one active ban.
func (moderationReaderForTest) ListBans(context.Context, int64) ([]moderationmodel.Sanction, error) {
	return []moderationmodel.Sanction{{RoomID: 9, PlayerID: 2, Username: "Alice", EndsAt: time.Unix(2000, 0)}}, nil
}

// ListMutes returns one active mute.
func (moderationReaderForTest) ListMutes(context.Context, int64) ([]moderationmodel.Sanction, error) {
	return []moderationmodel.Sanction{{RoomID: 9, PlayerID: 3, Username: "Bob", EndsAt: time.Unix(2000, 0)}}, nil
}

// TestAuditRoutesMapRoomAndPlayerFilters verifies route query mapping.
func TestAuditRoutesMapRoomAndPlayerFilters(t *testing.T) {
	manager := &auditManagerForTest{}
	app := fiber.New()
	registerAudit(app, Dependencies{Audit: manager, Moderation: moderationReaderForTest{}})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/rooms/9/moderation/history?type=kick,ban&before=8&limit=25", nil))
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("room history status=%v err=%v", response.StatusCode, err)
	}
	if manager.query.RoomID == nil || *manager.query.RoomID != 9 || len(manager.query.ActionTypes) != 2 || manager.query.Limit != 25 {
		t.Fatalf("unexpected room query %#v", manager.query)
	}
	response, err = app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/players/2/moderation/actions?roomId=9", nil))
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("player history status=%v err=%v", response.StatusCode, err)
	}
	if manager.query.ActorPlayerID == nil || *manager.query.ActorPlayerID != 2 || manager.query.RoomID == nil || *manager.query.RoomID != 9 {
		t.Fatalf("unexpected player query %#v", manager.query)
	}
}

// TestAuditRoutesReturnSanctionsAndRejectInvalidIDs verifies current-state routes.
func TestAuditRoutesReturnSanctionsAndRejectInvalidIDs(t *testing.T) {
	app := fiber.New()
	registerAudit(app, Dependencies{Audit: &auditManagerForTest{}, Moderation: moderationReaderForTest{}})
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/rooms/9/bans", nil))
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("bans status=%v err=%v", response.StatusCode, err)
	}
	response, err = app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/rooms/nope/mutes", nil))
	if err != nil || response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid room status=%v err=%v", response.StatusCode, err)
	}
	response, err = app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/rooms/9/moderation/history?type=nope", nil))
	if err != nil || response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid action status=%v err=%v", response.StatusCode, err)
	}
}
