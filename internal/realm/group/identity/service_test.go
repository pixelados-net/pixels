package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	"github.com/niflaot/pixels/internal/realm/group/badge"
	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	createdevent "github.com/niflaot/pixels/internal/realm/group/identity/events/created"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// identityStore supplies deterministic identity persistence behavior.
type identityStore struct {
	// Store supplies unused persistence methods.
	grouprecord.Store
	// group stores the deterministic group fixture.
	group grouprecord.Group
	// replayed controls administrative idempotency behavior.
	replayed bool
	// insertCalls counts durable group inserts.
	insertCalls int
	// rooms stores deterministic creator room options.
	rooms []grouprecord.EligibleRoom
}

// WithinTransaction executes one deterministic transaction body.
func (store *identityStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// BadgeRegistry returns one complete minimal editor registry.
func (store *identityStore) BadgeRegistry(context.Context) ([]grouprecord.BadgeElement, []grouprecord.BadgeColor, error) {
	return []grouprecord.BadgeElement{{Kind: grouprecord.BadgeBase, ID: 1}, {Kind: grouprecord.BadgeSymbol, ID: 2}}, []grouprecord.BadgeColor{{Family: grouprecord.BaseColor, ID: 1, Hex: "FFFFFF"}, {Family: grouprecord.SymbolColor, ID: 1, Hex: "FFFFFF"}, {Family: grouprecord.BackgroundColor, ID: 1, Hex: "112233"}, {Family: grouprecord.BackgroundColor, ID: 2, Hex: "AABBCC"}}, nil
}

// EligibleRooms returns deterministic creator room options.
func (store *identityStore) EligibleRooms(context.Context, int64) ([]grouprecord.EligibleRoom, error) {
	return store.rooms, nil
}

// LockEligibleRoom accepts the configured owner room.
func (store *identityStore) LockEligibleRoom(context.Context, int64, int64) error { return nil }

// CountOwned returns no existing owned groups.
func (store *identityStore) CountOwned(context.Context, int64) (int, error) { return 0, nil }

// CountMemberships returns no existing memberships.
func (store *identityStore) CountMemberships(context.Context, int64) (int, error) { return 0, nil }

// ClaimCreateOperation returns the configured replay state.
func (store *identityStore) ClaimCreateOperation(context.Context, string, string) (int64, bool, error) {
	return store.group.ID, store.replayed, nil
}

// CompleteCreateOperation accepts one claimed creation.
func (store *identityStore) CompleteCreateOperation(context.Context, string, int64) error { return nil }

// InsertGroup returns one created group.
func (store *identityStore) InsertGroup(_ context.Context, params grouprecord.CreateParams) (grouprecord.Group, error) {
	store.insertCalls++
	store.group.Name = params.Name
	return store.group, nil
}

// Group returns the configured group.
func (store *identityStore) Group(context.Context, int64, bool) (grouprecord.Group, bool, error) {
	return store.group, store.group.ID > 0, nil
}

// BadgeParts returns an empty retained badge projection.
func (store *identityStore) BadgeParts(context.Context, int64) ([]grouprecord.BadgePart, error) {
	return nil, nil
}

// PlayerGroups returns an empty player projection.
func (store *identityStore) PlayerGroups(context.Context, int64) ([]grouprecord.PlayerGroup, error) {
	return nil, nil
}

// identityPlayers returns one deterministic player.
type identityPlayers struct {
	// record stores the deterministic player fixture.
	record playerservice.Record
	// found controls player lookup behavior.
	found bool
}

// FindByID returns the configured player.
func (players identityPlayers) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return players.record, players.found, nil
}

// FindByUsername returns no alternate player.
func (players identityPlayers) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// identityCurrencies records the last signed currency mutation.
type identityCurrencies struct {
	// grant stores the latest signed currency mutation.
	grant currencyservice.GrantParams
}

// Grant records one deterministic currency mutation.
func (currencies *identityCurrencies) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currencies.grant = params
	return 90, nil
}

// identityPermissions returns one configured override decision.
type identityPermissions bool

// HasPermission returns the configured decision.
func (allowed identityPermissions) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return bool(allowed), nil
}

// TestCreateChargesAndCreatesAtomically verifies the validated creation workflow.
func TestCreateChargesAndCreatesAtomically(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	store := &identityStore{group: grouprecord.Group{ID: 9, OwnerPlayerID: 7, Version: 1}}
	registry := badge.New(store)
	if err := registry.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	currencies := &identityCurrencies{}
	events := bus.New()
	var published createdevent.Payload
	_, _ = events.Subscribe(createdevent.Name, 0, func(_ context.Context, event bus.Event) error {
		published = event.Payload.(createdevent.Payload)
		return nil
	})
	service := New(groupconfig.Config{CreationCost: 10, RequireClub: true, OwnedLimit: 2, MembershipLimit: 2}, store, badge.NewCompiler(registry), registry, currencies, identityPlayers{record: playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, Club: playermodel.Club{Level: playermodel.ClubLevelHC, ExpiresAt: &expires}}}, found: true}, identityPermissions(true), nil, groupruntime.NewCache(), nil, nil, events)
	created, err := service.Create(context.Background(), CreateParams{OwnerPlayerID: 7, Name: " Pixels ", HomeRoomID: 3, ColorA: 1, ColorB: 2, BadgeParts: []grouprecord.BadgePart{{Kind: grouprecord.BadgeBase, ElementID: 1, ColorID: 1}}})
	if err != nil || created.ID != 9 || created.Name != "Pixels" || currencies.grant.Amount != -10 || store.insertCalls != 1 || published.GroupID != 9 {
		t.Fatalf("created=%#v grant=%#v calls=%d event=%#v err=%v", created, currencies.grant, store.insertCalls, published, err)
	}
}

// TestCreateRejectsExpiredClubBeforeCharge verifies server-side entitlement enforcement.
func TestCreateRejectsExpiredClubBeforeCharge(t *testing.T) {
	store := &identityStore{}
	registry := badge.New(store)
	_ = registry.Refresh(context.Background())
	currencies := &identityCurrencies{}
	service := New(groupconfig.Config{RequireClub: true}, store, badge.NewCompiler(registry), registry, currencies, identityPlayers{record: playerservice.Record{}, found: true}, identityPermissions(true), nil, groupruntime.NewCache(), nil, nil, nil)
	_, err := service.Create(context.Background(), CreateParams{OwnerPlayerID: 7, Name: "Pixels", HomeRoomID: 3, ColorA: 1, ColorB: 2, BadgeParts: []grouprecord.BadgePart{{Kind: grouprecord.BadgeBase, ElementID: 1, ColorID: 1}}})
	if !errors.Is(err, grouprecord.ErrForbidden) || currencies.grant.Amount != 0 {
		t.Fatalf("grant=%#v err=%v", currencies.grant, err)
	}
}

// TestAdministrativeCreateReplaysWithoutCharge verifies durable idempotency replay.
func TestAdministrativeCreateReplaysWithoutCharge(t *testing.T) {
	store := &identityStore{group: grouprecord.Group{ID: 9, OwnerPlayerID: 7}, replayed: true}
	registry := badge.New(store)
	_ = registry.Refresh(context.Background())
	service := New(groupconfig.Config{OwnedLimit: 2, MembershipLimit: 2}, store, badge.NewCompiler(registry), registry, &identityCurrencies{}, identityPlayers{record: playerservice.Record{}, found: true}, nil, nil, groupruntime.NewCache(), nil, nil, nil)
	ctx := grouprecord.WithAudit(context.Background(), 1, "qa replay")
	group, replayed, err := service.CreateAdministrative(ctx, AdministrativeCreateParams{CreateParams: CreateParams{OwnerPlayerID: 7, Name: "Pixels", HomeRoomID: 3, ColorA: 1, ColorB: 2, BadgeParts: []grouprecord.BadgePart{{Kind: grouprecord.BadgeBase, ElementID: 1, ColorID: 1}}}, IdempotencyKey: "groups-qa-replay", Charge: false})
	if err != nil || !replayed || group.ID != 9 || store.insertCalls != 0 {
		t.Fatalf("group=%#v replayed=%v calls=%d err=%v", group, replayed, store.insertCalls, err)
	}
}

// TestOptionsAlwaysReturnsConfiguredCreatorData verifies groups cannot be disabled by configuration.
func TestOptionsAlwaysReturnsConfiguredCreatorData(t *testing.T) {
	store := &identityStore{rooms: []grouprecord.EligibleRoom{{ID: 130, Name: "GROUPS QA Creator"}}}
	service := New(groupconfig.Config{CreationCost: 10}, store, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	cost, rooms, err := service.Options(context.Background(), 1)
	if err != nil || cost != 10 || len(rooms) != 1 || rooms[0].ID != 130 {
		t.Fatalf("cost=%d rooms=%#v err=%v", cost, rooms, err)
	}
}
