package membership

import (
	"context"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbadges "github.com/niflaot/pixels/networking/inbound/group/badge/list"
	inaccept "github.com/niflaot/pixels/networking/inbound/group/membership/accept"
	inadminadd "github.com/niflaot/pixels/networking/inbound/group/membership/admin/add"
	inadminremove "github.com/niflaot/pixels/networking/inbound/group/membership/admin/remove"
	inapprove "github.com/niflaot/pixels/networking/inbound/group/membership/approveall"
	indecline "github.com/niflaot/pixels/networking/inbound/group/membership/decline"
	inclearfavorite "github.com/niflaot/pixels/networking/inbound/group/membership/favorite/clear"
	insetfavorite "github.com/niflaot/pixels/networking/inbound/group/membership/favorite/set"
	injoin "github.com/niflaot/pixels/networking/inbound/group/membership/join"
	inlist "github.com/niflaot/pixels/networking/inbound/group/membership/list"
	inremove "github.com/niflaot/pixels/networking/inbound/group/membership/remove"
	inconfirm "github.com/niflaot/pixels/networking/inbound/group/membership/remove/confirm"
	insearch "github.com/niflaot/pixels/networking/inbound/group/membership/search"
	inunblock "github.com/niflaot/pixels/networking/inbound/group/membership/unblock"
	outbadges "github.com/niflaot/pixels/networking/outbound/group/badge/list"
	outjoinfailed "github.com/niflaot/pixels/networking/outbound/group/identity/joinfailed"
	outlist "github.com/niflaot/pixels/networking/outbound/group/membership/list"
	outfailed "github.com/niflaot/pixels/networking/outbound/group/membership/managementfailed"
	outsearch "github.com/niflaot/pixels/networking/outbound/group/membership/search"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler contains membership packet adapter dependencies.
type Handler struct {
	// Membership manages social roles and preferences.
	Membership *Service
	// Delivery resolves authenticated actors.
	Delivery *groupruntime.Delivery
	// Translations localizes expected errors.
	Translations i18n.Translator
}

// RegisterHandlers registers consumed membership and badge packets.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inlist.Header, handler.list)
	_ = registry.Register(inbadges.Header, handler.badges)
	_ = registry.Register(insearch.Header, handler.search)
	_ = registry.Register(injoin.Header, handler.join)
	_ = registry.Register(inaccept.Header, handler.accept)
	_ = registry.Register(indecline.Header, handler.decline)
	_ = registry.Register(inapprove.Header, handler.approveAll)
	_ = registry.Register(inconfirm.Header, handler.confirmRemoval)
	_ = registry.Register(inremove.Header, handler.remove)
	_ = registry.Register(inadminadd.Header, handler.promote)
	_ = registry.Register(inadminremove.Header, handler.demote)
	_ = registry.Register(insetfavorite.Header, handler.setFavorite)
	_ = registry.Register(inclearfavorite.Header, handler.clearFavorite)
	_ = registry.Register(inunblock.Header, handler.unblock)
}

// unblock safely accepts the compatibility-only packet without mutating state.
func (handler Handler) unblock(_ netconn.Context, packet codec.Packet) error {
	if _, err := inunblock.Decode(packet); err != nil {
		return err
	}
	handler.Membership.metrics.Record(groupobservability.Membership, groupobservability.KindDefault, groupobservability.Unsupported)
	return nil
}

// list sends active memberships for the authenticated player.
func (handler Handler) list(connection netconn.Context, packet codec.Packet) error {
	if err := inlist.Decode(packet); err != nil {
		return err
	}
	return handler.sendGroups(connection, false)
}

// badges sends the relevant group badge map.
func (handler Handler) badges(connection netconn.Context, packet codec.Packet) error {
	if err := inbadges.Decode(packet); err != nil {
		return err
	}
	return handler.sendGroups(connection, true)
}

// sendGroups sends one player snapshot through the requested projection.
func (handler Handler) sendGroups(connection netconn.Context, badges bool) error {
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	groups, err := handler.Membership.PlayerGroups(context.Background(), playerID)
	if err != nil {
		return err
	}
	var response codec.Packet
	if badges {
		response, err = outbadges.Encode(groups)
	} else {
		response, err = outlist.Encode(groups)
	}
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// search sends one viewer-authorized bounded roster page.
func (handler Handler) search(connection netconn.Context, packet codec.Packet) error {
	payload, err := insearch.Decode(packet)
	if err != nil {
		return err
	}
	if payload.Level == 3 {
		payload.Level = 0
	}
	actorID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	page, canManage, err := handler.Membership.MemberPage(context.Background(), actorID, payload.GroupID, payload.Page, payload.Query, payload.Level)
	if err != nil {
		return handler.feedback(connection, err, "group.member.rank_forbidden")
	}
	response, err := outsearch.Encode(page, canManage)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// join creates an immediate membership or exclusive request.
func (handler Handler) join(connection netconn.Context, packet codec.Packet) error {
	groupID, err := injoin.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	_, _, err = handler.Membership.Join(context.Background(), actorID, groupID)
	if err != nil {
		response, encodeErr := outjoinfailed.Encode(joinFailure(err))
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(context.Background(), response)
	}
	if err = handler.sendGroups(connection, false); err != nil {
		return err
	}
	return handler.sendInformation(connection, actorID, groupID)
}

// accept promotes one pending request.
func (handler Handler) accept(connection netconn.Context, packet codec.Packet) error {
	payload, err := inaccept.Decode(packet)
	if err != nil {
		return err
	}
	return handler.manage(connection, payload.GroupID, payload.PlayerID, func(actorID int64) error {
		_, actionErr := handler.Membership.Accept(context.Background(), actorID, payload.GroupID, payload.PlayerID)
		return actionErr
	})
}

// decline removes one pending request.
func (handler Handler) decline(connection netconn.Context, packet codec.Packet) error {
	payload, err := indecline.Decode(packet)
	if err != nil {
		return err
	}
	return handler.manage(connection, payload.GroupID, payload.PlayerID, func(actorID int64) error {
		_, actionErr := handler.Membership.Decline(context.Background(), actorID, payload.GroupID, payload.PlayerID)
		return actionErr
	})
}

// approveAll accepts one bounded request batch.
func (handler Handler) approveAll(connection netconn.Context, packet codec.Packet) error {
	groupID, err := inapprove.Decode(packet)
	if err != nil {
		return err
	}
	return handler.manage(connection, groupID, 0, func(actorID int64) error {
		_, actionErr := handler.Membership.ApproveAll(context.Background(), actorID, groupID)
		return actionErr
	})
}

// promote grants the social administrator role.
func (handler Handler) promote(connection netconn.Context, packet codec.Packet) error {
	payload, err := inadminadd.Decode(packet)
	if err != nil {
		return err
	}
	return handler.changeRole(connection, payload.GroupID, payload.PlayerID, grouprecord.Admin)
}

// demote grants the ordinary member role.
func (handler Handler) demote(connection netconn.Context, packet codec.Packet) error {
	payload, err := inadminremove.Decode(packet)
	if err != nil {
		return err
	}
	return handler.changeRole(connection, payload.GroupID, payload.PlayerID, grouprecord.Member)
}

// changeRole applies one protected role transition.
func (handler Handler) changeRole(connection netconn.Context, groupID int64, playerID int64, role grouprecord.Role) error {
	return handler.manage(connection, groupID, playerID, func(actorID int64) error {
		_, err := handler.Membership.ChangeRole(context.Background(), actorID, groupID, playerID, role)
		return err
	})
}

// manage resolves an actor and maps expected management failures.
func (handler Handler) manage(connection netconn.Context, groupID int64, playerID int64, action func(int64) error) error {
	actorID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	if err = action(actorID); err == nil {
		if playerID > 0 {
			_ = handler.projectInformation(playerID, groupID)
		}
		return nil
	}
	response, encodeErr := outfailed.Encode(groupID, managementFailure(err))
	if encodeErr != nil {
		return encodeErr
	}
	return connection.Send(context.Background(), response)
}
