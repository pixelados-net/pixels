package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// FindCampaign finds one enabled campaign by name.
func (repository *Repository) FindCampaign(ctx context.Context, name string) (record.Campaign, bool, error) {
	var campaign record.Campaign
	var start pgtype.Date
	err := repository.executorFor(ctx).QueryRow(ctx, `select id,name,image,start_date,day_count,enabled from calendar_campaigns where name=$1 and enabled`, name).Scan(&campaign.ID, &campaign.Name, &campaign.Image, &start, &campaign.DayCount, &campaign.Enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.Campaign{}, false, nil
	}
	if start.Valid {
		campaign.StartDate = start.Time
	}
	return campaign, err == nil, err
}

// FindActiveCampaign finds the current enabled campaign.
func (repository *Repository) FindActiveCampaign(ctx context.Context, now time.Time) (record.Campaign, bool, error) {
	var campaign record.Campaign
	var start pgtype.Date
	err := repository.executorFor(ctx).QueryRow(ctx, `select id,name,image,start_date,day_count,enabled from calendar_campaigns where enabled and start_date<=$1::date and start_date+day_count>$1::date order by start_date desc,id limit 1`, now).Scan(&campaign.ID, &campaign.Name, &campaign.Image, &start, &campaign.DayCount, &campaign.Enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.Campaign{}, false, nil
	}
	if start.Valid {
		campaign.StartDate = start.Time
	}
	return campaign, err == nil, err
}

// FindCampaignDay finds one campaign reward.
func (repository *Repository) FindCampaignDay(ctx context.Context, campaignID int64, day int32) (record.CampaignDay, bool, error) {
	var result record.CampaignDay
	var definition pgtype.Int8
	err := repository.executorFor(ctx).QueryRow(ctx, `select campaign_id,day_number,product_definition_id,custom_image,credits_reward,points_reward,points_type from calendar_campaign_days where campaign_id=$1 and day_number=$2`, campaignID, day).Scan(&result.CampaignID, &result.DayNumber, &definition, &result.CustomImage, &result.CreditsReward, &result.PointsReward, &result.PointsType)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.CampaignDay{}, false, nil
	}
	if definition.Valid {
		result.ProductDefinitionID = &definition.Int64
	}
	return result, err == nil, err
}

// ListCampaignDays lists all campaign rewards.
func (repository *Repository) ListCampaignDays(ctx context.Context, campaignID int64) ([]record.CampaignDay, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select campaign_id,day_number,product_definition_id,custom_image,credits_reward,points_reward,points_type from calendar_campaign_days where campaign_id=$1 order by day_number`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.CampaignDay, 0)
	for rows.Next() {
		var day record.CampaignDay
		var definition pgtype.Int8
		if err := rows.Scan(&day.CampaignID, &day.DayNumber, &definition, &day.CustomImage, &day.CreditsReward, &day.PointsReward, &day.PointsType); err != nil {
			return nil, err
		}
		if definition.Valid {
			day.ProductDefinitionID = &definition.Int64
		}
		result = append(result, day)
	}
	return result, rows.Err()
}

// ListOpenedDays lists claimed door numbers.
func (repository *Repository) ListOpenedDays(ctx context.Context, campaignID int64, playerID int64) ([]int32, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select day_number from calendar_door_claims where campaign_id=$1 and player_id=$2 order by day_number`, campaignID, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]int32, 0)
	for rows.Next() {
		var day int32
		if err := rows.Scan(&day); err != nil {
			return nil, err
		}
		result = append(result, day)
	}
	return result, rows.Err()
}

// InsertDoorClaim records one claimed door.
func (repository *Repository) InsertDoorClaim(ctx context.Context, campaignID int64, playerID int64, day int32) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into calendar_door_claims (campaign_id,player_id,day_number) values ($1,$2,$3)`, campaignID, playerID, day)
	return err
}

// FindSeasonalOffer finds the linked offer for one date.
func (repository *Repository) FindSeasonalOffer(ctx context.Context, date time.Time) (record.SeasonalOffer, bool, error) {
	var result record.SeasonalOffer
	var foundDate pgtype.Date
	err := repository.executorFor(ctx).QueryRow(ctx, `select offer_date,catalog_page_id,catalog_item_id from calendar_seasonal_offers where offer_date=$1::date`, date).Scan(&foundDate, &result.CatalogPageID, &result.CatalogItemID)
	if errors.Is(err, pgx.ErrNoRows) {
		return record.SeasonalOffer{}, false, nil
	}
	if foundDate.Valid {
		result.Date = foundDate.Time
	}
	return result, err == nil, err
}
