package repository

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

// scanGroup scans one permission group row.
func scanGroup(row pgx.Row) (permissionmodel.Group, error) {
	var group permissionmodel.Group
	var roomEffectID pgtype.Int4
	var parentID pgtype.Int8
	var deletedAt pgtype.Timestamptz
	err := row.Scan(&group.ID, &group.Name, &group.Weight, &group.Prefix, &group.PrefixColor, &roomEffectID, &parentID,
		&group.CreatedAt, &group.UpdatedAt, &deletedAt, &group.Version.Version)
	if roomEffectID.Valid {
		group.RoomEffectID = &roomEffectID.Int32
	}
	if parentID.Valid {
		group.ParentGroupID = &parentID.Int64
	}
	if deletedAt.Valid {
		group.DeletedAt = &deletedAt.Time
	}

	return group, err
}

// scanGroups scans permission groups from rows.
func scanGroups(rows pgx.Rows) ([]permissionmodel.Group, error) {
	groups := make([]permissionmodel.Group, 0)
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

// scanGrants scans permission grants from rows.
func scanGrants(rows pgx.Rows) ([]permissionmodel.Grant, error) {
	grants := make([]permissionmodel.Grant, 0)
	for rows.Next() {
		var grant permissionmodel.Grant
		var node string
		if err := rows.Scan(&node, &grant.Allowed); err != nil {
			return nil, err
		}
		grant.Node = permission.Node(node)
		grants = append(grants, grant)
	}

	return grants, rows.Err()
}
