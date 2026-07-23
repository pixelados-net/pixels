package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
)

const listingColumns = `id,seller_player_id,buyer_player_id,furniture_item_id,furniture_definition_id,raw_price,state,expires_at,sold_at,redeemed_at,created_at`

// CreateListing inserts an open listing.
func (repository *Repository) CreateListing(ctx context.Context, listing marketrecord.Listing) (marketrecord.Listing, error) {
	return scanListing(repository.executor(ctx).QueryRow(ctx, `insert into marketplace_listings(seller_player_id,furniture_item_id,furniture_definition_id,raw_price,expires_at) values($1,$2,$3,$4,$5) returning `+listingColumns, listing.SellerPlayerID, listing.FurnitureItemID, listing.FurnitureDefinitionID, listing.RawPrice, listing.ExpiresAt))
}

// FindListingForUpdate locks and reads one listing.
func (repository *Repository) FindListingForUpdate(ctx context.Context, listingID int64) (marketrecord.Listing, bool, error) {
	listing, err := scanListing(repository.executor(ctx).QueryRow(ctx, `select `+listingColumns+` from marketplace_listings where id=$1 for update`, listingID))
	if errors.Is(err, pgx.ErrNoRows) {
		return marketrecord.Listing{}, false, nil
	}
	return listing, err == nil, err
}

// FindCheapestListing returns the cheapest open listing for a definition.
func (repository *Repository) FindCheapestListing(ctx context.Context, definitionID int64) (marketrecord.Listing, bool, error) {
	listing, err := scanListing(repository.executor(ctx).QueryRow(ctx, `select `+listingColumns+` from marketplace_listings where furniture_definition_id=$1 and state=0 and expires_at>now() order by raw_price,created_at,id limit 1`, definitionID))
	if errors.Is(err, pgx.ErrNoRows) {
		return marketrecord.Listing{}, false, nil
	}
	return listing, err == nil, err
}

// MarkSold conditionally completes one open listing.
func (repository *Repository) MarkSold(ctx context.Context, listingID int64, buyerID int64) (bool, error) {
	result, err := repository.executor(ctx).Exec(ctx, `with sold as (update marketplace_listings set state=1,buyer_player_id=$2,sold_at=now(),updated_at=now(),version=version+1 where id=$1 and state=0 and expires_at>now() returning furniture_definition_id,raw_price) insert into marketplace_daily_stats(furniture_definition_id,day,average_raw_price,sold_count) select furniture_definition_id,current_date,raw_price,1 from sold on conflict(furniture_definition_id,day) do update set average_raw_price=((marketplace_daily_stats.average_raw_price*marketplace_daily_stats.sold_count)+excluded.average_raw_price)/(marketplace_daily_stats.sold_count+1),sold_count=marketplace_daily_stats.sold_count+1`, listingID, buyerID)
	return err == nil && result.RowsAffected() == 1, err
}

// CloseListing conditionally closes one open listing.
func (repository *Repository) CloseListing(ctx context.Context, listingID int64, sellerID int64, force bool) (marketrecord.Listing, bool, error) {
	query := `update marketplace_listings set state=2,closed_at=now(),updated_at=now(),version=version+1 where id=$1 and state=0 and seller_player_id=$2 returning ` + listingColumns
	arguments := []any{listingID, sellerID}
	if force {
		query = `update marketplace_listings set state=2,closed_at=now(),updated_at=now(),version=version+1 where id=$1 and state=0 returning ` + listingColumns
		arguments = []any{listingID}
	}
	listing, err := scanListing(repository.executor(ctx).QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return marketrecord.Listing{}, false, nil
	}
	return listing, err == nil, err
}

// RedeemSold marks all unredeemed sold listings and returns their raw total.
func (repository *Repository) RedeemSold(ctx context.Context, sellerID int64) (int64, int32, error) {
	var total int64
	var count int32
	err := repository.executor(ctx).QueryRow(ctx, `with redeemed as (update marketplace_listings set redeemed_at=now(),updated_at=now(),version=version+1 where seller_player_id=$1 and state=1 and redeemed_at is null returning raw_price) select coalesce(sum(raw_price),0),count(*) from redeemed`, sellerID).Scan(&total, &count)
	return total, count, err
}

// ExpireListings closes expired listings and returns them.
func (repository *Repository) ExpireListings(ctx context.Context, limit int32) ([]marketrecord.Listing, error) {
	rows, err := repository.executor(ctx).Query(ctx, `with expired as (select id from marketplace_listings where state=0 and expires_at<=now() order by expires_at for update skip locked limit $1) update marketplace_listings m set state=2,closed_at=now(),updated_at=now(),version=version+1 from expired where m.id=expired.id returning `+listingColumns, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListings(rows)
}

// scanListing scans one listing row.
func scanListing(row pgx.Row) (marketrecord.Listing, error) {
	var listing marketrecord.Listing
	var buyer pgtype.Int8
	var sold, redeemed pgtype.Timestamptz
	err := row.Scan(&listing.ID, &listing.SellerPlayerID, &buyer, &listing.FurnitureItemID, &listing.FurnitureDefinitionID, &listing.RawPrice, &listing.State, &listing.ExpiresAt, &sold, &redeemed, &listing.CreatedAt)
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
	return listing, err
}

// scanListings scans listing rows.
func scanListings(rows pgx.Rows) ([]marketrecord.Listing, error) {
	listings := make([]marketrecord.Listing, 0)
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	return listings, rows.Err()
}
