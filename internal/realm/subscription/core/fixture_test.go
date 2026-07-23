package core

import (
	"context"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// coreFixture contains focused subscription collaborators.
type coreFixture struct {
	// service stores tested subscription behavior.
	service *Service
	// store stores durable fixtures.
	store *fakeStore
	// currencies records charges.
	currencies *fakeCurrencies
	// players records club projections.
	players *fakeClubWriter
	// catalog supplies reward and spending behavior.
	catalog *fakeCatalog
}

// newCoreFixture creates subscription behavior with deterministic time.
func newCoreFixture(now time.Time) coreFixture {
	store := &fakeStore{offers: []record.Offer{{ID: 1, Name: "hc_31_days", DayCount: 31, PriceCredits: 25, Enabled: true}}}
	currencies := &fakeCurrencies{}
	players := &fakeClubWriter{}
	catalog := &fakeCatalog{}
	service := New(Options{PaydayInterval: 31 * 24 * time.Hour, KickbackPercentage: 0.1, PaydayCurrencyType: -1},
		store, players, currencies, fakeFurniture{}, catalog, nil)
	service.now = func() time.Time { return now }

	return coreFixture{service: service, store: store, currencies: currencies, players: players, catalog: catalog}
}

// fakeStore supplies focused subscription persistence.
type fakeStore struct {
	record.Store
	// membership stores one membership.
	membership record.Membership
	// found reports whether membership exists.
	found bool
	// offers stores club offers.
	offers []record.Offer
	// targeted stores one targeted offer.
	targeted record.TargetedOffer
	// campaign stores one calendar campaign.
	campaign record.Campaign
	// campaignDay stores one calendar reward.
	campaignDay record.CampaignDay
	// opened stores claimed calendar days.
	opened []int32
	// paydays stores unclaimed payday records.
	paydays []record.Payday
	// active stores scheduler membership fixtures.
	active []record.Membership
	// targetedPurchases counts targeted purchases.
	targetedPurchases int32
	// doorClaimed reports whether the calendar door was claimed.
	doorClaimed bool
	// giftClaimed reports whether a monthly gift period was consumed.
	giftClaimed bool
}

// WithinTransaction runs work synchronously.
func (store *fakeStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// FindMembership returns one membership.
func (store *fakeStore) FindMembership(context.Context, int64, bool) (record.Membership, bool, error) {
	return store.membership, store.found, nil
}

// UpsertMembership stores one membership.
func (store *fakeStore) UpsertMembership(_ context.Context, membership record.Membership) error {
	store.membership, store.found = membership, true
	return nil
}

// ListOffers lists offers by deal state.
func (store *fakeStore) ListOffers(_ context.Context, deals bool) ([]record.Offer, error) {
	result := make([]record.Offer, 0, len(store.offers))
	for _, offer := range store.offers {
		if offer.Deal == deals {
			result = append(result, offer)
		}
	}
	return result, nil
}

// FindOffer finds one offer.
func (store *fakeStore) FindOffer(_ context.Context, id int64) (record.Offer, bool, error) {
	for _, offer := range store.offers {
		if offer.ID == id {
			return offer, true, nil
		}
	}
	return record.Offer{}, false, nil
}

// FindTargetedOffer returns one eligible targeted offer.
func (store *fakeStore) FindTargetedOffer(_ context.Context, _ int64, afterID int64) (record.TargetedOffer, bool, error) {
	if store.targeted.ID == 0 || store.targeted.ID == afterID || store.targeted.Dismissed {
		return record.TargetedOffer{}, false, nil
	}
	return store.targeted, true, nil
}

// FindTargetedOfferByID returns the configured eligible offer by id.
func (store *fakeStore) FindTargetedOfferByID(_ context.Context, _ int64, offerID int64) (record.TargetedOffer, bool, error) {
	if store.targeted.ID != offerID || store.targeted.Dismissed || store.targetedPurchases >= store.targeted.PurchaseLimit {
		return record.TargetedOffer{}, false, nil
	}
	offer := store.targeted
	offer.PurchasesCount = store.targetedPurchases
	return offer, true, nil
}

// IncrementTargetedPurchase records one targeted purchase.
func (store *fakeStore) IncrementTargetedPurchase(_ context.Context, _ int64, _ int64, quantity int32) (bool, error) {
	if quantity <= 0 || store.targetedPurchases+quantity > store.targeted.PurchaseLimit {
		return false, nil
	}
	store.targetedPurchases += quantity
	return true, nil
}

// UpdateTargetedState records targeted dismissal state.
func (store *fakeStore) UpdateTargetedState(_ context.Context, _ int64, _ int64, dismissed bool) error {
	store.targeted.Dismissed = dismissed
	return nil
}

// FindCampaign returns one campaign by name.
func (store *fakeStore) FindCampaign(_ context.Context, name string) (record.Campaign, bool, error) {
	return store.campaign, store.campaign.ID != 0 && store.campaign.Name == name, nil
}

// FindActiveCampaign returns the configured campaign.
func (store *fakeStore) FindActiveCampaign(context.Context, time.Time) (record.Campaign, bool, error) {
	return store.campaign, store.campaign.ID != 0 && store.campaign.Enabled, nil
}

// FindCampaignDay returns one configured campaign day.
func (store *fakeStore) FindCampaignDay(_ context.Context, campaignID int64, dayNumber int32) (record.CampaignDay, bool, error) {
	return store.campaignDay, store.campaignDay.CampaignID == campaignID && store.campaignDay.DayNumber == dayNumber, nil
}

// InsertDoorClaim records one calendar claim exactly once.
func (store *fakeStore) InsertDoorClaim(context.Context, int64, int64, int32) error {
	if store.doorClaimed {
		return ErrCalendarDoorUnavailable
	}
	store.doorClaimed = true
	store.opened = append(store.opened, store.campaignDay.DayNumber)
	return nil
}

// ListCampaignDays lists the configured calendar reward.
func (store *fakeStore) ListCampaignDays(context.Context, int64) ([]record.CampaignDay, error) {
	return []record.CampaignDay{store.campaignDay}, nil
}

// ListOpenedDays lists claimed calendar doors.
func (store *fakeStore) ListOpenedDays(context.Context, int64, int64) ([]int32, error) {
	return store.opened, nil
}

// ListDueMemberships lists scheduler membership fixtures.
func (store *fakeStore) ListDueMemberships(context.Context, time.Time, time.Duration, int64) ([]record.Membership, error) {
	return store.active, nil
}

// InsertPayday stores one pending payday.
func (store *fakeStore) InsertPayday(_ context.Context, payday record.Payday) (record.Payday, error) {
	payday.ID = int64(len(store.paydays) + 1)
	store.paydays = append(store.paydays, payday)
	return payday, nil
}

// ListUnclaimedPaydays lists pending payday records.
func (store *fakeStore) ListUnclaimedPaydays(context.Context, int64) ([]record.Payday, error) {
	return store.paydays, nil
}

// MarkPaydayClaimed marks one payday claimed.
func (store *fakeStore) MarkPaydayClaimed(_ context.Context, id int64) error {
	for index := range store.paydays {
		if store.paydays[index].ID == id {
			store.paydays[index].Claimed = true
		}
	}
	return nil
}

// InsertGiftClaim records one monthly gift exactly once.
func (store *fakeStore) InsertGiftClaim(context.Context, int64, time.Time, int64) error {
	if store.giftClaimed {
		return ErrOfferNotFound
	}
	store.giftClaimed = true
	return nil
}

// fakeClubWriter records player club projection.
type fakeClubWriter struct {
	// club stores the latest projection.
	club playermodel.Club
}

// SetClub records one club projection.
func (writer *fakeClubWriter) SetClub(_ context.Context, _ int64, club playermodel.Club) error {
	writer.club = club
	return nil
}

// fakeCurrencies records signed balance mutations.
type fakeCurrencies struct {
	// amounts stores signed mutations.
	amounts []int64
	// balance stores the configured test balance.
	balance int64
	// balanceErr stores an optional read failure.
	balanceErr error
}

// Grant records one currency mutation.
func (currencies *fakeCurrencies) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currencies.amounts = append(currencies.amounts, params.Amount)
	return 100 + params.Amount, nil
}

// Balance returns the configured test balance.
func (currencies *fakeCurrencies) Balance(context.Context, int64, int32) (int64, error) {
	return currencies.balance, currencies.balanceErr
}
