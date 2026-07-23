package entry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/niflaot/pixels/internal/permission"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/model"
	redisstore "github.com/niflaot/pixels/pkg/redis"
	"golang.org/x/crypto/bcrypt"
)

// fixedPermissions resolves configured permission nodes.
type fixedPermissions map[permission.Node]bool

// HasPermission resolves one configured permission decision.
func (permissions fixedPermissions) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return permissions[node], nil
}

// fixedRights resolves one room-rights decision.
type fixedRights bool

// HasRights returns the configured room-rights decision.
func (rights fixedRights) HasRights(context.Context, int64, int64) (bool, error) {
	return bool(rights), nil
}

// fixedBans resolves one room-ban decision.
type fixedBans bool

// IsBanned returns the configured room-ban decision.
func (bans fixedBans) IsBanned(context.Context, int64, int64) (bool, error) {
	return bool(bans), nil
}

// TestAuthorizeFastPaths verifies open, owner, invisible, and global permission paths.
func TestAuthorizeFastPaths(t *testing.T) {
	nodes := Nodes{EnterAny: "room.enter.any", EnterFull: "room.enter.full", AnswerAnyDoorbell: "room.doorbell.answer.any"}
	service := New(Config{}, nil, fixedPermissions{nodes.EnterAny: true, nodes.EnterFull: true, nodes.AnswerAnyDoorbell: true}, nil, nodes)
	open := roomRecord(DoorMode(roommodel.DoorModeOpen))
	if _, err := service.Authorize(context.Background(), Request{Room: open, PlayerID: 8}); err != nil {
		t.Fatalf("authorize open room: %v", err)
	}
	invisible := roomRecord(DoorMode(roommodel.DoorModeInvisible))
	if _, err := service.Authorize(context.Background(), Request{Room: invisible, PlayerID: 8}); err != nil {
		t.Fatalf("authorize global bypass: %v", err)
	}
	allowed, err := service.CanEnterFull(context.Background(), 8)
	if err != nil || !allowed {
		t.Fatalf("expected capacity bypass allowed=%v err=%v", allowed, err)
	}
	allowed, err = service.CanAnswerDoorbell(context.Background(), 9, 7, 8)
	if err != nil || !allowed {
		t.Fatalf("expected global doorbell permission allowed=%v err=%v", allowed, err)
	}
	denied := New(Config{}, nil, nil, nil, nodes)
	if _, err := denied.Authorize(context.Background(), Request{Room: invisible, PlayerID: 8}); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
	if _, err := denied.Authorize(context.Background(), Request{Room: invisible, PlayerID: 7}); err != nil {
		t.Fatalf("authorize owner: %v", err)
	}
	doorbell := roomRecord(DoorMode(roommodel.DoorModeDoorbell))
	if _, err := denied.Authorize(context.Background(), Request{Room: doorbell, PlayerID: 8}); !errors.Is(err, ErrDoorbellRequired) {
		t.Fatalf("expected doorbell requirement, got %v", err)
	}
}

// TestAuthorizeRightsTrustAndBans verifies scoped and server-controlled bypasses.
func TestAuthorizeRightsTrustAndBans(t *testing.T) {
	room := roomRecord(DoorMode(roommodel.DoorModePassword))
	service := New(Config{TrustedTTL: time.Minute}, nil, nil, nil, Nodes{}).WithRights(fixedRights(true))
	if _, err := service.Authorize(context.Background(), Request{Room: room, PlayerID: 8}); err != nil {
		t.Fatalf("authorize rights holder: %v", err)
	}
	service = New(Config{TrustedTTL: time.Minute}, nil, nil, nil, Nodes{})
	if !service.GrantTrusted(8, room.ID) {
		t.Fatal("expected trusted grant")
	}
	if _, err := service.Authorize(context.Background(), Request{Room: room, PlayerID: 8}); err != nil {
		t.Fatalf("authorize trusted entry: %v", err)
	}
	banned := New(Config{}, nil, nil, nil, Nodes{}).WithBans(fixedBans(true))
	if _, err := banned.Authorize(context.Background(), Request{Room: room, PlayerID: 8, Trusted: true}); !errors.Is(err, ErrBanned) {
		t.Fatalf("expected ban before trust, got %v", err)
	}
	allowed := New(Config{}, nil, fixedPermissions{"room.enter.any": true}, nil, Nodes{EnterAny: "room.enter.any"}).WithBans(fixedBans(true))
	if _, err := allowed.Authorize(context.Background(), Request{Room: room, PlayerID: 8}); err != nil {
		t.Fatalf("authorize global ban bypass: %v", err)
	}
}

// TestServiceAccessorsAndPasswordWithoutRedis verifies lightweight optional dependencies.
func TestServiceAccessorsAndPasswordWithoutRedis(t *testing.T) {
	hash, err := HashPassword("correct", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	service := New(Config{MaxPasswordAttempts: 3}, nil, nil, nil, Nodes{})
	if service.Config().MaxPasswordAttempts != 3 {
		t.Fatalf("unexpected config %#v", service.Config())
	}
	room := roomRecord(DoorMode(roommodel.DoorModePassword))
	room.PasswordHash = &hash
	if _, err := service.Authorize(context.Background(), Request{Room: room, PlayerID: 8, Password: "correct"}); err != nil {
		t.Fatalf("authorize password without redis: %v", err)
	}
	if _, err := service.Authorize(context.Background(), Request{Room: room, PlayerID: 8, Password: "wrong"}); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("expected wrong password, got %v", err)
	}
	allowed, err := service.CanAnswerDoorbell(context.Background(), room.ID, 7, 7)
	if err != nil || !allowed {
		t.Fatalf("expected owner answer permission allowed=%v err=%v", allowed, err)
	}
}

// TestAuthorizePasswordLocksAndRecovers verifies Redis attempt and lockout behavior.
func TestAuthorizePasswordLocksAndRecovers(t *testing.T) {
	server := miniredis.RunT(t)
	client := redisstore.New(redisstore.Config{Address: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	hash, err := HashPassword("correct", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	translations := i18n.NewCatalog(i18n.Config{DefaultLocale: "en"}, map[i18n.Locale]map[i18n.Key]string{
		"en": {
			"room.entry.locked":     "Locked for {duration}.",
			"duration.minute.other": "{count} minutes",
		},
	})
	config := Config{MaxPasswordAttempts: 2, AttemptWindow: 5 * time.Minute, LockoutSeconds: 600, PasswordCost: bcrypt.MinCost}
	service := New(config, client, nil, translations, Nodes{})
	room := roomRecord(DoorMode(roommodel.DoorModePassword))
	room.PasswordHash = &hash
	request := Request{Room: room, PlayerID: 8, Password: "wrong"}
	if _, err := service.Authorize(context.Background(), request); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("expected wrong password, got %v", err)
	}
	result, err := service.Authorize(context.Background(), request)
	if !errors.Is(err, ErrEntryLocked) || result.Alert == "" {
		t.Fatalf("expected new lockout and alert result=%#v err=%v", result, err)
	}
	if result.Alert != "Locked for 10 minutes." {
		t.Fatalf("unexpected localized lockout %q", result.Alert)
	}
	request.Password = "correct"
	if _, err := service.Authorize(context.Background(), request); !errors.Is(err, ErrEntryLocked) {
		t.Fatalf("expected active lockout, got %v", err)
	}
	server.FastForward(11 * time.Minute)
	if _, err := service.Authorize(context.Background(), request); err != nil {
		t.Fatalf("authorize after expiration: %v", err)
	}
	if server.Exists(attemptKey(room.ID, 8)) {
		t.Fatal("expected successful password to clear attempts")
	}
}

// TestAuthorizePasswordResetsAttemptsAfterShortLockout verifies fresh post-lockout counting.
func TestAuthorizePasswordResetsAttemptsAfterShortLockout(t *testing.T) {
	server := miniredis.RunT(t)
	client := redisstore.New(redisstore.Config{Address: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	hash, err := HashPassword("correct", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	service := New(Config{MaxPasswordAttempts: 2, AttemptWindow: 5 * time.Minute, LockoutSeconds: 30, PasswordCost: bcrypt.MinCost}, client, nil, nil, Nodes{})
	room := roomRecord(DoorMode(roommodel.DoorModePassword))
	room.PasswordHash = &hash
	request := Request{Room: room, PlayerID: 8, Password: "wrong"}
	if _, err := service.Authorize(context.Background(), request); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("expected first wrong password, got %v", err)
	}
	if _, err := service.Authorize(context.Background(), request); !errors.Is(err, ErrEntryLocked) {
		t.Fatalf("expected lockout, got %v", err)
	}
	server.FastForward(31 * time.Second)
	if _, err := service.Authorize(context.Background(), request); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("expected fresh first attempt, got %v", err)
	}
}

// roomRecord creates one durable room fixture.
func roomRecord(mode DoorMode) roommodel.Room {
	return roommodel.Room{Base: model.Base{Identity: model.Identity{ID: 9}}, OwnerPlayerID: 7, DoorMode: roommodel.DoorMode(mode), MaxUsers: 25}
}

// DoorMode aliases room mode for concise table fixtures.
type DoorMode roommodel.DoorMode
