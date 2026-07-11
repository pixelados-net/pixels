package entry

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/i18n"
	i18nduration "github.com/niflaot/pixels/pkg/i18n/duration"
	"github.com/niflaot/pixels/pkg/redis"
)

// Nodes stores global room entry permission nodes.
type Nodes struct {
	// EnterAny bypasses room access modes and room bans.
	EnterAny permission.Node
	// EnterFull bypasses normal room capacity.
	EnterFull permission.Node
	// AnswerAnyDoorbell allows answering requests without room-scoped rights.
	AnswerAnyDoorbell permission.Node
}

// RightsChecker resolves future room-scoped rights.
type RightsChecker interface {
	// HasRights reports whether a player has room-scoped rights.
	HasRights(ctx context.Context, roomID int64, playerID int64) (bool, error)
}

// BanChecker resolves future room-scoped bans.
type BanChecker interface {
	// IsBanned reports whether a player is banned from a room.
	IsBanned(ctx context.Context, roomID int64, playerID int64) (bool, error)
}

// Request contains one room entry authorization request.
type Request struct {
	// Room stores durable target room settings.
	Room roommodel.Room
	// PlayerID identifies the entering player.
	PlayerID int64
	// Password stores optional plaintext received from the protocol.
	Password string
	// Trusted marks a server-controlled direct entry.
	Trusted bool
}

// Result contains authorization side effects required by the caller.
type Result struct {
	// Alert stores a localized one-time lockout message.
	Alert string
}

// Service decides entry access and manages temporary protection state.
type Service struct {
	// config stores normalized behavior settings.
	config Config
	// redis stores password-attempt counters and lockouts.
	redis *redis.Client
	// permissions resolves global permission nodes.
	permissions permissionservice.Checker
	// lockoutMessage stores the localized immutable lockout alert.
	lockoutMessage string
	// nodes stores room entry nodes.
	nodes Nodes
	// trust stores short-lived server entry bypasses.
	trust TrustStore
	// rights resolves future room-scoped rights.
	rights RightsChecker
	// bans resolves future room-scoped bans.
	bans BanChecker
	// now returns current time for deterministic tests.
	now func() time.Time
}

// New creates a room entry service.
func New(config Config, redisClient *redis.Client, permissions permissionservice.Checker, translations i18n.Translator, nodes Nodes) *Service {
	config = config.Normalize()
	message := ""
	if translations != nil {
		message = translations.Default("room.entry.locked", i18nduration.DefaultParams(translations, config.LockoutDuration()))
	}

	return &Service{config: config, redis: redisClient, permissions: permissions, lockoutMessage: message, nodes: nodes, now: time.Now}
}

// WithRights configures room-scoped rights resolution.
func (service *Service) WithRights(checker RightsChecker) *Service {
	service.rights = checker

	return service
}

// WithBans configures room-scoped ban resolution.
func (service *Service) WithBans(checker BanChecker) *Service {
	service.bans = checker

	return service
}

// Config returns normalized entry settings.
func (service *Service) Config() Config {
	return service.config
}

// GrantTrusted grants one short-lived server-controlled bypass.
func (service *Service) GrantTrusted(playerID int64, roomID int64) bool {
	return service.trust.Grant(playerID, roomID, service.now().Add(service.config.TrustedTTL))
}

// Authorize decides whether an entry may proceed.
func (service *Service) Authorize(ctx context.Context, request Request) (Result, error) {
	if request.PlayerID <= 0 || request.Room.ID <= 0 {
		return Result{}, ErrAccessDenied
	}
	enterAny, err := service.checkBan(ctx, request)
	if err != nil {
		return Result{}, err
	}
	if request.Room.OwnerPlayerID == request.PlayerID || request.Trusted || enterAny {
		return Result{}, nil
	}
	hasRights, err := service.hasRights(ctx, request.Room.ID, request.PlayerID)
	if err != nil {
		return Result{}, err
	}
	if hasRights {
		return Result{}, nil
	}
	if request.Room.DoorMode == roommodel.DoorModeOpen {
		return Result{}, nil
	}
	if service.trust.Consume(request.PlayerID, request.Room.ID, service.now()) {
		return Result{}, nil
	}
	if !enterAny {
		enterAny, err = service.hasPermission(ctx, request.PlayerID, service.nodes.EnterAny)
		if err != nil {
			return Result{}, err
		}
	}
	if enterAny {
		return Result{}, nil
	}

	switch request.Room.DoorMode {
	case roommodel.DoorModePassword:
		return service.authorizePassword(ctx, request)
	case roommodel.DoorModeDoorbell:
		return Result{}, ErrDoorbellRequired
	default:
		return Result{}, ErrAccessDenied
	}
}

// CanEnterFull reports whether a player may bypass room capacity.
func (service *Service) CanEnterFull(ctx context.Context, playerID int64) (bool, error) {
	return service.hasPermission(ctx, playerID, service.nodes.EnterFull)
}

// CanAnswerDoorbell reports whether a player may resolve waiting requests.
func (service *Service) CanAnswerDoorbell(ctx context.Context, roomID int64, ownerPlayerID int64, playerID int64) (bool, error) {
	if playerID > 0 && playerID == ownerPlayerID {
		return true, nil
	}
	hasRights, err := service.hasRights(ctx, roomID, playerID)
	if err != nil || hasRights {
		return hasRights, err
	}

	return service.hasPermission(ctx, playerID, service.nodes.AnswerAnyDoorbell)
}

// checkBan resolves bans and the global bypass only when needed.
func (service *Service) checkBan(ctx context.Context, request Request) (bool, error) {
	if service.bans == nil {
		return false, nil
	}
	banned, err := service.bans.IsBanned(ctx, request.Room.ID, request.PlayerID)
	if err != nil || !banned {
		return false, err
	}
	allowed, err := service.hasPermission(ctx, request.PlayerID, service.nodes.EnterAny)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, ErrBanned
	}

	return true, nil
}

// hasRights resolves optional room-scoped rights.
func (service *Service) hasRights(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	if service.rights == nil {
		return false, nil
	}

	return service.rights.HasRights(ctx, roomID, playerID)
}

// hasPermission resolves an optional global permission checker.
func (service *Service) hasPermission(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil || node == "" {
		return false, nil
	}

	return service.permissions.HasPermission(ctx, playerID, node)
}
