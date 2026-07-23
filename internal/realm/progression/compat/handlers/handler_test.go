package handlers

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrandom "github.com/niflaot/pixels/networking/inbound/navigator/competition/random"
	inroom "github.com/niflaot/pixels/networking/inbound/navigator/competition/room"
	insearch "github.com/niflaot/pixels/networking/inbound/navigator/competition/search"
	insubmittable "github.com/niflaot/pixels/networking/inbound/navigator/competition/submittable"
	incompetitioninit "github.com/niflaot/pixels/networking/inbound/progression/competition/init"
	inpartof "github.com/niflaot/pixels/networking/inbound/progression/competition/partof"
	insubmit "github.com/niflaot/pixels/networking/inbound/progression/competition/submit"
	ingamelist "github.com/niflaot/pixels/networking/inbound/progression/game/list"
	ingameuser "github.com/niflaot/pixels/networking/inbound/progression/game/user"
	inresolutionopen "github.com/niflaot/pixels/networking/inbound/progression/resolution/open"
	inresolutionreset "github.com/niflaot/pixels/networking/inbound/progression/resolution/reset"
	outstatus "github.com/niflaot/pixels/networking/outbound/camera/competitionstatus"
	outentry "github.com/niflaot/pixels/networking/outbound/progression/competition/entry"
	outpartof "github.com/niflaot/pixels/networking/outbound/progression/competition/partof"
	outrooms "github.com/niflaot/pixels/networking/outbound/progression/competition/rooms"
	outgamelist "github.com/niflaot/pixels/networking/outbound/progression/game/list"
	outgameuser "github.com/niflaot/pixels/networking/outbound/progression/game/user"
	outresolutionlist "github.com/niflaot/pixels/networking/outbound/progression/resolution/list"
)

// TestCompatibilityHandlersReturnTruthfulEmptyState verifies every registered neutral adapter.
func TestCompatibilityHandlersReturnTruthfulEmptyState(t *testing.T) {
	connection, sent := compatibilityConnection(t)
	requests := []struct {
		packet codec.Packet
		want   uint16
	}{
		{compatibilityPacket(t, inresolutionopen.Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(7), codec.Int32(2)), outresolutionlist.Header},
		{compatibilityPacket(t, ingamelist.Header, nil), outgamelist.Header},
		{compatibilityPacket(t, ingameuser.Header, codec.Definition{codec.Int32Field}, codec.Int32(7)), outgameuser.Header},
		{compatibilityPacket(t, incompetitioninit.Header, nil), outstatus.Header},
		{compatibilityPacket(t, inpartof.Header, codec.Definition{codec.StringField}, codec.String("summer")), outpartof.Header},
		{compatibilityPacket(t, insubmit.Header, codec.Definition{codec.StringField, codec.Int32Field}, codec.String("summer"), codec.Int32(1)), outentry.Header},
		{compatibilityPacket(t, insearch.Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(4), codec.Int32(2)), outrooms.Header},
		{compatibilityPacket(t, inroom.Header, codec.Definition{codec.StringField, codec.Int32Field}, codec.String("summer"), codec.Int32(3)), outrooms.Header},
		{compatibilityPacket(t, inrandom.Header, codec.Definition{codec.StringField}, codec.String("summer")), outrooms.Header},
		{compatibilityPacket(t, insubmittable.Header, nil), outrooms.Header},
	}
	for _, request := range requests {
		before := len(*sent)
		if err := (Handler{}).handle(connection, request.packet); err != nil {
			t.Fatalf("header %d: %v", request.packet.Header, err)
		}
		if len(*sent) != before+1 || (*sent)[before].Header != request.want {
			t.Fatalf("header %d packets=%#v", request.packet.Header, *sent)
		}
	}
	reset := compatibilityPacket(t, inresolutionreset.Header, codec.Definition{codec.Int32Field}, codec.Int32(7))
	if err := (Handler{}).handle(connection, reset); err != nil {
		t.Fatal(err)
	}
}

// TestCompatibilityRegistrationAndValidation verifies safe registration and strict packet decoding.
func TestCompatibilityRegistrationAndValidation(t *testing.T) {
	Register(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	Register(registry, Handler{})
	if registry.Len() != 11 {
		t.Fatalf("handlers=%d want=11", registry.Len())
	}
	connection, _ := compatibilityConnection(t)
	if err := (Handler{}).handle(connection, codec.Packet{Header: inresolutionopen.Header}); err == nil {
		t.Fatal("malformed packet accepted")
	}
	if err := (Handler{}).handle(connection, codec.Packet{Header: 65535}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("unexpected header error %v", err)
	}
}

// compatibilityPacket builds one request using the declared field kinds.
func compatibilityPacket(t testing.TB, header uint16, definition codec.Definition, values ...codec.Value) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(header, definition, values...)
	if err != nil {
		t.Fatal(err)
	}
	return packet
}

// compatibilityConnection captures one connection context and its outbound packets.
func compatibilityConnection(t testing.TB) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { connection = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 10)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "progression-compat", Kind: "test", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	return connection, &sent
}
