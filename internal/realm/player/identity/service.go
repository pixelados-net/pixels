package identity

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/pkg/redis"
)

// Service coordinates username checks, reservations, and committed renames.
type Service struct {
	// store commits atomic renames.
	store Store
	// players resolves exact active usernames and policy state.
	players playerservice.Finder
	// reservations stores short candidate claims.
	reservations *redis.Client
	// config contains bounded username policy.
	config Config
	// filter applies the immutable hotel word dictionary.
	filter WordFilter
}

// WordFilter applies the immutable hotel dictionary to one candidate.
type WordFilter interface {
	// Censor returns filtered text and whether the dictionary matched.
	Censor(string) (string, bool)
}

// New creates identity behavior.
func New(store Store, players playerservice.Finder, reservations *redis.Client) *Service {
	return NewConfigured(store, players, reservations, nil, DefaultConfig())
}

// NewConfigured creates identity behavior with explicit policy.
func NewConfigured(store Store, players playerservice.Finder, reservations *redis.Client, filter WordFilter, config Config) *Service {
	if config.MinimumLength <= 0 || config.MaximumLength < config.MinimumLength || config.ReservationTTL <= 0 {
		config = DefaultConfig()
	}
	return &Service{store: store, players: players, reservations: reservations, config: config, filter: filter}
}

// Check validates availability and reserves an available candidate for one player.
func (service *Service) Check(ctx context.Context, playerID int64, candidate string) (CheckResult, error) {
	result := CheckResult{Username: strings.TrimSpace(candidate)}
	result.Code = service.validate(result.Username)
	if result.Code != ResultAvailable {
		return result, nil
	}
	if service.reserved(result.Username) || service.filtered(result.Username) {
		result.Code = ResultInvalid
		return result, nil
	}
	record, found, err := service.players.FindByID(ctx, playerID)
	if err != nil {
		return CheckResult{}, err
	}
	if !found || !record.Profile.AllowNameChange {
		result.Code = ResultDisabled
		return result, nil
	}
	_, taken, err := service.players.FindByUsername(ctx, result.Username)
	if err != nil {
		return CheckResult{}, err
	}
	if taken {
		result.Code = ResultTaken
		result.Suggestions, err = service.suggestions(ctx, result.Username)
		return result, err
	}
	if service.reservations != nil {
		reserved, reserveErr := service.reservations.SetIfAbsent(ctx, reservationKey(result.Username), []byte(strconv.FormatInt(playerID, 10)), service.config.ReservationTTL)
		if reserveErr != nil {
			return CheckResult{}, reserveErr
		}
		if !reserved {
			result.Code = ResultTaken
			result.Suggestions, err = service.suggestions(ctx, result.Username)
		}
	}
	return result, err
}

// Rename consumes a matching reservation and commits one one-shot rename.
func (service *Service) Rename(ctx context.Context, playerID int64, candidate string) (RenameResult, error) {
	candidate = strings.TrimSpace(candidate)
	if service.validate(candidate) != ResultAvailable || service.reserved(candidate) || service.filtered(candidate) {
		return RenameResult{}, ErrReservationMissing
	}
	if service.reservations != nil {
		value, found, err := service.reservations.Take(ctx, reservationKey(candidate))
		if err != nil {
			return RenameResult{}, err
		}
		if !found || string(value) != strconv.FormatInt(playerID, 10) {
			return RenameResult{}, ErrReservationMissing
		}
	}
	return service.store.Rename(ctx, playerID, candidate)
}

// suggestions returns at most four verified deterministic alternatives.
func (service *Service) suggestions(ctx context.Context, candidate string) ([]string, error) {
	values := make([]string, 0, 4)
	for suffix := 1; suffix <= 16 && len(values) < 4; suffix++ {
		value := candidate
		addition := strconv.Itoa(suffix)
		for utf8.RuneCountInString(value)+len(addition) > service.config.MaximumLength {
			_, width := utf8.DecodeLastRuneInString(value)
			value = value[:len(value)-width]
		}
		value += addition
		_, found, err := service.players.FindByUsername(ctx, value)
		if err != nil {
			return nil, err
		}
		if !found {
			values = append(values, value)
		}
	}
	return values, nil
}

// validate returns Nitro's stable username result code.
func (service *Service) validate(username string) int32 {
	length := utf8.RuneCountInString(username)
	if length < service.config.MinimumLength {
		return ResultTooShort
	}
	if length > service.config.MaximumLength {
		return ResultTooLong
	}
	for _, value := range username {
		if value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || strings.ContainsRune(service.config.AllowedSymbols, value) {
			continue
		}
		return ResultInvalid
	}
	return ResultAvailable
}

// reserved reports whether a candidate matches a configured protected identity.
func (service *Service) reserved(candidate string) bool {
	for _, value := range service.config.ReservedNames {
		if strings.EqualFold(strings.TrimSpace(value), candidate) {
			return true
		}
	}
	return false
}

// filtered reports whether the immutable hotel dictionary matches a candidate.
func (service *Service) filtered(candidate string) bool {
	if service.filter == nil {
		return false
	}
	_, changed := service.filter.Censor(candidate)
	return changed
}

// reservationKey returns one case-insensitive candidate key.
func reservationKey(candidate string) string {
	return "player:name:reservation:" + strings.ToLower(candidate)
}
