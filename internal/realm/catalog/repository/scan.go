package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niflaot/pixels/internal/permission"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

const (
	// definitionColumns stores the furniture projection used by sanitation.
	definitionColumns = `d.id, d.sprite_id, d.name, d.public_name, d.kind, d.width, d.length, d.stack_height::float8, d.allow_stack, d.allow_walk, d.allow_sit, d.allow_lay, d.allow_inventory_stack, d.interaction_type, d.interaction_modes_count, d.multiheight, d.custom_params, d.metadata, d.created_at, d.updated_at, d.deleted_at, d.version`
)

// scanPage scans one catalog page.
func scanPage(row pgx.Row) (catalogmodel.Page, error) {
	var page catalogmodel.Page
	var parentID pgtype.Int8
	var requiredNode pgtype.Text
	var expiresAt pgtype.Timestamptz
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&page.ID, &parentID, &page.Name, &page.Layout, &page.IconColor, &page.IconImage,
		&requiredNode, &page.OrderNum, &page.Visible, &page.Enabled, &page.ClubOnly, &page.NewAdditions,
		&expiresAt, &page.ExcludedFromKickback,
		&page.CreatedAt, &page.UpdatedAt, &deletedAt, &page.Version.Version,
	)
	if err != nil {
		return catalogmodel.Page{}, err
	}
	page.ParentID = int64Pointer(parentID)
	if requiredNode.Valid {
		node := permission.Node(requiredNode.String)
		page.RequiredNode = &node
	}
	page.DeletedAt = timePointer(deletedAt)
	page.ExpiresAt = timePointer(expiresAt)

	return page, nil
}

// scanPages scans catalog page rows.
func scanPages(rows pgx.Rows) ([]catalogmodel.Page, error) {
	pages := make([]catalogmodel.Page, 0)
	for rows.Next() {
		page, err := scanPage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog page: %w", err)
		}
		pages = append(pages, page)
	}

	return pages, rows.Err()
}

// scanItem scans one catalog offer.
func scanItem(row pgx.Row) (catalogmodel.Item, error) {
	var item catalogmodel.Item
	var definitionID pgtype.Int8
	var petTypeID pgtype.Int4
	var templateRoomID pgtype.Int8
	var grantsEffectID pgtype.Int4
	var scheduledAt pgtype.Timestamptz
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&item.ID, &item.PageID, &definitionID, &item.RewardKind, &petTypeID, &item.PetProductCode, &templateRoomID, &grantsEffectID, &item.GrantsEffectDurationSeconds, &item.Name, &item.CostCredits,
		&item.CostPoints, &item.PointsType, &item.Amount, &item.LimitedStack,
		&item.LimitedSells, &item.BundleDiscountEnabled, &item.Giftable, &item.ClubOnly, &item.OrderNum, &item.Enabled,
		&item.ExtraData, &scheduledAt, &item.CreatedAt, &item.UpdatedAt, &deletedAt, &item.Version.Version,
	)
	if err != nil {
		return catalogmodel.Item{}, err
	}
	if definitionID.Valid {
		item.DefinitionID = definitionID.Int64
	}
	if petTypeID.Valid {
		item.PetTypeID = &petTypeID.Int32
	}
	item.RoomBundleTemplateRoomID = int64Pointer(templateRoomID)
	if grantsEffectID.Valid {
		item.GrantsEffectID = &grantsEffectID.Int32
	}
	item.ScheduledAt = timePointer(scheduledAt)
	item.DeletedAt = timePointer(deletedAt)

	return item, nil
}

// scanItems scans catalog offer rows.
func scanItems(rows pgx.Rows) ([]catalogmodel.Item, error) {
	items := make([]catalogmodel.Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// scanDefinition scans one furniture definition for catalog sanitation.
func scanDefinition(row pgx.Row) (furnituremodel.Definition, error) {
	var definition furnituremodel.Definition
	var kind string
	var metadata []byte
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&definition.ID, &definition.SpriteID, &definition.Name, &definition.PublicName, &kind,
		&definition.Width, &definition.Length, &definition.StackHeight, &definition.AllowStack,
		&definition.AllowWalk, &definition.AllowSit, &definition.AllowLay, &definition.AllowInventoryStack,
		&definition.InteractionType, &definition.InteractionModesCount, &definition.Multiheight,
		&definition.CustomParams, &metadata, &definition.CreatedAt, &definition.UpdatedAt,
		&deletedAt, &definition.Version.Version,
	)
	if err != nil {
		return furnituremodel.Definition{}, err
	}
	definition.Kind = furnituremodel.Kind(kind)
	definition.Metadata = json.RawMessage(metadata)
	definition.DeletedAt = timePointer(deletedAt)

	return definition, nil
}

// scanDefinitions scans furniture definitions for catalog sanitation.
func scanDefinitions(rows pgx.Rows) ([]furnituremodel.Definition, error) {
	definitions := make([]furnituremodel.Definition, 0)
	for rows.Next() {
		definition, err := scanDefinition(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog sanitize definition: %w", err)
		}
		definitions = append(definitions, definition)
	}

	return definitions, rows.Err()
}

// int64Pointer converts a nullable bigint into a pointer.
func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}

// timePointer converts a nullable timestamp into a pointer.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
