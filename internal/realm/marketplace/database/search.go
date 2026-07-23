package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
)

// SearchOffers aggregates open unexpired listings by definition and LTD serial.
func (repository *Repository) SearchOffers(ctx context.Context, search marketrecord.Search) ([]marketrecord.SearchOffer, int32, error) {
	order := "raw_price desc,created_at asc"
	switch search.SortType {
	case 2:
		order = "raw_price asc,created_at asc"
	case 3:
		order = "sold_today desc,raw_price asc"
	case 4:
		order = "sold_today asc,raw_price asc"
	case 5:
		order = "offer_count desc,raw_price asc"
	case 6:
		order = "offer_count asc,raw_price asc"
	}
	query := `with grouped as (
		select l.*,round(avg(l.raw_price) over(partition by l.furniture_definition_id,fi.limited_edition_number))::bigint average_raw_price,
			count(*) over(partition by l.furniture_definition_id,fi.limited_edition_number) offer_count,
			coalesce(ds.sold_count,0) sold_today,
			row_number() over(partition by l.furniture_definition_id,fi.limited_edition_number order by l.raw_price,l.created_at,l.id) choice
		from marketplace_listings l
		join furniture_items fi on fi.id=l.furniture_item_id
		left join marketplace_daily_stats ds on ds.furniture_definition_id=l.furniture_definition_id and ds.day=current_date
		where l.state=0 and l.expires_at>now()
			and l.raw_price+(l.raw_price*$3+99)/100 between $1 and $2
			and (cardinality($4::bigint[])=0 or l.furniture_definition_id=any($4))
	), chosen as (
		select * from grouped where choice=1
	)
	select ` + listingColumns + `,average_raw_price,offer_count,count(*) over()
	from chosen order by ` + order + `,furniture_definition_id,id limit $5`
	rows, err := repository.executor(ctx).Query(ctx, query, search.MinimumBuyerPrice, search.MaximumBuyerPrice, search.CommissionPercent, search.DefinitionIDs, search.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	offers := make([]marketrecord.SearchOffer, 0, search.Limit)
	var totalRows int64
	for rows.Next() {
		var listing marketrecord.Listing
		var buyer pgtype.Int8
		var sold, redeemed pgtype.Timestamptz
		var averageRawPrice, offerCount int64
		if err := rows.Scan(&listing.ID, &listing.SellerPlayerID, &buyer, &listing.FurnitureItemID, &listing.FurnitureDefinitionID, &listing.RawPrice, &listing.State, &listing.ExpiresAt, &sold, &redeemed, &listing.CreatedAt, &averageRawPrice, &offerCount, &totalRows); err != nil {
			return nil, 0, err
		}
		if buyer.Valid {
			listing.BuyerPlayerID = &buyer.Int64
		}
		if sold.Valid {
			value := sold.Time
			listing.SoldAt = &value
		}
		if redeemed.Valid {
			value := redeemed.Time
			listing.RedeemedAt = &value
		}
		offers = append(offers, marketrecord.SearchOffer{Listing: listing, AverageRawPrice: averageRawPrice, OfferCount: boundedInt32(offerCount)})
	}
	return offers, boundedInt32(totalRows), rows.Err()
}

// boundedInt32 safely projects PostgreSQL aggregate counts to Nitro integers.
func boundedInt32(value int64) int32 {
	if value > 2147483647 {
		return 2147483647
	}
	if value < 0 {
		return 0
	}
	return int32(value)
}

// ListOwnListings lists one seller's listings.
func (repository *Repository) ListOwnListings(ctx context.Context, sellerID int64, visibleSince time.Time) ([]marketrecord.Listing, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select `+listingColumns+` from marketplace_listings where seller_player_id=$1 and (created_at>=$2 or state=0 or (state=1 and redeemed_at is null)) order by created_at desc limit 250`, sellerID, visibleSince)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListings(rows)
}

// DefinitionStats returns recent daily history and current open count.
func (repository *Repository) DefinitionStats(ctx context.Context, definitionID int64, days int32) ([]marketrecord.DayStat, int32, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select furniture_definition_id,day,average_raw_price,sold_count from marketplace_daily_stats where furniture_definition_id=$1 and day>=current_date-$2::integer order by day desc`, definitionID, days)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	stats := make([]marketrecord.DayStat, 0, days)
	for rows.Next() {
		var stat marketrecord.DayStat
		if err := rows.Scan(&stat.DefinitionID, &stat.Day, &stat.AverageRawPrice, &stat.SoldCount); err != nil {
			return nil, 0, err
		}
		stats = append(stats, stat)
	}
	var count int32
	err = repository.executor(ctx).QueryRow(ctx, `select count(*) from marketplace_listings where furniture_definition_id=$1 and state=0 and expires_at>now()`, definitionID).Scan(&count)
	return stats, count, err
}
