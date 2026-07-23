package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// ListItemProducts lists ordered products for one offer.
func (repository *Repository) ListItemProducts(ctx context.Context, catalogItemID int64) ([]catalogmodel.Product, error) {
	return repository.listProducts(ctx, &catalogItemID)
}

// ListProducts lists every configured bundle product.
func (repository *Repository) ListProducts(ctx context.Context) ([]catalogmodel.Product, error) {
	return repository.listProducts(ctx, nil)
}

// listProducts lists bundle products with an optional offer filter.
func (repository *Repository) listProducts(ctx context.Context, catalogItemID *int64) ([]catalogmodel.Product, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id, catalog_item_id, definition_id, quantity, order_num from catalog_item_products where ($1::bigint is null or catalog_item_id=$1) order by catalog_item_id,order_num,id`, catalogItemID)
	if err != nil {
		return nil, fmt.Errorf("list catalog products: %w", err)
	}
	defer rows.Close()
	products := make([]catalogmodel.Product, 0)
	for rows.Next() {
		var product catalogmodel.Product
		if err := rows.Scan(&product.ID, &product.CatalogItemID, &product.DefinitionID, &product.Quantity, &product.OrderNum); err != nil {
			return nil, fmt.Errorf("scan catalog product: %w", err)
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

// FindVoucherByCode finds one voucher by case-insensitive code.
func (repository *Repository) FindVoucherByCode(ctx context.Context, code string) (catalogmodel.Voucher, bool, error) {
	var voucher catalogmodel.Voucher
	var itemID pgtype.Int8
	var cap pgtype.Int4
	var expires pgtype.Timestamptz
	err := repository.executorFor(ctx).QueryRow(ctx, `select id,code,cost_credits,cost_points,points_type,catalog_item_id,redemption_cap,per_player_cap,enabled,expires_at from catalog_vouchers where upper(code)=upper($1)`, strings.TrimSpace(code)).Scan(&voucher.ID, &voucher.Code, &voucher.CostCredits, &voucher.CostPoints, &voucher.PointsType, &itemID, &cap, &voucher.PerPlayerCap, &voucher.Enabled, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return catalogmodel.Voucher{}, false, nil
	}
	if err != nil {
		return catalogmodel.Voucher{}, false, fmt.Errorf("find catalog voucher: %w", err)
	}
	voucher.CatalogItemID = int64Pointer(itemID)
	if cap.Valid {
		voucher.RedemptionCap = &cap.Int32
	}
	voucher.ExpiresAt = timePointer(expires)
	return voucher, true, nil
}

// CountVoucherRedemptions counts global voucher redemptions.
func (repository *Repository) CountVoucherRedemptions(ctx context.Context, voucherID int64) (int32, error) {
	var count int32
	err := repository.executorFor(ctx).QueryRow(ctx, `select count(*)::integer from catalog_voucher_redemptions where voucher_id=$1`, voucherID).Scan(&count)
	return count, err
}

// InsertVoucherRedemption records one player redemption.
func (repository *Repository) InsertVoucherRedemption(ctx context.Context, voucherID int64, playerID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into catalog_voucher_redemptions (voucher_id,player_id) values ($1,$2)`, voucherID, playerID)
	return err
}

// MarkNewAdditionsSeen records the latest novelty view.
func (repository *Repository) MarkNewAdditionsSeen(ctx context.Context, playerID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into catalog_new_additions_seen (player_id,last_seen_at) values ($1,now()) on conflict (player_id) do update set last_seen_at=excluded.last_seen_at`, playerID)
	return err
}

// NewAdditionsAvailable reports whether unseen novelty offers exist.
func (repository *Repository) NewAdditionsAvailable(ctx context.Context, playerID int64) (bool, error) {
	var available bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from catalog_items i join catalog_pages p on p.id=i.page_id left join catalog_new_additions_seen s on s.player_id=$1 where p.new_additions and i.deleted_at is null and i.enabled and i.created_at>coalesce(s.last_seen_at,'epoch'))`, playerID).Scan(&available)
	return available, err
}

// LogPurchase records a catalog purchase and its granted furniture instances.
func (repository *Repository) LogPurchase(ctx context.Context, playerID int64, item catalogmodel.Item, quantity int32, costCredits int64, costPoints int64, furnitureItemIDs []int64) error {
	var purchaseID int64
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into catalog_purchase_log (player_id,catalog_item_id,quantity,cost_credits,cost_points,points_type) values ($1,$2,$3,$4,$5,$6) returning id`, playerID, item.ID, quantity, costCredits, costPoints, item.PointsType).Scan(&purchaseID)
	if err != nil || len(furnitureItemIDs) == 0 {
		return err
	}
	_, err = repository.executorFor(ctx).Exec(ctx, `insert into catalog_purchase_items (purchase_id,furniture_item_id) select $1,unnest($2::bigint[])`, purchaseID, furnitureItemIDs)
	return err
}

// CreditsSpentSince sums kickback-eligible purchases.
func (repository *Repository) CreditsSpentSince(ctx context.Context, playerID int64, since time.Time) (int64, error) {
	var amount int64
	err := repository.executorFor(ctx).QueryRow(ctx, `select coalesce(sum(l.cost_credits),0)::bigint from catalog_purchase_log l join catalog_items i on i.id=l.catalog_item_id join catalog_pages p on p.id=i.page_id where l.player_id=$1 and l.purchased_at>$2 and not p.excluded_from_kickback`, playerID, since).Scan(&amount)
	return amount, err
}

// CreditsSpentBetween sums kickback-eligible purchases inside one payday period.
func (repository *Repository) CreditsSpentBetween(ctx context.Context, playerID int64, after time.Time, through time.Time) (int64, error) {
	var amount int64
	err := repository.executorFor(ctx).QueryRow(ctx, `select coalesce(sum(l.cost_credits),0)::bigint from catalog_purchase_log l join catalog_items i on i.id=l.catalog_item_id join catalog_pages p on p.id=i.page_id where l.player_id=$1 and l.purchased_at>$2 and l.purchased_at<=$3 and not p.excluded_from_kickback`, playerID, after, through).Scan(&amount)
	return amount, err
}

// ListVouchers lists every voucher.
func (repository *Repository) ListVouchers(ctx context.Context) ([]catalogmodel.Voucher, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,code,cost_credits,cost_points,points_type,catalog_item_id,redemption_cap,per_player_cap,enabled,expires_at from catalog_vouchers order by id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]catalogmodel.Voucher, 0)
	for rows.Next() {
		voucher, scanErr := scanVoucher(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, voucher)
	}
	return result, rows.Err()
}

// UpsertVoucher creates or updates one voucher.
func (repository *Repository) UpsertVoucher(ctx context.Context, voucher catalogmodel.Voucher) (catalogmodel.Voucher, error) {
	if voucher.ID == 0 {
		row := repository.executorFor(ctx).QueryRow(ctx, `insert into catalog_vouchers (code,cost_credits,cost_points,points_type,catalog_item_id,redemption_cap,per_player_cap,enabled,expires_at) values (upper($1),$2,$3,$4,$5,$6,$7,$8,$9) returning id,code,cost_credits,cost_points,points_type,catalog_item_id,redemption_cap,per_player_cap,enabled,expires_at`, voucher.Code, voucher.CostCredits, voucher.CostPoints, voucher.PointsType, voucher.CatalogItemID, voucher.RedemptionCap, voucher.PerPlayerCap, voucher.Enabled, voucher.ExpiresAt)
		return scanVoucher(row)
	}
	row := repository.executorFor(ctx).QueryRow(ctx, `update catalog_vouchers set code=upper($2),cost_credits=$3,cost_points=$4,points_type=$5,catalog_item_id=$6,redemption_cap=$7,per_player_cap=$8,enabled=$9,expires_at=$10 where id=$1 returning id,code,cost_credits,cost_points,points_type,catalog_item_id,redemption_cap,per_player_cap,enabled,expires_at`, voucher.ID, voucher.Code, voucher.CostCredits, voucher.CostPoints, voucher.PointsType, voucher.CatalogItemID, voucher.RedemptionCap, voucher.PerPlayerCap, voucher.Enabled, voucher.ExpiresAt)
	return scanVoucher(row)
}

// ListVoucherRedemptions lists voucher redemption history.
func (repository *Repository) ListVoucherRedemptions(ctx context.Context, voucherID int64) ([]catalogmodel.VoucherRedemption, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select voucher_id,player_id,redeemed_at from catalog_voucher_redemptions where voucher_id=$1 order by redeemed_at desc`, voucherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]catalogmodel.VoucherRedemption, 0)
	for rows.Next() {
		var redemption catalogmodel.VoucherRedemption
		if err := rows.Scan(&redemption.VoucherID, &redemption.PlayerID, &redemption.RedeemedAt); err != nil {
			return nil, err
		}
		result = append(result, redemption)
	}
	return result, rows.Err()
}

// scanVoucher scans one voucher row.
func scanVoucher(row interface{ Scan(...any) error }) (catalogmodel.Voucher, error) {
	var voucher catalogmodel.Voucher
	var item pgtype.Int8
	var cap pgtype.Int4
	var expires pgtype.Timestamptz
	err := row.Scan(&voucher.ID, &voucher.Code, &voucher.CostCredits, &voucher.CostPoints, &voucher.PointsType,
		&item, &cap, &voucher.PerPlayerCap, &voucher.Enabled, &expires)
	if item.Valid {
		voucher.CatalogItemID = &item.Int64
	}
	if cap.Valid {
		voucher.RedemptionCap = &cap.Int32
	}
	if expires.Valid {
		voucher.ExpiresAt = &expires.Time
	}
	return voucher, err
}
