// Package search owns Navigator search routing, including explicit legacy NOOP aliases.
// Nitro 1.6.6 declares all fourteen legacy composers but nitro-react has no call
// site for any of them; Arcturus and Comet use only generic category dispatch.
// Guild bases are additionally opened through TryVisitRoom without Navigator.
// Their real equivalents remain NAVIGATOR_SEARCH codes such as official_view,
// my_rooms, favorites, hotel_view, and the generic text-query fallback.
package search

import (
	"context"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	infavourites "github.com/niflaot/pixels/networking/inbound/navigator/legacy/favouriterooms"
	infrequent "github.com/niflaot/pixels/networking/inbound/navigator/legacy/frequenthistory"
	infriendsowned "github.com/niflaot/pixels/networking/inbound/navigator/legacy/friendsownedrooms"
	infriends "github.com/niflaot/pixels/networking/inbound/navigator/legacy/friendsrooms"
	inguilds "github.com/niflaot/pixels/networking/inbound/navigator/legacy/guildbases"
	inguildsearch "github.com/niflaot/pixels/networking/inbound/navigator/legacy/guildbasesearch"
	inhighest "github.com/niflaot/pixels/networking/inbound/navigator/legacy/highestscore"
	inmyrooms "github.com/niflaot/pixels/networking/inbound/navigator/legacy/myrooms"
	inofficial "github.com/niflaot/pixels/networking/inbound/navigator/legacy/officialrooms"
	inpopular "github.com/niflaot/pixels/networking/inbound/navigator/legacy/popularrooms"
	inrecommended "github.com/niflaot/pixels/networking/inbound/navigator/legacy/recommendedrooms"
	inhistory "github.com/niflaot/pixels/networking/inbound/navigator/legacy/roomhistory"
	inrights "github.com/niflaot/pixels/networking/inbound/navigator/legacy/roomrights"
	intext "github.com/niflaot/pixels/networking/inbound/navigator/legacy/roomtextsearch"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/browse/searchresult"
)

// legacyAliasHeaders lists the unreachable dedicated search composers.
var legacyAliasHeaders = [...]uint16{
	inguilds.Header, inrights.Header, infrequent.Header, inofficial.Header,
	infriends.Header, inhistory.Header, infriendsowned.Header, inmyrooms.Header,
	inrecommended.Header, infavourites.Header, inpopular.Header,
	inguildsearch.Header, inhighest.Header, intext.Header,
}

// legacyAlias stores the canonical generic-search identity echoed to Nitro.
type legacyAlias struct {
	// code identifies the existing NAVIGATOR_SEARCH equivalent.
	code string
	// data preserves a text filter when the legacy composer supplied one.
	data string
}

// NewLegacyAliasHandler creates the stateless empty-result compatibility handler.
func NewLegacyAliasHandler() netconn.Handler {
	return func(connection netconn.Context, packet codec.Packet) error {
		alias, err := decodeLegacyAlias(packet)
		if err != nil {
			return err
		}
		response, err := outsearch.Encode(alias.code, alias.data, nil)
		if err != nil {
			return err
		}
		return connection.Send(context.Background(), response)
	}
}

// RegisterLegacyAliases registers all unreachable composers against one NOOP handler.
func RegisterLegacyAliases(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	if registry == nil || handler == nil {
		return
	}
	for _, header := range legacyAliasHeaders {
		_ = registry.Register(header, handler)
	}
}

// decodeLegacyAlias validates the exact renderer shape and resolves its existing alias.
func decodeLegacyAlias(packet codec.Packet) (legacyAlias, error) {
	switch packet.Header {
	case inguilds.Header:
		return legacyAlias{code: "my_guild_bases_search"}, inguilds.Decode(packet)
	case inrights.Header:
		return legacyAlias{code: "my_room_rights_search"}, inrights.Decode(packet)
	case infrequent.Header:
		return legacyAlias{code: "my_frequent_room_history_search"}, infrequent.Decode(packet)
	case inofficial.Header:
		_, err := inofficial.Decode(packet)
		return legacyAlias{code: "official_view"}, err
	case infriends.Header:
		return legacyAlias{code: "rooms_where_my_friends_are"}, infriends.Decode(packet)
	case inhistory.Header:
		return legacyAlias{code: "my_room_history_search"}, inhistory.Decode(packet)
	case infriendsowned.Header:
		return legacyAlias{code: "my_friends_rooms_search"}, infriendsowned.Decode(packet)
	case inmyrooms.Header:
		return legacyAlias{code: "my_rooms"}, inmyrooms.Decode(packet)
	case inrecommended.Header:
		return legacyAlias{code: "my_recommended_rooms"}, inrecommended.Decode(packet)
	case infavourites.Header:
		return legacyAlias{code: "favorites"}, infavourites.Decode(packet)
	case inpopular.Header:
		request, err := inpopular.Decode(packet)
		return legacyAlias{code: "hotel_view", data: request.Query}, err
	case inguildsearch.Header:
		_, err := inguildsearch.Decode(packet)
		return legacyAlias{code: "guild_base_search"}, err
	case inhighest.Header:
		_, err := inhighest.Decode(packet)
		return legacyAlias{code: "rooms_with_highest_score_search"}, err
	case intext.Header:
		query, err := intext.Decode(packet)
		return legacyAlias{code: "room_text_search", data: query}, err
	default:
		return legacyAlias{}, codec.ErrUnexpectedHeader
	}
}
