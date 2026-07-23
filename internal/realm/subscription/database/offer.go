package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// ListOffers lists enabled club offers.
func (repository *Repository) ListOffers(ctx context.Context, deals bool) ([]record.Offer, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,name,day_count,price_credits,price_points,points_type,is_vip,is_deal,enabled,order_num from subscription_club_offers where enabled and is_deal=$1 order by order_num,id`, deals)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.Offer, 0)
	for rows.Next() {
		var offer record.Offer
		if err := rows.Scan(&offer.ID, &offer.Name, &offer.DayCount, &offer.PriceCredits, &offer.PricePoints, &offer.PointsType, &offer.VIP, &offer.Deal, &offer.Enabled, &offer.OrderNum); err != nil {
			return nil, err
		}
		result = append(result, offer)
	}
	return result, rows.Err()
}

// FindOffer finds one enabled club offer.
func (repository *Repository) FindOffer(ctx context.Context, id int64) (record.Offer, bool, error) {
	var offer record.Offer
	err := repository.executorFor(ctx).QueryRow(ctx, `select id,name,day_count,price_credits,price_points,points_type,is_vip,is_deal,enabled,order_num from subscription_club_offers where id=$1 and enabled`, id).Scan(&offer.ID, &offer.Name, &offer.DayCount, &offer.PriceCredits, &offer.PricePoints, &offer.PointsType, &offer.VIP, &offer.Deal, &offer.Enabled, &offer.OrderNum)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.Offer{}, false, nil
	}
	return offer, err == nil, err
}

// FindTargetedOffer finds one eligible personalized offer.
func (repository *Repository) FindTargetedOffer(ctx context.Context, playerID int64, afterID int64) (record.TargetedOffer, bool, error) {
	var offer record.TargetedOffer
	var expires pgtype.Timestamptz
	err := repository.executorFor(ctx).QueryRow(ctx, `select o.id,o.catalog_item_id,o.price_credits,o.price_points,o.points_type,o.purchase_limit,o.title_key,o.description_key,o.image_url,o.icon_url,o.expires_at,o.order_num,coalesce(p.purchases_count,0),coalesce(p.dismissed,false) from subscription_targeted_offers o left join subscription_targeted_offer_progress p on p.offer_id=o.id and p.player_id=$1 where o.enabled and o.expires_at>now() and o.image_url<>'' and o.icon_url<>'' and o.id<>$2 and coalesce(p.dismissed,false)=false and coalesce(p.purchases_count,0)<o.purchase_limit order by o.order_num,o.id limit 1`, playerID, afterID).Scan(&offer.ID, &offer.CatalogItemID, &offer.PriceCredits, &offer.PricePoints, &offer.PointsType, &offer.PurchaseLimit, &offer.TitleKey, &offer.DescriptionKey, &offer.ImageURL, &offer.IconURL, &expires, &offer.OrderNum, &offer.PurchasesCount, &offer.Dismissed)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.TargetedOffer{}, false, nil
	}
	if expires.Valid {
		offer.ExpiresAt = &expires.Time
	}
	return offer, err == nil, err
}

// FindTargetedOfferByID finds one eligible personalized offer by id.
func (repository *Repository) FindTargetedOfferByID(ctx context.Context, playerID int64, offerID int64) (record.TargetedOffer, bool, error) {
	var offer record.TargetedOffer
	var expires pgtype.Timestamptz
	err := repository.executorFor(ctx).QueryRow(ctx, `select o.id,o.catalog_item_id,o.price_credits,o.price_points,o.points_type,o.purchase_limit,o.title_key,o.description_key,o.image_url,o.icon_url,o.expires_at,o.order_num,coalesce(p.purchases_count,0),coalesce(p.dismissed,false) from subscription_targeted_offers o left join subscription_targeted_offer_progress p on p.offer_id=o.id and p.player_id=$1 where o.id=$2 and o.enabled and o.expires_at>now() and o.image_url<>'' and o.icon_url<>'' and coalesce(p.dismissed,false)=false and coalesce(p.purchases_count,0)<o.purchase_limit`, playerID, offerID).Scan(&offer.ID, &offer.CatalogItemID, &offer.PriceCredits, &offer.PricePoints, &offer.PointsType, &offer.PurchaseLimit, &offer.TitleKey, &offer.DescriptionKey, &offer.ImageURL, &offer.IconURL, &expires, &offer.OrderNum, &offer.PurchasesCount, &offer.Dismissed)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.TargetedOffer{}, false, nil
	}
	if expires.Valid {
		offer.ExpiresAt = &expires.Time
	}
	return offer, err == nil, err
}

// UpdateTargetedState records viewed or dismissed state.
func (repository *Repository) UpdateTargetedState(ctx context.Context, playerID int64, offerID int64, dismissed bool) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into subscription_targeted_offer_progress (player_id,offer_id,last_viewed_at,dismissed) values ($1,$2,now(),$3) on conflict (player_id,offer_id) do update set last_viewed_at=now(),dismissed=excluded.dismissed`, playerID, offerID, dismissed)
	return err
}

// IncrementTargetedPurchase increments a purchase count under its limit.
func (repository *Repository) IncrementTargetedPurchase(ctx context.Context, playerID int64, offerID int64, quantity int32) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `insert into subscription_targeted_offer_progress (player_id,offer_id,purchases_count) select $1,o.id,$3 from subscription_targeted_offers o where o.id=$2 and o.enabled and o.expires_at>now() and o.image_url<>'' and o.icon_url<>'' and $3>0 and $3<=o.purchase_limit on conflict (player_id,offer_id) do update set purchases_count=subscription_targeted_offer_progress.purchases_count+$3 where $3>0 and subscription_targeted_offer_progress.purchases_count+$3<=(select purchase_limit from subscription_targeted_offers where id=$2 and enabled and expires_at>now())`, playerID, offerID, quantity)
	return err == nil && tag.RowsAffected() == 1, err
}

// targetedNow anchors tests and future selection rules.
var targetedNow = time.Now
