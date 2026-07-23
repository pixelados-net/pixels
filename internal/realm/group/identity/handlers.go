package identity

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/group/membership"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incolors "github.com/niflaot/pixels/networking/inbound/group/badge/colors"
	inparts "github.com/niflaot/pixels/networking/inbound/group/badge/parts"
	insavebadge "github.com/niflaot/pixels/networking/inbound/group/badge/save"
	increate "github.com/niflaot/pixels/networking/inbound/group/identity/create"
	inoptions "github.com/niflaot/pixels/networking/inbound/group/identity/create/options"
	indelete "github.com/niflaot/pixels/networking/inbound/group/identity/delete"
	ininfo "github.com/niflaot/pixels/networking/inbound/group/identity/info"
	inpreferences "github.com/niflaot/pixels/networking/inbound/group/identity/preferences"
	insave "github.com/niflaot/pixels/networking/inbound/group/identity/save"
	insettings "github.com/niflaot/pixels/networking/inbound/group/identity/settings"
	outparts "github.com/niflaot/pixels/networking/outbound/group/badge/parts"
	outoptions "github.com/niflaot/pixels/networking/outbound/group/identity/create/options"
	outdetails "github.com/niflaot/pixels/networking/outbound/group/identity/detailschanged"
	outinfo "github.com/niflaot/pixels/networking/outbound/group/identity/info"
	outpurchased "github.com/niflaot/pixels/networking/outbound/group/identity/purchased"
	outsettings "github.com/niflaot/pixels/networking/outbound/group/identity/settings"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler contains identity packet adapter dependencies.
type Handler struct {
	// Identity manages group identity behavior.
	Identity *Service
	// Membership resolves viewer social state.
	Membership *membership.Service
	// Delivery resolves authenticated actors.
	Delivery *groupruntime.Delivery
	// Translations localizes expected errors.
	Translations i18n.Translator
}

// RegisterHandlers registers all group identity and badge packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inoptions.Header, handler.options)
	_ = registry.Register(inparts.Header, handler.parts)
	_ = registry.Register(increate.Header, handler.create)
	_ = registry.Register(ininfo.Header, handler.info)
	_ = registry.Register(insettings.Header, handler.settings)
	_ = registry.Register(insave.Header, handler.save)
	_ = registry.Register(incolors.Header, handler.colors)
	_ = registry.Register(inpreferences.Header, handler.preferences)
	_ = registry.Register(insavebadge.Header, handler.saveBadge)
	_ = registry.Register(indelete.Header, handler.deactivate)
}

// options handles creator option requests.
func (handler Handler) options(connection netconn.Context, packet codec.Packet) error {
	if err := inoptions.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	cost, rooms, err := handler.Identity.Options(context.Background(), playerID)
	if err != nil {
		return handler.feedback(connection, "group.create.options_failed")
	}
	response, err := outoptions.Encode(cost, rooms)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// parts handles badge editor reference requests.
func (handler Handler) parts(connection netconn.Context, packet codec.Packet) error {
	if err := inparts.Decode(packet); err != nil {
		return err
	}
	snapshot, err := handler.Identity.BadgeRegistry(context.Background())
	if err != nil {
		return err
	}
	response, err := outparts.Encode(snapshot)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// create handles one atomic group purchase.
func (handler Handler) create(connection netconn.Context, packet codec.Packet) error {
	payload, err := increate.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	group, err := handler.Identity.Create(context.Background(), CreateParams{OwnerPlayerID: playerID, Name: payload.Name, Description: payload.Description, HomeRoomID: payload.RoomID, ColorA: payload.ColorA, ColorB: payload.ColorB, BadgeParts: payload.Parts})
	if err != nil {
		return handler.feedback(connection, errorKey(err, "group.create.room_ineligible"))
	}
	response, err := outpurchased.Encode(group.HomeRoomID, group.ID)
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), response); err != nil {
		return err
	}
	return handler.sendInfo(connection, playerID, group.ID, true)
}

// info handles standalone and room group information requests.
func (handler Handler) info(connection netconn.Context, packet codec.Packet) error {
	payload, err := ininfo.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	return handler.sendInfo(connection, playerID, payload.GroupID, payload.Flag)
}

// sendInfo resolves and sends one viewer-specific info packet.
func (handler Handler) sendInfo(connection netconn.Context, playerID int64, groupID int64, flag bool) error {
	group, found, err := handler.Identity.Group(context.Background(), groupID, false)
	if err != nil || !found {
		return handler.feedback(connection, "group.deactivated")
	}
	role, member, pending, favorite, err := handler.Membership.Status(context.Background(), playerID, groupID)
	if err != nil {
		return err
	}
	response, err := outinfo.Encode(group, role, member, pending, favorite, flag, member && role <= grouprecord.Admin)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// settings handles owner manager data requests.
func (handler Handler) settings(connection netconn.Context, packet codec.Packet) error {
	groupID, err := insettings.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	group, found, err := handler.Identity.Group(context.Background(), groupID, false)
	if err != nil || !found || group.OwnerPlayerID != playerID {
		return handler.feedback(connection, "group.edit.forbidden")
	}
	parts, err := handler.Identity.Parts(context.Background(), groupID)
	if err != nil {
		return err
	}
	response, err := outsettings.Encode(group, parts)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// save handles owner identity edits.
func (handler Handler) save(connection netconn.Context, packet codec.Packet) error {
	payload, err := insave.Decode(packet)
	if err != nil {
		return err
	}
	name, description := payload.Name, payload.Description
	return handler.update(connection, payload.GroupID, grouprecord.GroupPatch{Name: &name, Description: &description})
}

// colors handles owner color edits.
func (handler Handler) colors(connection netconn.Context, packet codec.Packet) error {
	payload, err := incolors.Decode(packet)
	if err != nil {
		return err
	}
	return handler.update(connection, payload.GroupID, grouprecord.GroupPatch{ColorA: &payload.ColorA, ColorB: &payload.ColorB})
}

// preferences handles state and decoration edits.
func (handler Handler) preferences(connection netconn.Context, packet codec.Packet) error {
	payload, err := inpreferences.Decode(packet)
	if err != nil {
		return err
	}
	state := grouprecord.State(payload.State)
	decorate := payload.OnlyAdminsDecorate == 0
	return handler.update(connection, payload.GroupID, grouprecord.GroupPatch{State: &state, CanMembersDecorate: &decorate})
}

// update serializes a Nitro edit against the latest version.
func (handler Handler) update(connection netconn.Context, groupID int64, patch grouprecord.GroupPatch) error {
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	group, found, err := handler.Identity.Group(context.Background(), groupID, false)
	if err != nil || !found {
		return handler.feedback(connection, "group.deactivated")
	}
	if _, err = handler.Identity.Update(context.Background(), playerID, groupID, group.Version, patch); err != nil {
		return handler.feedback(connection, errorKey(err, "group.edit.conflict"))
	}
	response, err := outdetails.Encode(groupID)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}
