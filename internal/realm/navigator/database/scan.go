package database

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

// scanFavoriteRoomIDs scans favorite room id rows.
func scanFavoriteRoomIDs(rows pgx.Rows) ([]int64, error) {
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan favorite room id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// scanSavedSearch scans a saved search row.
func scanSavedSearch(row pgx.Row) (navmodel.SavedSearch, error) {
	var search navmodel.SavedSearch
	err := row.Scan(&search.ID, &search.PlayerID, &search.Code, &search.Filter, &search.Localization, &search.CreatedAt)
	if err != nil {
		return navmodel.SavedSearch{}, err
	}

	return search, nil
}

// scanSavedSearches scans saved search rows.
func scanSavedSearches(rows pgx.Rows) ([]navmodel.SavedSearch, error) {
	var searches []navmodel.SavedSearch
	for rows.Next() {
		search, err := scanSavedSearch(rows)
		if err != nil {
			return nil, fmt.Errorf("scan saved search: %w", err)
		}
		searches = append(searches, search)
	}

	return searches, rows.Err()
}

// scanPreference scans a navigator preference row.
func scanPreference(row pgx.Row) (navmodel.Preference, error) {
	var preference navmodel.Preference
	err := row.Scan(&preference.PlayerID, &preference.WindowX, &preference.WindowY, &preference.WindowWidth, &preference.WindowHeight, &preference.LeftPanelHidden, &preference.ResultsMode, &preference.CreatedAt, &preference.UpdatedAt)
	if err != nil {
		return navmodel.Preference{}, err
	}

	return preference, nil
}

// scanCategoryPreference scans a category preference row.
func scanCategoryPreference(row pgx.Row) (navmodel.CategoryPreference, error) {
	var preference navmodel.CategoryPreference
	err := row.Scan(&preference.PlayerID, &preference.Code, &preference.Collapsed, &preference.ListMode, &preference.CreatedAt, &preference.UpdatedAt)
	if err != nil {
		return navmodel.CategoryPreference{}, err
	}

	return preference, nil
}

// scanCategoryPreferences scans category preference rows.
func scanCategoryPreferences(rows pgx.Rows) ([]navmodel.CategoryPreference, error) {
	var preferences []navmodel.CategoryPreference
	for rows.Next() {
		preference, err := scanCategoryPreference(rows)
		if err != nil {
			return nil, fmt.Errorf("scan category preference: %w", err)
		}
		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

// scanLiftedRooms scans lifted room rows.
func scanLiftedRooms(rows pgx.Rows) ([]navmodel.LiftedRoom, error) {
	var rooms []navmodel.LiftedRoom
	for rows.Next() {
		room, err := scanLiftedRoom(rows)
		if err != nil {
			return nil, fmt.Errorf("scan lifted room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

// scanLiftedRoom scans a lifted room row.
func scanLiftedRoom(row pgx.Row) (navmodel.LiftedRoom, error) {
	var room navmodel.LiftedRoom
	var startsAt pgtype.Timestamptz
	var endsAt pgtype.Timestamptz
	var deletedAt pgtype.Timestamptz
	err := row.Scan(&room.ID, &room.RoomID, &room.AreaID, &room.Image, &room.Caption, &room.Order, &startsAt, &endsAt, &room.CreatedAt, &room.UpdatedAt, &deletedAt, &room.Version.Version)
	if err != nil {
		return navmodel.LiftedRoom{}, err
	}
	room.StartsAt = timePointer(startsAt)
	room.EndsAt = timePointer(endsAt)
	room.DeletedAt = timePointer(deletedAt)

	return room, nil
}

// timePointer converts a PostgreSQL timestamp to an optional time.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
