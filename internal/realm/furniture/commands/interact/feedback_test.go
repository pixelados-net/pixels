package interact

import (
	"context"
	"testing"

	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	"github.com/niflaot/pixels/pkg/i18n"
)

// TestActionMatchesDedicatedFurnitureTypes verifies packet-type misuse is rejected.
func TestActionMatchesDedicatedFurnitureTypes(t *testing.T) {
	tests := []struct {
		action Action
		kind   string
		match  bool
	}{
		{action: ActionUse, kind: "cannon", match: true},
		{action: ActionDice, kind: "dice", match: true},
		{action: ActionDiceClose, kind: "cannon", match: false},
		{action: ActionColorWheel, kind: "colorwheel", match: true},
		{action: ActionColorWheel, kind: "dice", match: false},
	}
	for _, test := range tests {
		if matched := actionMatches(test.action, test.kind); matched != test.match {
			t.Fatalf("action=%d kind=%s match=%t", test.action, test.kind, matched)
		}
	}
}

// TestBroadcastNormalizesMalformedState verifies defensive protocol projection.
func TestBroadcastNormalizesMalformedState(t *testing.T) {
	handler, _, sent, active := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), Definition: worldfurniture.Definition{Width: 1, Length: 1},
	})
	if err := handler.broadcast(context.Background(), active, 1, "invalid"); err != nil {
		t.Fatalf("broadcast normalized state: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outstate.Header {
		t.Fatalf("unexpected state packets %#v", *sent)
	}
}

// TestStartTeleportIgnoresInvalidUse verifies non-teleport use remains a soft miss.
func TestStartTeleportIgnoresInvalidUse(t *testing.T) {
	handler := Handler{Teleports: &teleporterForTest{err: teleport.ErrInvalidUse}}
	if err := handler.startTeleport(context.Background(), 7, nil, 1); err != nil {
		t.Fatalf("ignore invalid teleport use: %v", err)
	}
}

// TestSendNoRightsUsesTranslationAndResyncIgnoresMissingItem verifies feedback branches.
func TestSendNoRightsUsesTranslationAndResyncIgnoresMissingItem(t *testing.T) {
	handler, connection, _, active := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), Definition: worldfurniture.Definition{Width: 1, Length: 1},
	})
	handler.Translations = i18n.NewCatalog(i18n.Config{}, map[i18n.Locale]map[i18n.Key]string{
		"en": {noRightsTranslationKey: "No rights"},
	})
	if err := handler.sendNoRights(context.Background(), connection); err != nil {
		t.Fatalf("send translated feedback: %v", err)
	}
	if err := handler.resync(context.Background(), active, 99, "gate"); err != nil {
		t.Fatalf("ignore missing resync item: %v", err)
	}
}
