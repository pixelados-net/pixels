package event

import sdkplayer "github.com/niflaot/pixels/sdk/player"

// PlayerConnectedName identifies the authenticated-player notification.
const PlayerConnectedName = "player.connected"

// PlayerConnected fires after a player has an authenticated live session.
type PlayerConnected struct {
	// Player stores the immutable connected-player snapshot.
	Player sdkplayer.Player
}

// Name returns the stable player event identifier.
func (*PlayerConnected) Name() string { return PlayerConnectedName }
