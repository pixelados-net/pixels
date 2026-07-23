// Package core coordinates subscription behavior.
package core

import (
	"context"
	"errors"
	"time"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

var (
	// ErrMembershipNotFound reports a missing membership.
	ErrMembershipNotFound = errors.New("subscription membership not found")
	// ErrOfferNotFound reports an unknown subscription offer.
	ErrOfferNotFound = errors.New("subscription offer not found")
	// ErrInvalidAmount reports an unsupported subscription quantity.
	ErrInvalidAmount = errors.New("invalid subscription amount")
	// ErrTargetedOfferUnavailable reports an exhausted personalized offer.
	ErrTargetedOfferUnavailable = errors.New("targeted offer unavailable")
	// ErrCalendarDoorUnavailable reports an invalid or claimed calendar door.
	ErrCalendarDoorUnavailable = errors.New("calendar door unavailable")
)

// Options controls subscription calculations.
type Options struct {
	// PaydayInterval stores the kickback cycle.
	PaydayInterval time.Duration
	// KickbackPercentage stores spending returned as credits.
	KickbackPercentage float64
	// PaydayCurrencyType identifies the reward currency.
	PaydayCurrencyType int32
	// BonusRareCurrencyType identifies the currency counted by Bonus Rare.
	BonusRareCurrencyType int32
	// BonusRareThreshold stores the required currency balance.
	BonusRareThreshold int64
	// BonusRareProductID identifies the displayed furniture reward.
	BonusRareProductID int32
}

// CurrencyManager supplies the mutations and reads used by subscription features.
type CurrencyManager interface {
	currencyservice.Granter
	// Balance returns one configured currency balance.
	Balance(ctx context.Context, playerID int64, currencyType int32) (int64, error)
}

// Catalog manages catalog-backed rewards and spending history.
type Catalog interface {
	catalogservice.Manager
	catalogservice.SpendingReader
}

// Service coordinates memberships, offers, and rewards.
type Service struct {
	// options stores immutable calculation settings.
	options Options
	// store persists subscription records.
	store record.Store
	// players writes derived club entitlement.
	players playerservice.ClubWriter
	// livePlayers projects committed membership state into online players.
	livePlayers *playerlive.Registry
	// currencies grants and charges balances.
	currencies CurrencyManager
	// furniture grants calendar furniture.
	furniture furnitureservice.DefinitionGranter
	// catalog buys catalog-backed rewards.
	catalog Catalog
	// events publishes committed lifecycle facts.
	events bus.Publisher
	// now supplies current time.
	now func() time.Time
}

// New creates subscription behavior.
func New(options Options, store record.Store, players playerservice.ClubWriter, currencies CurrencyManager, furniture furnitureservice.DefinitionGranter, catalog Catalog, events bus.Publisher, live ...*playerlive.Registry) *Service {
	service := &Service{options: options, store: store, players: players, currencies: currencies, furniture: furniture, catalog: catalog, events: events, now: time.Now}
	if len(live) != 0 {
		service.livePlayers = live[0]
	}
	return service
}
