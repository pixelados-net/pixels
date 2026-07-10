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
	// clientPerks stores immutable client perk node groups.
	clientPerks []perkRegistration
}

// NewProjector creates a permission protocol projector.
func NewProjector(permissions permissionservice.Manager) *Projector {
	return &Projector{permissions: permissions, clientPerks: registeredPerks()}
}

// perkRegistration groups nodes that unlock the same client perk.
type perkRegistration struct {
	// code identifies the Nitro perk.
	code string
	// nodes stores permission nodes mapped to the perk.
	nodes []permission.Node
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
	entries := make([]outperks.Entry, len(projector.clientPerks))
	for index, perk := range projector.clientPerks {
		allowed := false
		for _, node := range perk.nodes {
			current, err := projector.permissions.HasPermission(ctx, playerID, node)
			if err != nil {
				return nil, err
			}
			allowed = allowed || current
		}
		entries[index] = outperks.Entry{Code: perk.code, Allowed: allowed}
	}

	return entries, nil
}

// registeredPerks snapshots registered client perk mappings in stable order.
func registeredPerks() []perkRegistration {
	indices := make(map[string]int)
	perks := make([]perkRegistration, 0)
	for _, registration := range permission.RegisteredNodes() {
		if registration.PerkName == "" {
			continue
		}
		index, found := indices[registration.PerkName]
		if !found {
			index = len(perks)
			indices[registration.PerkName] = index
			perks = append(perks, perkRegistration{code: registration.PerkName})
		}
		perks[index].nodes = append(perks[index].nodes, registration.Node)
	}
	sort.Slice(perks, func(left int, right int) bool {
		return perks[left].code < perks[right].code
	})

	return perks
}
