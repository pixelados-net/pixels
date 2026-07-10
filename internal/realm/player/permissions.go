package player

import "github.com/niflaot/pixels/internal/permission"

var (
	// HotelAmbassador marks a player as a client-visible hotel ambassador.
	HotelAmbassador = permission.RegisterNode("player.hotel.ambassador", "")
)
