package record

import (
	"context"
	"time"
)

// Store persists subscription records and transactions.
type Store interface {
	// WithinTransaction runs work atomically.
	WithinTransaction(ctx context.Context, work func(context.Context) error) error
	// FindMembership finds one membership and locks it when requested.
	FindMembership(ctx context.Context, playerID int64, lock bool) (Membership, bool, error)
	// UpsertMembership writes one membership.
	UpsertMembership(ctx context.Context, membership Membership) error
	// ListDueMemberships lists memberships crossing one durable lifecycle boundary.
	ListDueMemberships(ctx context.Context, now time.Time, paydayInterval time.Duration, giftPeriodSeconds int64) ([]Membership, error)
	// ListOffers lists enabled club offers.
	ListOffers(ctx context.Context, deals bool) ([]Offer, error)
	// FindOffer finds one enabled club offer.
	FindOffer(ctx context.Context, id int64) (Offer, bool, error)
	// InsertPayday writes one kickback reward.
	InsertPayday(ctx context.Context, payday Payday) (Payday, error)
	// ListUnclaimedPaydays lists pending rewards.
	ListUnclaimedPaydays(ctx context.Context, playerID int64) ([]Payday, error)
	// MarkPaydayClaimed marks one reward delivered.
	MarkPaydayClaimed(ctx context.Context, id int64) error
	// InsertGiftClaim records one monthly gift claim.
	InsertGiftClaim(ctx context.Context, playerID int64, period time.Time, itemID int64) error
	// FindTargetedOffer finds an eligible offer after an optional id.
	FindTargetedOffer(ctx context.Context, playerID int64, afterID int64) (TargetedOffer, bool, error)
	// FindTargetedOfferByID finds one eligible offer by id.
	FindTargetedOfferByID(ctx context.Context, playerID int64, offerID int64) (TargetedOffer, bool, error)
	// UpdateTargetedState records viewed or dismissed state.
	UpdateTargetedState(ctx context.Context, playerID int64, offerID int64, dismissed bool) error
	// IncrementTargetedPurchase increments a purchase count under its limit.
	IncrementTargetedPurchase(ctx context.Context, playerID int64, offerID int64, quantity int32) (bool, error)
	// FindCampaign finds an enabled campaign by name.
	FindCampaign(ctx context.Context, name string) (Campaign, bool, error)
	// FindActiveCampaign finds the current enabled campaign.
	FindActiveCampaign(ctx context.Context, now time.Time) (Campaign, bool, error)
	// FindCampaignDay finds one campaign reward.
	FindCampaignDay(ctx context.Context, campaignID int64, day int32) (CampaignDay, bool, error)
	// ListCampaignDays lists all campaign rewards.
	ListCampaignDays(ctx context.Context, campaignID int64) ([]CampaignDay, error)
	// ListOpenedDays lists claimed door numbers.
	ListOpenedDays(ctx context.Context, campaignID int64, playerID int64) ([]int32, error)
	// InsertDoorClaim records one claimed door.
	InsertDoorClaim(ctx context.Context, campaignID int64, playerID int64, day int32) error
	// FindSeasonalOffer finds today's linked catalog offer.
	FindSeasonalOffer(ctx context.Context, date time.Time) (SeasonalOffer, bool, error)
}
