// Package mount owns bidirectional movement links between room units.
package mount

import (
	"errors"

	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

var (
	// ErrInvalid reports incompatible or already-linked riding units.
	ErrInvalid = errors.New("invalid room world mount")
)

// State stores rider-to-mount relationships for one room world.
type State struct {
	// mounts stores the mounted entity key for each rider entity key.
	mounts map[int64]int64
	// riders stores the rider entity key for each mounted entity key.
	riders map[int64]int64
}

// New creates empty mount state.
func New() *State {
	return &State{mounts: make(map[int64]int64), riders: make(map[int64]int64)}
}

// Set attaches or detaches one player and pet unit.
func (state *State) Set(units map[int64]*worldunit.Unit, riderKey int64, mountKey int64, mounted bool) (*worldunit.Unit, *worldunit.Unit, error) {
	rider, riderFound := units[riderKey]
	mountUnit, mountFound := units[mountKey]
	if !riderFound || !mountFound {
		return nil, nil, nil
	}
	if rider.Kind() != worldunit.KindPlayer || mountUnit.Kind() != worldunit.KindPet {
		return nil, nil, ErrInvalid
	}
	if !mounted {
		if state.mounts[riderKey] == mountKey {
			state.Unlink(units, riderKey)
		}
		return rider, mountUnit, nil
	}
	if current, found := state.mounts[riderKey]; found && current != mountKey {
		return nil, nil, ErrInvalid
	}
	if current, found := state.riders[mountKey]; found && current != riderKey {
		return nil, nil, ErrInvalid
	}
	state.mounts[riderKey] = mountKey
	state.riders[mountKey] = riderKey
	mountPosition := mountUnit.Position()
	mountUnit.Reposition(mountPosition, mountUnit.BodyRotation())
	mountUnit.SetControl(worldunit.ControlNone)
	rider.Reposition(mountPosition, mountUnit.BodyRotation())
	rider.SetControl(worldunit.ControlNone)
	rider.SetRenderOffset(worldunit.RidingHeightOffset)

	return rider, mountUnit, nil
}

// Linked returns the entity sharing authoritative movement with one mounted unit.
func (state *State) Linked(entityKey int64) (int64, bool) {
	if mountKey, found := state.mounts[entityKey]; found {
		return mountKey, true
	}
	riderKey, found := state.riders[entityKey]
	return riderKey, found
}

// Unlink clears either side of one rider and pet relationship.
func (state *State) Unlink(units map[int64]*worldunit.Unit, entityKey int64) {
	riderKey, mountKey, found := state.keys(entityKey)
	if !found {
		return
	}
	delete(state.mounts, riderKey)
	delete(state.riders, mountKey)
	if rider, exists := units[riderKey]; exists {
		rider.StopMovement()
		rider.SetRenderOffset(0)
		rider.SetControl(worldunit.ControlNone)
	}
	if mountUnit, exists := units[mountKey]; exists {
		mountUnit.StopMovement()
		mountUnit.SetControl(worldunit.ControlNone)
	}
}

// keys resolves normalized rider and mount keys from either side.
func (state *State) keys(entityKey int64) (int64, int64, bool) {
	if mountKey, found := state.mounts[entityKey]; found {
		return entityKey, mountKey, true
	}
	riderKey, found := state.riders[entityKey]
	return riderKey, entityKey, found
}
