// Package wired handles authorized room WIRED editor commands.
package wired

import (
	"errors"

	"github.com/niflaot/pixels/internal/command"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/inbound/furniture/wired/common"
)

const (
	// OpenName identifies WIRED editor open commands.
	OpenName command.Name = "room.wired.open"
	// SaveName identifies WIRED editor save commands.
	SaveName command.Name = "room.wired.save"
	// SnapshotName identifies WIRED snapshot capture commands.
	SnapshotName command.Name = "room.wired.snapshot"
)

var (
	// ErrNoRights reports a player without room WIRED configuration rights.
	ErrNoRights = errors.New("player cannot configure room WIRED")
	// ErrNotInRoom reports a WIRED request without active room presence.
	ErrNotInRoom = errors.New("player is not in a room")
)

// Family identifies the save packet family.
type Family uint8

const (
	// TriggerFamily saves trigger settings.
	TriggerFamily Family = iota + 1
	// EffectFamily saves effect settings.
	EffectFamily
	// ConditionFamily saves condition settings.
	ConditionFamily
)

// OpenCommand requests one WIRED editor definition.
type OpenCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// ItemID identifies WIRED furniture.
	ItemID int64
}

// SaveCommand stores one decoded editor configuration.
type SaveCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Family identifies the editor packet family.
	Family Family
	// Settings stores shared Nitro save fields.
	Settings common.Settings
}

// SnapshotCommand captures selected furniture state.
type SnapshotCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// ItemID identifies the snapshot action.
	ItemID int64
}

// CommandName returns the stable command name.
func (OpenCommand) CommandName() command.Name { return OpenName }

// CommandName returns the stable command name.
func (SaveCommand) CommandName() command.Name { return SaveName }

// CommandName returns the stable command name.
func (SnapshotCommand) CommandName() command.Name { return SnapshotName }
