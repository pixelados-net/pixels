package model

import "time"

// BanDuration names one Nitro room ban duration.
type BanDuration string

const (
	// BanDurationHour bans for one hour.
	BanDurationHour BanDuration = "RWUAM_BAN_USER_HOUR"
	// BanDurationDay bans for one day.
	BanDurationDay BanDuration = "RWUAM_BAN_USER_DAY"
	// BanDurationPermanent bans for ten years.
	BanDurationPermanent BanDuration = "RWUAM_BAN_USER_PERM"
)

// Duration returns the concrete ban duration.
func (duration BanDuration) Duration() (time.Duration, bool) {
	switch duration {
	case BanDurationHour:
		return time.Hour, true
	case BanDurationDay:
		return 24 * time.Hour, true
	case BanDurationPermanent:
		return 10 * 365 * 24 * time.Hour, true
	default:
		return 0, false
	}
}
