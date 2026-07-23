package live

// SetMount attaches or detaches one player unit and one pet unit in the shared room world.
func (room *Room) SetMount(riderKey int64, mountKey int64, mounted bool) (UnitSnapshot, UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, UnitSnapshot{}, ErrWorldNotLoaded
	}

	return room.world.SetMount(riderKey, mountKey, mounted)
}
