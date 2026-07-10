package rights

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	rightsmodel "github.com/niflaot/pixels/internal/realm/room/rights/model"
	rightsrepo "github.com/niflaot/pixels/internal/realm/room/rights/repository"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// rightsStoreForTest stores room rights in memory.
type rightsStoreForTest struct {
	// rights stores rights by player id.
	rights map[int64]rightsmodel.Right
}

// WithinTransaction runs work directly.
func (store *rightsStoreForTest) WithinTransaction(ctx context.Context, work rightsrepo.TransactionWork) error {
	return work(ctx)
}

// Grant creates rights when absent.
func (store *rightsStoreForTest) Grant(_ context.Context, roomID int64, playerID int64, actorID int64) (bool, error) {
	if _, found := store.rights[playerID]; found {
		return false, nil
	}
	store.rights[playerID] = rightsmodel.Right{RoomID: roomID, PlayerID: playerID, GrantedByPlayerID: actorID}
	return true, nil
}

// Revoke removes one rights holder.
func (store *rightsStoreForTest) Revoke(_ context.Context, _ int64, playerID int64) (bool, error) {
	_, found := store.rights[playerID]
	delete(store.rights, playerID)
	return found, nil
}

// RevokeAll removes every rights holder.
func (store *rightsStoreForTest) RevokeAll(context.Context, int64) ([]rightsmodel.Right, error) {
	rights := store.ListValues()
	clear(store.rights)
	return rights, nil
}

// List returns current rights holders.
func (store *rightsStoreForTest) List(context.Context, int64) ([]rightsmodel.Right, error) {
	return store.ListValues(), nil
}

// Exists reports whether rights exist.
func (store *rightsStoreForTest) Exists(_ context.Context, _ int64, playerID int64) (bool, error) {
	_, found := store.rights[playerID]
	return found, nil
}

// ListValues returns in-memory rights.
func (store *rightsStoreForTest) ListValues() []rightsmodel.Right {
	rights := make([]rightsmodel.Right, 0, len(store.rights))
	for _, right := range store.rights {
		rights = append(rights, right)
	}
	return rights
}

// rightsRoomForTest returns one owned room.
type rightsRoomForTest struct {
	// room optionally replaces the default room.
	room roommodel.Room
	// found optionally reports a missing room when false with a nonzero room.
	found *bool
	// err optionally fails room lookup.
	err error
}

// FindByID finds one room.
func (finder rightsRoomForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	if finder.err != nil {
		return roommodel.Room{}, false, finder.err
	}
	if finder.found != nil && !*finder.found {
		return roommodel.Room{}, false, nil
	}
	if finder.room.ID > 0 {
		return finder.room, true, nil
	}
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 1}, true, nil
}

// rightsPermissionsForTest stores allowed permission nodes.
type rightsPermissionsForTest map[int64]map[permission.Node]bool

// HasPermission reports one configured permission decision.
func (permissions rightsPermissionsForTest) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	return permissions[playerID][node], nil
}

// rightsEventsForTest captures published events.
type rightsEventsForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures one event.
func (publisher *rightsEventsForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestServiceGrantAndRevokeAll verifies authorized rights lifecycle and events.
func TestServiceGrantAndRevokeAll(t *testing.T) {
	store := &rightsStoreForTest{rights: make(map[int64]rightsmodel.Right)}
	events := &rightsEventsForTest{}
	nodes := Nodes{OwnGrant: "own.grant", OwnRevoke: "own.revoke", AnyGrant: "any.grant", AnyRevoke: "any.revoke"}
	permissions := rightsPermissionsForTest{1: {nodes.OwnGrant: true, nodes.OwnRevoke: true}}
	service := New(store, rightsRoomForTest{}, permissions, events, nodes)

	if err := service.GrantRights(context.Background(), 9, 1, 2); err != nil {
		t.Fatalf("grant rights: %v", err)
	}
	if allowed, err := service.HasRights(context.Background(), 9, 2); err != nil || !allowed {
		t.Fatalf("expected rights, allowed=%v err=%v", allowed, err)
	}
	count, err := service.RevokeAllRights(context.Background(), 9, 1)
	if err != nil || count != 1 {
		t.Fatalf("revoke all: count=%d err=%v", count, err)
	}
	if len(events.events) != 2 {
		t.Fatalf("expected grant and revoke events, got %#v", events.events)
	}
}

// TestServiceRejectsUnauthorizedGrant verifies owner nodes remain revocable.
func TestServiceRejectsUnauthorizedGrant(t *testing.T) {
	service := New(&rightsStoreForTest{rights: make(map[int64]rightsmodel.Right)}, rightsRoomForTest{}, rightsPermissionsForTest{}, nil, Nodes{OwnGrant: "own.grant", AnyGrant: "any.grant"})

	err := service.GrantRights(context.Background(), 9, 1, 2)
	if !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
}

// TestServiceAllowsStaffAndRelinquish verifies global grants and unconditional self-revoke.
func TestServiceAllowsStaffAndRelinquish(t *testing.T) {
	store := &rightsStoreForTest{rights: make(map[int64]rightsmodel.Right)}
	nodes := Nodes{AnyGrant: "any.grant"}
	service := New(store, rightsRoomForTest{}, rightsPermissionsForTest{3: {nodes.AnyGrant: true}}, nil, nodes)

	if err := service.GrantRights(context.Background(), 9, 3, 2); err != nil {
		t.Fatalf("staff grant: %v", err)
	}
	if err := service.RelinquishRights(context.Background(), 9, 2); err != nil {
		t.Fatalf("relinquish: %v", err)
	}
}

// TestServiceValidatesTargetsAndMissingRooms verifies identity and ownership boundaries.
func TestServiceValidatesTargetsAndMissingRooms(t *testing.T) {
	store := &rightsStoreForTest{rights: make(map[int64]rightsmodel.Right)}
	nodes := Nodes{OwnGrant: "own.grant", AnyGrant: "any.grant"}
	permissions := rightsPermissionsForTest{1: {nodes.OwnGrant: true}}
	service := New(store, rightsRoomForTest{}, permissions, nil, nodes)
	if err := service.GrantRights(context.Background(), 9, 1, 1); !errors.Is(err, ErrOwnerTarget) {
		t.Fatalf("expected owner target, got %v", err)
	}
	if _, err := service.ListRights(context.Background(), 0); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("expected invalid identity, got %v", err)
	}
	found := false
	service = New(store, rightsRoomForTest{found: &found}, permissions, nil, nodes)
	if err := service.GrantRights(context.Background(), 9, 1, 2); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected room not found, got %v", err)
	}
}

// TestServiceRevokesOneRightAndListsCurrentState verifies explicit revocation.
func TestServiceRevokesOneRightAndListsCurrentState(t *testing.T) {
	store := &rightsStoreForTest{rights: map[int64]rightsmodel.Right{2: {RoomID: 9, PlayerID: 2}}}
	nodes := Nodes{OwnRevoke: "own.revoke"}
	service := New(store, rightsRoomForTest{}, rightsPermissionsForTest{1: {nodes.OwnRevoke: true}}, nil, nodes)
	if err := service.RevokeRights(context.Background(), 9, 1, 2); err != nil {
		t.Fatalf("revoke right: %v", err)
	}
	rights, err := service.ListRights(context.Background(), 9)
	if err != nil || len(rights) != 0 {
		t.Fatalf("unexpected rights %#v err=%v", rights, err)
	}
	if allowed, err := service.HasRights(context.Background(), 0, 2); err != nil || allowed {
		t.Fatalf("invalid lookup allowed=%v err=%v", allowed, err)
	}
}

// BenchmarkHasRights measures allocation-free rights resolution above repository I/O.
func BenchmarkHasRights(b *testing.B) {
	store := &rightsStoreForTest{rights: map[int64]rightsmodel.Right{2: {RoomID: 9, PlayerID: 2}}}
	service := New(store, rightsRoomForTest{}, nil, nil, Nodes{})
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		allowed, err := service.HasRights(ctx, 9, 2)
		if err != nil || !allowed {
			b.Fatal("rights lookup failed")
		}
	}
}
