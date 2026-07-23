// Package calendar executes campaign calendar requests.
package calendar

import (
	"github.com/niflaot/pixels/internal/command"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies calendar commands.
	Name command.Name = "subscription.calendar"
)

// Action identifies one calendar operation.
type Action uint8

const (
	// Open requests a normal calendar door claim.
	Open Action = iota
	// OpenStaff requests a staff calendar door claim.
	OpenStaff
	// Seasonal requests today's seasonal catalog offer.
	Seasonal
)

// Command contains one calendar request.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Action identifies requested behavior.
	Action Action
	// CampaignName identifies the calendar campaign.
	CampaignName string
	// DayNumber identifies the requested door.
	DayNumber int32
}

// Handler executes calendar commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections.
	Bindings *binding.Registry
	// Subscriptions manages calendar rewards.
	Subscriptions *core.Service
	// Catalog reads seasonal catalog offers.
	Catalog *catalogservice.Service
	// Furniture reads furniture product names.
	Furniture furnitureservice.DefinitionFinder
	// Permissions resolves staff bypass capability.
	Permissions permissionservice.Checker
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddUint8("action", uint8(input.Action))
	encoder.AddString("campaign", input.CampaignName)
	encoder.AddInt32("day", input.DayNumber)

	return nil
}
