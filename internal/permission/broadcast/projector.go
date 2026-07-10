// Package broadcast projects permission changes to live player connections.
package broadcast

import (
	"context"
	"sort"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	realmplayer "github.com/niflaot/pixels/internal/realm/player"
	"github.com/niflaot/pixels/networking/codec"
	outperks "github.com/niflaot/pixels/networking/outbound/session/perks"
	outpermissions "github.com/niflaot/pixels/networking/outbound/session/permissions"
)

const (
	// defaultClubLevel stores the unsupported club subscription level.
	defaultClubLevel int32 = 0
)

// Projector builds and sends Nitro permission state.
type Projector struct {
	// permissions resolves player permission state.
	permissions permissionservice.Manager
}

// NewProjector creates a permission protocol projector.
func NewProjector(permissions permissionservice.Manager) *Projector {
	return &Projector{permissions: permissions}
}

// Packets builds one player's permission protocol state.
func (projector *Projector) Packets(ctx context.Context, playerID int64) ([]codec.Packet, error) {
	group, found, err := projector.permissions.PrimaryGroup(ctx, playerID)
	if err != nil {
		return nil, err
	}
	securityLevel := int32(0)
	if found {
		securityLevel = group.Weight
	}
	ambassador, err := projector.permissions.HasPermission(ctx, playerID, realmplayer.HotelAmbassador)
	if err != nil {
		return nil, err
	}
	packet, err := outpermissions.Encode(defaultClubLevel, securityLevel, ambassador)
	if err != nil {
		return nil, err
	}
	packets := []codec.Packet{packet}

	entries, err := projector.perks(ctx, playerID)
	if err != nil {
		return nil, err
	}
	packet, err = outperks.Encode(entries)
	if err != nil {
		return nil, err
	}

	return append(packets, packet), nil
}

// perks resolves every registered client perk into one allowance.
func (projector *Projector) perks(ctx context.Context, playerID int64) ([]outperks.Entry, error) {
	allowances := make(map[string]bool)
	for _, registration := range permission.RegisteredNodes() {
		if registration.PerkName == "" {
			continue
		}
		allowed, err := projector.permissions.HasPermission(ctx, playerID, registration.Node)
		if err != nil {
			return nil, err
		}
		allowances[registration.PerkName] = allowances[registration.PerkName] || allowed
	}
	codes := make([]string, 0, len(allowances))
	for code := range allowances {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	entries := make([]outperks.Entry, 0, len(codes))
	for _, code := range codes {
		entries = append(entries, outperks.Entry{Code: code, Allowed: allowances[code]})
	}

	return entries, nil
}
