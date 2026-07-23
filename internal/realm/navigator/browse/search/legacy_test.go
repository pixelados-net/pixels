package search

import (
	"context"
	"testing"

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

// TestLegacyAliasesReturnEmptyResults verifies every NOOP through a live session context.
func TestLegacyAliasesReturnEmptyResults(t *testing.T) {
	connection, sent := legacyConnection(t)
	requests := []struct {
		packet codec.Packet
		code   string
		data   string
	}{
		{legacyPacket(t, inguilds.Header, nil), "my_guild_bases_search", ""},
		{legacyPacket(t, inrights.Header, nil), "my_room_rights_search", ""},
		{legacyPacket(t, infrequent.Header, nil), "my_frequent_room_history_search", ""},
		{legacyPacket(t, inofficial.Header, inofficial.Definition, codec.Int32(0)), "official_view", ""},
		{legacyPacket(t, infriends.Header, nil), "rooms_where_my_friends_are", ""},
		{legacyPacket(t, inhistory.Header, nil), "my_room_history_search", ""},
		{legacyPacket(t, infriendsowned.Header, nil), "my_friends_rooms_search", ""},
		{legacyPacket(t, inmyrooms.Header, nil), "my_rooms", ""},
		{legacyPacket(t, inrecommended.Header, nil), "my_recommended_rooms", ""},
		{legacyPacket(t, infavourites.Header, nil), "favorites", ""},
		{legacyPacket(t, inpopular.Header, inpopular.Definition, codec.String("busy"), codec.Int32(0)), "hotel_view", "busy"},
		{legacyPacket(t, inguildsearch.Header, inguildsearch.Definition, codec.Int32(9)), "guild_base_search", ""},
		{legacyPacket(t, inhighest.Header, inhighest.Definition, codec.Int32(0)), "rooms_with_highest_score_search", ""},
		{legacyPacket(t, intext.Header, intext.Definition, codec.String("pixels")), "room_text_search", "pixels"},
	}
	handle := NewLegacyAliasHandler()
	for _, request := range requests {
		before := len(*sent)
		if err := handle(connection, request.packet); err != nil {
			t.Fatalf("header %d: %v", request.packet.Header, err)
		}
		if len(*sent) != before+1 || (*sent)[before].Header != outsearch.Header {
			t.Fatalf("header %d packets=%+v", request.packet.Header, (*sent)[before:])
		}
		values, err := codec.DecodePacketExact((*sent)[before], outsearch.Definition)
		if err != nil || values[0].String != request.code || values[1].String != request.data || values[2].Int32 != 0 {
			t.Fatalf("header %d values=%+v err=%v", request.packet.Header, values, err)
		}
	}
}

// TestRegisterLegacyAliases verifies nil safety, count, and strict dispatch.
func TestRegisterLegacyAliases(t *testing.T) {
	RegisterLegacyAliases(nil, nil)
	registry := netconn.NewHandlerRegistry()
	handle := NewLegacyAliasHandler()
	RegisterLegacyAliases(registry, handle)
	if registry.Len() != len(legacyAliasHeaders) {
		t.Fatalf("handlers=%d want=%d", registry.Len(), len(legacyAliasHeaders))
	}
	connection, _ := legacyConnection(t)
	if err := handle(connection, codec.Packet{Header: inofficial.Header}); err == nil {
		t.Fatal("malformed official request accepted")
	}
	if err := handle(connection, codec.Packet{Header: 65535}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("unexpected error %v", err)
	}
}

// legacyPacket builds one request against its renderer-derived shape.
func legacyPacket(t testing.TB, header uint16, definition codec.Definition, values ...codec.Value) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(header, definition, values...)
	if err != nil {
		t.Fatal(err)
	}
	return packet
}

// legacyConnection creates a session context that records outbound packets.
func legacyConnection(t testing.TB) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { connection = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, len(legacyAliasHeaders))
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "navigator-legacy", Kind: "test", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	return connection, &sent
}
