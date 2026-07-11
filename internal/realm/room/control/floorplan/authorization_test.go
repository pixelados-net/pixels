package floorplan

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// permissionCheckerForTest resolves configured permission nodes.
type permissionCheckerForTest map[permission.Node]bool

// HasPermission resolves one configured node.
func (checker permissionCheckerForTest) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return checker[node], nil
}

// rightsCheckerForTest stores one local rights result.
type rightsCheckerForTest bool

// HasRights reports the configured local rights result.
func (checker rightsCheckerForTest) HasRights(context.Context, int64, int64) (bool, error) {
	return bool(checker), nil
}

// TestAuthorizerSupportsOwnerRightsAndGlobalCapability verifies floor plan authorization paths.
func TestAuthorizerSupportsOwnerRightsAndGlobalCapability(t *testing.T) {
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7}
	tests := []struct {
		name        string
		actorID     int64
		permissions permissionCheckerForTest
		rights      rightsCheckerForTest
		allowed     bool
	}{
		{name: "owner", actorID: 7, permissions: permissionCheckerForTest{"own": true}, allowed: true},
		{name: "rights", actorID: 8, permissions: permissionCheckerForTest{"own": true}, rights: true, allowed: true},
		{name: "staff", actorID: 8, permissions: permissionCheckerForTest{"any": true}, allowed: true},
		{name: "denied", actorID: 8, permissions: permissionCheckerForTest{"own": true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authorizer := NewAuthorizer(test.permissions, test.rights, Nodes{OwnEdit: "own", AnyEdit: "any"})
			err := authorizer.Authorize(context.Background(), room, test.actorID)
			if test.allowed && err != nil {
				t.Fatalf("authorize: %v", err)
			}
			if !test.allowed && !errors.Is(err, ErrAccessDenied) {
				t.Fatalf("expected access denied, got %v", err)
			}
		})
	}
}
