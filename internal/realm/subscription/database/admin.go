package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// ListPaydays lists player payday history.
func (repository *Repository) ListPaydays(ctx context.Context, playerID int64) ([]record.Payday, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,player_id,occurred_at,streak_days,credits_spent,streak_bonus,monthly_bonus,total_awarded,currency_type,claimed from subscription_payday_log where player_id=$1 order by id desc`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.Payday, 0)
	for rows.Next() {
		var payday record.Payday
		if err := rows.Scan(&payday.ID, &payday.PlayerID, &payday.OccurredAt, &payday.StreakDays, &payday.CreditsSpent, &payday.StreakBonus, &payday.MonthlyBonus, &payday.TotalAwarded, &payday.CurrencyType, &payday.Claimed); err != nil {
			return nil, err
		}
		result = append(result, payday)
	}
	return result, rows.Err()
}

// ListAllOffers lists club offers including disabled records.
func (repository *Repository) ListAllOffers(ctx context.Context) ([]record.Offer, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,name,day_count,price_credits,price_points,points_type,is_vip,is_deal,enabled,order_num from subscription_club_offers order by order_num,id`)
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

// UpsertOffer creates or updates a club offer.
func (repository *Repository) UpsertOffer(ctx context.Context, offer record.Offer) (record.Offer, error) {
	var err error
	if offer.ID == 0 {
		err = repository.executorFor(ctx).QueryRow(ctx, `insert into subscription_club_offers (name,day_count,price_credits,price_points,points_type,is_vip,is_deal,enabled,order_num) values ($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id`, offer.Name, offer.DayCount, offer.PriceCredits, offer.PricePoints, offer.PointsType, offer.VIP, offer.Deal, offer.Enabled, offer.OrderNum).Scan(&offer.ID)
	} else {
		err = repository.executorFor(ctx).QueryRow(ctx, `update subscription_club_offers set name=$2,day_count=$3,price_credits=$4,price_points=$5,points_type=$6,is_vip=$7,is_deal=$8,enabled=$9,order_num=$10 where id=$1 returning id`, offer.ID, offer.Name, offer.DayCount, offer.PriceCredits, offer.PricePoints, offer.PointsType, offer.VIP, offer.Deal, offer.Enabled, offer.OrderNum).Scan(&offer.ID)
	}
	return offer, err
}

// ListTargetedOffers lists targeted offers.
func (repository *Repository) ListTargetedOffers(ctx context.Context) ([]record.TargetedOffer, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,catalog_item_id,price_credits,price_points,points_type,purchase_limit,title_key,description_key,image_url,icon_url,expires_at,order_num,enabled from subscription_targeted_offers order by order_num,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.TargetedOffer, 0)
	for rows.Next() {
		var offer record.TargetedOffer
		var expires pgtype.Timestamptz
		if err := rows.Scan(&offer.ID, &offer.CatalogItemID, &offer.PriceCredits, &offer.PricePoints, &offer.PointsType, &offer.PurchaseLimit, &offer.TitleKey, &offer.DescriptionKey, &offer.ImageURL, &offer.IconURL, &expires, &offer.OrderNum, &offer.Enabled); err != nil {
			return nil, err
		}
		if expires.Valid {
			offer.ExpiresAt = &expires.Time
		}
		result = append(result, offer)
	}
	return result, rows.Err()
}

// UpsertTargetedOffer creates or updates a targeted offer.
func (repository *Repository) UpsertTargetedOffer(ctx context.Context, offer record.TargetedOffer) (record.TargetedOffer, error) {
	var err error
	if offer.ID == 0 {
		err = repository.executorFor(ctx).QueryRow(ctx, `insert into subscription_targeted_offers (catalog_item_id,price_credits,price_points,points_type,purchase_limit,title_key,description_key,image_url,icon_url,expires_at,order_num,enabled) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) returning id`, offer.CatalogItemID, offer.PriceCredits, offer.PricePoints, offer.PointsType, offer.PurchaseLimit, offer.TitleKey, offer.DescriptionKey, offer.ImageURL, offer.IconURL, offer.ExpiresAt, offer.OrderNum, offer.Enabled).Scan(&offer.ID)
	} else {
		err = repository.executorFor(ctx).QueryRow(ctx, `update subscription_targeted_offers set catalog_item_id=$2,price_credits=$3,price_points=$4,points_type=$5,purchase_limit=$6,title_key=$7,description_key=$8,image_url=$9,icon_url=$10,expires_at=$11,order_num=$12,enabled=$13 where id=$1 returning id`, offer.ID, offer.CatalogItemID, offer.PriceCredits, offer.PricePoints, offer.PointsType, offer.PurchaseLimit, offer.TitleKey, offer.DescriptionKey, offer.ImageURL, offer.IconURL, offer.ExpiresAt, offer.OrderNum, offer.Enabled).Scan(&offer.ID)
	}
	return offer, err
}

// ListCampaigns lists calendar campaigns.
func (repository *Repository) ListCampaigns(ctx context.Context) ([]record.Campaign, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,name,image,start_date,day_count,enabled from calendar_campaigns order by start_date desc,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.Campaign, 0)
	for rows.Next() {
		var campaign record.Campaign
		var start pgtype.Date
		if err := rows.Scan(&campaign.ID, &campaign.Name, &campaign.Image, &start, &campaign.DayCount, &campaign.Enabled); err != nil {
			return nil, err
		}
		campaign.StartDate = start.Time
		result = append(result, campaign)
	}
	return result, rows.Err()
}

// UpsertCampaign creates or updates a calendar campaign.
func (repository *Repository) UpsertCampaign(ctx context.Context, campaign record.Campaign) (record.Campaign, error) {
	var err error
	if campaign.ID == 0 {
		err = repository.executorFor(ctx).QueryRow(ctx, `insert into calendar_campaigns (name,image,start_date,day_count,enabled) values ($1,$2,$3,$4,$5) returning id`, campaign.Name, campaign.Image, campaign.StartDate, campaign.DayCount, campaign.Enabled).Scan(&campaign.ID)
	} else {
		err = repository.executorFor(ctx).QueryRow(ctx, `update calendar_campaigns set name=$2,image=$3,start_date=$4,day_count=$5,enabled=$6 where id=$1 returning id`, campaign.ID, campaign.Name, campaign.Image, campaign.StartDate, campaign.DayCount, campaign.Enabled).Scan(&campaign.ID)
	}
	return campaign, err
}

// UpsertCampaignDay creates or updates one campaign day.
func (repository *Repository) UpsertCampaignDay(ctx context.Context, day record.CampaignDay) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into calendar_campaign_days (campaign_id,day_number,product_definition_id,custom_image,credits_reward,points_reward,points_type) values ($1,$2,$3,$4,$5,$6,$7) on conflict (campaign_id,day_number) do update set product_definition_id=excluded.product_definition_id,custom_image=excluded.custom_image,credits_reward=excluded.credits_reward,points_reward=excluded.points_reward,points_type=excluded.points_type`, day.CampaignID, day.DayNumber, day.ProductDefinitionID, day.CustomImage, day.CreditsReward, day.PointsReward, day.PointsType)
	return err
}
