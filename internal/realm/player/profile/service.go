package profile

import (
	"context"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerfigure "github.com/niflaot/pixels/internal/realm/player/figure"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
	"github.com/niflaot/pixels/pkg/redis"
)

// Service coordinates cold-path public profile mutations.
type Service struct {
	// store persists tags and respect grants.
	store Store
	// players persists figure and motto through the player aggregate.
	players playerservice.AdminManager
	// now supplies the hotel clock for deterministic daily policy.
	now func() time.Time
	// permissions resolves explicit quota bypasses.
	permissions permissionservice.Checker
	// throttles stores cross-session abuse windows.
	throttles *redis.Client
	// config contains bounded public-profile policy.
	config Config
	// location defines the explicit hotel civil day.
	location *time.Location
	// figures stores immutable figure-data rules.
	figures *playerfigure.Catalog
	// unlocks reads player-owned clothing sets.
	unlocks ClothingFinder
	// live reads current club entitlement without a database query.
	live *playerlive.Registry
}

// ClothingFinder reads one player's complete clothing unlock snapshot.
type ClothingFinder interface {
	// Clothing returns all persistent clothing unlocks.
	Clothing(context.Context, int64) (playerwardrobe.ClothingSnapshot, error)
}

// New creates public profile behavior.
func New(store Store, players playerservice.AdminManager) *Service {
	return newService(store, players, nil, nil, DefaultConfig())
}

// NewConfigured creates public profile behavior with explicit dependencies and policy.
func NewConfigured(store Store, players playerservice.AdminManager, permissions permissionservice.Checker, throttles *redis.Client, figures *playerfigure.Catalog, unlocks ClothingFinder, live *playerlive.Registry, config Config) *Service {
	service := newService(store, players, permissions, throttles, config)
	service.figures, service.unlocks, service.live = figures, unlocks, live
	return service
}

// newService normalizes public-profile policy once during construction.
func newService(store Store, players playerservice.AdminManager, permissions permissionservice.Checker, throttles *redis.Client, config Config) *Service {
	if config.MottoMaximumRunes <= 0 || config.TagMaximumCount <= 0 || config.TagMaximumRunes <= 0 || config.DailyRespectLimit <= 0 || config.DailyPetRespectLimit <= 0 || config.RespectThrottle <= 0 {
		config = DefaultConfig()
	}
	location, err := time.LoadLocation(config.HotelTimezone)
	if err != nil {
		location = time.UTC
	}
	return &Service{store: store, players: players, now: time.Now, permissions: permissions, throttles: throttles, config: config, location: location}
}

// UpdateFigure validates and persists an avatar figure replacement.
func (service *Service) UpdateFigure(ctx context.Context, playerID int64, gender string, figure string) (playerservice.Record, error) {
	value := playermodel.Gender(strings.ToUpper(strings.TrimSpace(gender)))
	figure = strings.TrimSpace(figure)
	if !value.Valid() || !playerfigure.Valid(figure) {
		return playerservice.Record{}, ErrInvalidFigure
	}
	if service.figures != nil {
		allowed, err := service.figureAllowed(ctx, playerID, figure, value)
		if err != nil {
			return playerservice.Record{}, err
		}
		if !allowed {
			return playerservice.Record{}, ErrInvalidFigure
		}
	}
	return service.players.Update(ctx, playerID, playerservice.UpdateParams{Look: &figure, Gender: &value})
}

// UpdateMotto validates and persists one public motto.
func (service *Service) UpdateMotto(ctx context.Context, playerID int64, motto string) (playerservice.Record, error) {
	motto = strings.TrimSpace(motto)
	if utf8.RuneCountInString(motto) > service.config.MottoMaximumRunes {
		return playerservice.Record{}, ErrInvalidMotto
	}
	return service.players.Update(ctx, playerID, playerservice.UpdateParams{Motto: &motto})
}

// Tags returns ordered public tags.
func (service *Service) Tags(ctx context.Context, playerID int64) ([]string, error) {
	return service.store.Tags(ctx, playerID)
}

// ReplaceTags normalizes and atomically replaces public tags.
func (service *Service) ReplaceTags(ctx context.Context, playerID int64, tags []string) error {
	values, valid := service.normalizeTags(tags)
	if !valid {
		return ErrInvalidTags
	}
	return service.store.ReplaceTags(ctx, playerID, values)
}

// RespectState returns current respect counters using the hotel-local date.
func (service *Service) RespectState(ctx context.Context, playerID int64) (RespectState, error) {
	state, err := service.store.RespectState(ctx, playerID, service.hotelNow(), service.config.DailyRespectLimit, service.config.DailyPetRespectLimit)
	if err != nil {
		return RespectState{}, err
	}
	if service.unlimited(ctx, playerID) {
		state.UserRemaining = int32(service.config.DailyRespectLimit)
	}
	return state, nil
}

// GrantRespect applies one serialized room-validated respect grant.
func (service *Service) GrantRespect(ctx context.Context, actorID int64, targetID int64) (RespectResult, error) {
	if actorID <= 0 || targetID <= 0 || actorID == targetID {
		return RespectResult{}, ErrRespectNotAllowed
	}
	if service.throttles != nil {
		allowed, err := service.throttles.SetIfAbsent(ctx, "player:respect:throttle:"+strconv.FormatInt(actorID, 10), []byte{1}, service.config.RespectThrottle)
		if err != nil {
			return RespectResult{}, err
		}
		if !allowed {
			return RespectResult{}, ErrRespectThrottled
		}
	}
	unlimited := service.unlimited(ctx, actorID)
	result, err := service.store.GrantRespect(ctx, actorID, targetID, service.hotelNow(), service.config.DailyRespectLimit, unlimited)
	if err == nil && result.Duplicate {
		return result, ErrRespectAlreadyGranted
	}
	if err == nil && !result.Applied {
		return result, ErrRespectExhausted
	}
	return result, err
}

// normalizeTags returns distinct normalized tags in stable order.
func (service *Service) normalizeTags(tags []string) ([]string, bool) {
	if len(tags) > service.config.TagMaximumCount {
		return nil, false
	}
	values := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, raw := range tags {
		value := strings.TrimSpace(raw)
		key := strings.ToLower(value)
		if value == "" || utf8.RuneCountInString(value) > service.config.TagMaximumRunes {
			return nil, false
		}
		if _, found := seen[key]; found {
			return nil, false
		}
		seen[key] = struct{}{}
		values = append(values, value)
	}
	return values, true
}

// hotelNow returns the configured civil-day clock.
func (service *Service) hotelNow() time.Time { return service.now().In(service.location) }

// unlimited reports whether a player bypasses the ordinary daily quota.
func (service *Service) unlimited(ctx context.Context, playerID int64) bool {
	if service.permissions == nil {
		return false
	}
	allowed, err := service.permissions.HasPermission(ctx, playerID, RespectUnlimited)
	return err == nil && allowed
}

// figureAllowed resolves entitlement snapshots before immutable catalog validation.
func (service *Service) figureAllowed(ctx context.Context, playerID int64, figure string, gender playermodel.Gender) (bool, error) {
	club := playermodel.ClubLevelNone
	if service.live != nil {
		if player, found := service.live.Find(playerID); found {
			club = player.Snapshot().ClubLevelAt(service.now())
		}
	}
	unlocked := playerwardrobe.ClothingSnapshot{}
	if service.unlocks != nil {
		var err error
		unlocked, err = service.unlocks.Clothing(ctx, playerID)
		if err != nil {
			return false, err
		}
	}
	return service.figures.Allowed(figure, gender, club, unlocked.FigureSetIDs), nil
}
