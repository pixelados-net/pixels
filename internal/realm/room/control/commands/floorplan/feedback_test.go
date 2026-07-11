package floorplan

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// cooldownForTest returns a configured acquisition result.
type cooldownForTest struct {
	// acquired reports whether the key was reserved.
	acquired bool
	// deleted reports whether a failed save released its key.
	deleted bool
}

// SetIfAbsent returns the configured acquisition result.
func (cooldown *cooldownForTest) SetIfAbsent(context.Context, string, []byte, time.Duration) (bool, error) {
	return cooldown.acquired, nil
}

// Delete records cooldown release.
func (cooldown *cooldownForTest) Delete(context.Context, string) error {
	cooldown.deleted = true

	return nil
}

// TestSaveHandleSendsAccessAndCooldownFeedback verifies expected failures stay soft.
func TestSaveHandleSendsAccessAndCooldownFeedback(t *testing.T) {
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, ModelName: "model_a"}
	tests := []struct {
		name          string
		permissions   permissionsForTest
		cooldown      *cooldownForTest
		expectRelease bool
	}{
		{name: "denied", permissions: permissionsForTest{}},
		{name: "cooldown", permissions: permissionsForTest{"own": true}, cooldown: &cooldownForTest{}},
		{name: "invalid", permissions: permissionsForTest{"own": true}, cooldown: &cooldownForTest{acquired: true}, expectRelease: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			players, bindings, player := floorplanActorForTest(t)
			if err := player.EnterRoom(9); err != nil {
				t.Fatalf("enter room: %v", err)
			}
			connection, sent := floorplanConnectionContextForTest(t)
			authorizer := domain.NewAuthorizer(test.permissions, nil, domain.Nodes{OwnEdit: "own", AnyEdit: "any"})
			handler := SaveHandler{
				Players: players, Bindings: bindings, Rooms: roomsForTest{room: room},
				Layouts: &layoutsForTest{}, Runtime: roomlive.NewRegistry(nil),
				Authorize: authorizer, Cooldowns: test.cooldown,
			}
			err := handler.Handle(context.Background(), command.Envelope[SaveCommand]{Command: SaveCommand{Handler: connection}})
			if err != nil {
				t.Fatalf("expected soft error, got %v", err)
			}
			if len(*sent) != 1 || (*sent)[0].Header != 1992 {
				t.Fatalf("unexpected feedback %#v", *sent)
			}
			if test.expectRelease && !test.cooldown.deleted {
				t.Fatal("expected failed save to release cooldown")
			}
		})
	}
}

// TestErrorKeysRejectsUnexpectedFailures verifies technical errors remain explicit.
func TestErrorKeysRejectsUnexpectedFailures(t *testing.T) {
	if keys := errorKeys(errors.New("database unavailable")); keys != nil {
		t.Fatalf("unexpected technical error mapping %#v", keys)
	}
	validation := domain.ValidationErrors{Codes: []domain.ErrorCode{domain.CodeInvalidMap, domain.CodeInvalidDoor}}
	if keys := errorKeys(validation); len(keys) != 2 {
		t.Fatalf("unexpected validation mapping %#v", keys)
	}
}
