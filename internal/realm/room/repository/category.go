package repository

import (
	"context"
	"fmt"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

const (
	// categoryColumns contains the shared category select list.
	categoryColumns = `id, caption, caption_key, visible, automatic, automatic_key, global_key, staff_only, order_num, created_at, updated_at, deleted_at, version`

	// listCategoriesSQL reads active room categories.
	listCategoriesSQL = `select ` + categoryColumns + ` from room_categories where deleted_at is null order by order_num asc, caption asc`
)

// ListCategories lists active room categories.
func (repository *Repository) ListCategories(ctx context.Context) ([]roommodel.Category, error) {
	rows, err := repository.executor.Query(ctx, listCategoriesSQL)
	if err != nil {
		return nil, fmt.Errorf("list room categories: %w", err)
	}
	defer rows.Close()

	return scanCategories(rows)
}
