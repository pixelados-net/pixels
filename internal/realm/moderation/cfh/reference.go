package cfh

import playerlive "github.com/niflaot/pixels/internal/realm/player/live"

// positiveReference converts one positive wire identifier into an optional domain reference.
func positiveReference(value int32) *int64 {
	if value <= 0 {
		return nil
	}
	converted := int64(value)
	return &converted
}

// reportRoomReference resolves Nitro's missing room identifier from live reporter presence.
func reportRoomReference(players *playerlive.Registry, actorID int64, wireRoomID int32) *int64 {
	if value := positiveReference(wireRoomID); value != nil {
		return value
	}
	if players == nil {
		return nil
	}
	actor, found := players.Find(actorID)
	if !found {
		return nil
	}
	roomID, present := actor.CurrentRoom()
	if !present {
		return nil
	}
	return &roomID
}
