package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

const (
	// pageColumns stores the shared catalog page projection.
	pageColumns = `id, parent_id, name, layout, icon_color, icon_image, required_node, order_num, visible, enabled, club_only, new_additions, expires_at, excluded_from_kickback, created_at, updated_at, deleted_at, version`

	// listPagesSQL lists active catalog pages.
	listPagesSQL = `select ` + pageColumns + ` from catalog_pages where deleted_at is null order by parent_id nulls first, order_num, id`

	// findPageSQL finds one active catalog page.
	findPageSQL = `select ` + pageColumns + ` from catalog_pages where id = $1 and deleted_at is null`

	// createPageSQL creates one catalog page.
	createPageSQL = `insert into catalog_pages (parent_id, name, layout, icon_color, icon_image, required_node, order_num, visible, enabled, club_only, new_additions, expires_at, excluded_from_kickback)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) returning ` + pageColumns

	// updatePageSQL updates one active catalog page using its version.
	updatePageSQL = `update catalog_pages set parent_id=$2, name=$3, layout=$4, icon_color=$5, icon_image=$6,
required_node=$7, order_num=$8, visible=$9, enabled=$10, club_only=$11, new_additions=$12, expires_at=$13,
excluded_from_kickback=$14, updated_at=now(), version=version+1 where id=$1 and version=$15 and deleted_at is null returning ` + pageColumns
)

// ListPages lists every active catalog page.
func (repository *Repository) ListPages(ctx context.Context) ([]catalogmodel.Page, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, listPagesSQL)
	if err != nil {
		return nil, fmt.Errorf("list catalog pages: %w", err)
	}
	defer rows.Close()

	return scanPages(rows)
}

// FindPageByID finds one active catalog page.
func (repository *Repository) FindPageByID(ctx context.Context, id int64) (catalogmodel.Page, bool, error) {
	return repository.queryPage(ctx, findPageSQL, id)
}

// CreatePage creates one catalog page.
func (repository *Repository) CreatePage(ctx context.Context, page catalogmodel.Page) (catalogmodel.Page, error) {
	created, err := scanPage(repository.executorFor(ctx).QueryRow(ctx, createPageSQL, pageValues(page)...))
	if err != nil {
		return catalogmodel.Page{}, fmt.Errorf("create catalog page %q: %w", page.Name, err)
	}

	return created, nil
}

// UpdatePage updates one page using optimistic locking.
func (repository *Repository) UpdatePage(ctx context.Context, page catalogmodel.Page) (catalogmodel.Page, bool, error) {
	arguments := append([]any{page.ID}, pageValues(page)...)
	arguments = append(arguments, page.Version.Version)

	return repository.queryPage(ctx, updatePageSQL, arguments...)
}

// queryPage scans one optional page.
func (repository *Repository) queryPage(ctx context.Context, query string, arguments ...any) (catalogmodel.Page, bool, error) {
	page, err := scanPage(repository.executorFor(ctx).QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return catalogmodel.Page{}, false, nil
	}
	if err != nil {
		return catalogmodel.Page{}, false, err
	}

	return page, true, nil
}

// pageValues maps page persistence values in statement order.
func pageValues(page catalogmodel.Page) []any {
	return []any{page.ParentID, page.Name, page.Layout, page.IconColor, page.IconImage, page.RequiredNode, page.OrderNum, page.Visible, page.Enabled, page.ClubOnly, page.NewAdditions, page.ExpiresAt, page.ExcludedFromKickback}
}
