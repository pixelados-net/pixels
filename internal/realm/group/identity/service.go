// Package identity owns social-group creation, metadata, badge, and lifecycle workflows.
package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	"github.com/niflaot/pixels/internal/realm/group/badge"
	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	createdevent "github.com/niflaot/pixels/internal/realm/group/identity/events/created"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/pkg/bus"
)

// CreateParams contains one player group-creation request.
type CreateParams struct {
	// OwnerPlayerID identifies the creator.
	OwnerPlayerID int64
	// Name stores the requested visible title.
	Name string
	// Description stores requested public information.
	Description string
	// HomeRoomID identifies the creator-owned headquarters.
	HomeRoomID int64
	// ColorA identifies the primary editor color.
	ColorA int32
	// ColorB identifies the secondary editor color.
	ColorB int32
	// BadgeParts stores requested badge layers.
	BadgeParts []grouprecord.BadgePart
}

// AdministrativeCreateParams contains trusted administrative creation policy.
type AdministrativeCreateParams struct {
	// CreateParams stores the requested group identity.
	CreateParams CreateParams
	// IdempotencyKey identifies one logical administration request.
	IdempotencyKey string
	// Charge reports whether the configured creation price is applied.
	Charge bool
}

// creationPolicy controls trusted differences between client and administration creation.
type creationPolicy struct {
	// charge reports whether the configured price is applied.
	charge bool
	// requireClub reports whether the owner needs active club entitlement.
	requireClub bool
	// idempotencyKey identifies one replayable operation.
	idempotencyKey string
}

// Service coordinates group identity persistence and cache generations.
type Service struct {
	// config stores creation and limit policy.
	config groupconfig.Config
	// store persists all group records.
	store grouprecord.Store
	// badges validates editor reference data.
	badges *badge.Compiler
	// registry exposes warmed badge choices.
	registry *badge.Registry
	// currencies applies atomic creation charges.
	currencies currencyservice.Granter
	// players reads creator identity and club entitlement.
	players playerservice.Finder
	// permissions resolves hotel staff overrides.
	permissions permissionservice.Checker
	// filter censors configured hotel words.
	filter *chatfilter.Service
	// cache stores immutable hot-path generations.
	cache *groupruntime.Cache
	// projector applies supported current-room group projections after commit.
	projector *groupruntime.Projector
	// metrics stores bounded process-wide group telemetry.
	metrics *groupobservability.Metrics
	// events publishes committed domain changes.
	events bus.Publisher
}

// New creates social-group identity behavior.
func New(config groupconfig.Config, store grouprecord.Store, badges *badge.Compiler, registry *badge.Registry, currencies currencyservice.Granter, players playerservice.Finder, permissions permissionservice.Checker, filter *chatfilter.Service, cache *groupruntime.Cache, projector *groupruntime.Projector, metrics *groupobservability.Metrics, events bus.Publisher) *Service {
	return &Service{config: config, store: store, badges: badges, registry: registry, currencies: currencies, players: players, permissions: permissions, filter: filter, cache: cache, projector: projector, metrics: metrics, events: events}
}

// Create validates and atomically purchases one social group.
func (service *Service) Create(ctx context.Context, params CreateParams) (grouprecord.Group, error) {
	allowed, err := service.has(ctx, params.OwnerPlayerID, grouppolicy.CreateNode)
	if err != nil || !allowed {
		return grouprecord.Group{}, grouprecord.ErrForbidden
	}
	created, _, err := service.create(ctx, params, creationPolicy{charge: true, requireClub: service.config.RequireClub})
	return created, err
}

// CreateAdministrative creates or replays one explicitly charged or free administrative group.
func (service *Service) CreateAdministrative(ctx context.Context, params AdministrativeCreateParams) (grouprecord.Group, bool, error) {
	key := strings.TrimSpace(params.IdempotencyKey)
	if len(key) < 8 || len(key) > 128 {
		return grouprecord.Group{}, false, grouprecord.ErrInvalid
	}
	return service.create(ctx, params.CreateParams, creationPolicy{charge: params.Charge, idempotencyKey: key})
}

// create validates and atomically executes one creation policy.
func (service *Service) create(ctx context.Context, params CreateParams, policy creationPolicy) (created grouprecord.Group, replayed bool, err error) {
	started := time.Now()
	defer func() {
		service.metrics.Record(groupobservability.Operations, groupobservability.KindCreate, identityMetricResult(err))
		service.metrics.Observe(groupobservability.CreateTransaction, time.Since(started))
	}()
	if params.OwnerPlayerID <= 0 || params.HomeRoomID <= 0 {
		return grouprecord.Group{}, false, grouprecord.ErrInvalid
	}
	name, description, err := service.normalizeIdentity(params.Name, params.Description)
	if err != nil {
		return grouprecord.Group{}, false, err
	}
	if _, err = service.BadgeRegistry(ctx); err != nil {
		return grouprecord.Group{}, false, err
	}
	code, parts, err := service.badges.Compile(params.BadgeParts)
	if err != nil {
		return grouprecord.Group{}, false, grouprecord.ErrInvalid
	}
	if err = service.validateColors(params.ColorA, params.ColorB); err != nil {
		return grouprecord.Group{}, false, err
	}
	player, found, err := service.players.FindByID(ctx, params.OwnerPlayerID)
	if err != nil || !found {
		return grouprecord.Group{}, false, grouprecord.ErrNotFound
	}
	if policy.requireClub && !player.Player.Club.ActiveAt(time.Now()) {
		return grouprecord.Group{}, false, grouprecord.ErrForbidden
	}
	requestHash, err := createRequestHash(params, name, description, code, parts, policy.charge)
	if err != nil {
		return grouprecord.Group{}, false, err
	}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if policy.idempotencyKey != "" {
			groupID, replay, claimErr := service.store.ClaimCreateOperation(txCtx, policy.idempotencyKey, requestHash)
			if claimErr != nil {
				return claimErr
			}
			if replay {
				var replayFound bool
				created, replayFound, claimErr = service.store.Group(txCtx, groupID, false)
				if claimErr != nil {
					return claimErr
				}
				if !replayFound {
					return grouprecord.ErrConflict
				}
				replayed = true
				return nil
			}
		}
		if err := service.store.LockEligibleRoom(txCtx, params.HomeRoomID, params.OwnerPlayerID); err != nil {
			return err
		}
		owned, err := service.store.CountOwned(txCtx, params.OwnerPlayerID)
		if err != nil {
			return err
		}
		memberships, err := service.store.CountMemberships(txCtx, params.OwnerPlayerID)
		if err != nil {
			return err
		}
		if owned >= service.config.OwnedLimit || memberships >= service.config.MembershipLimit {
			return grouprecord.ErrLimit
		}
		if policy.charge && service.config.CreationCost > 0 {
			_, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: params.OwnerPlayerID, CurrencyType: -1, Amount: -service.config.CreationCost, Reason: "social_group_create", ActorKind: currencyservice.ActorPlayer})
			if err != nil {
				return err
			}
		}
		created, err = service.store.InsertGroup(txCtx, grouprecord.CreateParams{OwnerPlayerID: params.OwnerPlayerID, Name: name, Description: description, HomeRoomID: params.HomeRoomID, State: grouprecord.Regular, ColorA: params.ColorA, ColorB: params.ColorB, BadgeCode: code, BadgeParts: parts})
		if err != nil || policy.idempotencyKey == "" {
			return err
		}
		return service.store.CompleteCreateOperation(txCtx, policy.idempotencyKey, created.ID)
	})
	if err != nil {
		return grouprecord.Group{}, false, err
	}
	service.refresh(ctx, created.ID, params.OwnerPlayerID)
	if !replayed {
		service.publish(ctx, createdevent.Name, createdevent.Payload{GroupID: created.ID, OwnerPlayerID: params.OwnerPlayerID, Version: created.Version})
	}
	return created, replayed, nil
}

// createRequestHash returns a stable hash for administrative replay comparison.
func createRequestHash(params CreateParams, name string, description string, badgeCode string, parts []grouprecord.BadgePart, charge bool) (string, error) {
	payload, err := json.Marshal(struct {
		OwnerPlayerID int64
		Name          string
		Description   string
		HomeRoomID    int64
		ColorA        int32
		ColorB        int32
		BadgeCode     string
		BadgeParts    []grouprecord.BadgePart
		Charge        bool
	}{params.OwnerPlayerID, name, description, params.HomeRoomID, params.ColorA, params.ColorB, badgeCode, parts, charge})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}
