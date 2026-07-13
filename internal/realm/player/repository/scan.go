package repository

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// scanPlayer scans a player row.
func scanPlayer(row pgx.Row) (playermodel.Player, error) {
	var player playermodel.Player
	var deletedAt pgtype.Timestamptz
	var lastLoginAt pgtype.Timestamptz
	var lastLogoutAt pgtype.Timestamptz
	var lastSeenAt pgtype.Timestamptz
	var clubLevel int16
	var clubExpiresAt pgtype.Timestamptz

	err := row.Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt, &deletedAt, &player.Version.Version, &lastLoginAt, &lastLogoutAt, &lastSeenAt, &clubLevel, &clubExpiresAt, &player.AllowTrade)
	if err != nil {
		return playermodel.Player{}, err
	}

	player.DeletedAt = timePointer(deletedAt)
	player.LastLoginAt = timePointer(lastLoginAt)
	player.LastLogoutAt = timePointer(lastLogoutAt)
	player.LastSeenAt = timePointer(lastSeenAt)
	player.Club.Level = playermodel.ClubLevel(clubLevel)
	player.Club.ExpiresAt = timePointer(clubExpiresAt)

	return player, nil
}

// scanProfile scans a player profile row.
func scanProfile(row pgx.Row) (playermodel.Profile, error) {
	var profile playermodel.Profile
	var homeRoomID pgtype.Int8
	var gender string

	err := row.Scan(&profile.PlayerID, &profile.Look, &gender, &profile.Motto, &homeRoomID, &profile.AllowNameChange, &profile.BubbleStyle, &profile.BlockFriendRequests, &profile.BlockRoomInvites, &profile.BlockFollowing, &profile.CreatedAt, &profile.UpdatedAt, &profile.Version.Version)
	if err != nil {
		return playermodel.Profile{}, err
	}

	profile.Gender = playermodel.Gender(gender)
	profile.HomeRoomID = int64Pointer(homeRoomID)

	return profile, nil
}

// timePointer converts a PostgreSQL timestamp to an optional time.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

// int64Pointer converts a PostgreSQL int8 to an optional int64.
func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}
