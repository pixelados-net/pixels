// Package calendar contains the calendar door opened event.
package calendar

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies one claimed calendar door.
	Name bus.Name = "subscription.calendar.door_opened"
)

// Payload contains one calendar door claim.
type Payload struct {
	// PlayerID identifies the beneficiary.
	PlayerID int64
	// CampaignID identifies the campaign.
	CampaignID int64
	// Day stores the claimed door number.
	Day int32
}
