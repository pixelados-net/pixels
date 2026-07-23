package compatibility

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inconcurrent "github.com/niflaot/pixels/networking/inbound/economy/community/concurrentprogress"
	inreward "github.com/niflaot/pixels/networking/inbound/economy/community/concurrentreward"
	inearned "github.com/niflaot/pixels/networking/inbound/economy/community/earnedprizes"
	inhall "github.com/niflaot/pixels/networking/inbound/economy/community/halloffame"
	inprogress "github.com/niflaot/pixels/networking/inbound/economy/community/progress"
	inredeem "github.com/niflaot/pixels/networking/inbound/economy/community/redeemprize"
	invote "github.com/niflaot/pixels/networking/inbound/economy/community/vote"
	inaliases "github.com/niflaot/pixels/networking/inbound/furniture/aliases"
	insongs "github.com/niflaot/pixels/networking/inbound/furniture/songinfo"
	ingetinterstitial "github.com/niflaot/pixels/networking/inbound/notification/legacy/getinterstitial"
	interstitialshown "github.com/niflaot/pixels/networking/inbound/notification/legacy/interstitialshown"
	infaqtext "github.com/niflaot/pixels/networking/inbound/other/faq/gettext"
	infaqsearch "github.com/niflaot/pixels/networking/inbound/other/faq/search"
	inphonenumber "github.com/niflaot/pixels/networking/inbound/other/phone/trynumber"
	inphonecode "github.com/niflaot/pixels/networking/inbound/other/phone/verifycode"
	inpromos "github.com/niflaot/pixels/networking/inbound/session/promoarticles"
	outconcurrent "github.com/niflaot/pixels/networking/outbound/economy/community/concurrentprogress"
	outearned "github.com/niflaot/pixels/networking/outbound/economy/community/earnedprizes"
	outhall "github.com/niflaot/pixels/networking/outbound/economy/community/halloffame"
	outprogress "github.com/niflaot/pixels/networking/outbound/economy/community/progress"
	outvote "github.com/niflaot/pixels/networking/outbound/economy/community/voteevent"
	outaliases "github.com/niflaot/pixels/networking/outbound/furniture/aliases"
	outsongs "github.com/niflaot/pixels/networking/outbound/furniture/songinfo"
	outpromos "github.com/niflaot/pixels/networking/outbound/session/promoarticles"
)

// TestRegisterInstallsOptionalSystemHandlers verifies every compatibility route.
func TestRegisterInstallsOptionalSystemHandlers(t *testing.T) {
	Register(nil)
	registry := netconn.NewHandlerRegistry()
	Register(registry)
	if registry.Len() != 19 {
		t.Fatalf("unexpected handler count: %d", registry.Len())
	}
}

// TestRetiredSupportRequestsAreStrictNoops verifies abandoned UI requests do not emit responses.
func TestRetiredSupportRequestsAreStrictNoops(t *testing.T) {
	connection, sent := compatibilityConnection(t)
	requests := []struct {
		handle netconn.Handler
		packet codec.Packet
	}{
		{handle: retiredLanding, packet: codec.Packet{Header: interstitialshown.Header}},
		{handle: retiredLanding, packet: codec.Packet{Header: ingetinterstitial.Header}},
		{handle: retiredSupport, packet: codec.Packet{Header: inphonenumber.Header}},
		{handle: retiredSupport, packet: codec.Packet{Header: inphonecode.Header}},
		{handle: retiredSupport, packet: codec.Packet{Header: infaqtext.Header}},
		{handle: retiredSupport, packet: codec.Packet{Header: infaqsearch.Header}},
	}
	for _, request := range requests {
		if err := request.handle(connection, request.packet); err != nil {
			t.Fatalf("header %d: %v", request.packet.Header, err)
		}
	}
	if len(*sent) != 0 {
		t.Fatalf("unexpected packets: %+v", *sent)
	}
}

// TestHandlersRejectMalformedRequests verifies compatibility routes still validate wire shapes.
func TestHandlersRejectMalformedRequests(t *testing.T) {
	connection, _ := compatibilityConnection(t)
	requests := []netconn.Handler{promoArticles, furnitureAliases, communityHallOfFame, songInfo}
	for _, handle := range requests {
		if err := handle(connection, codec.Packet{}); err == nil {
			t.Fatal("expected malformed compatibility request rejection")
		}
	}
}

// TestCommunityGoalsRejectMalformedRequests verifies all compatibility decoders remain strict.
func TestCommunityGoalsRejectMalformedRequests(t *testing.T) {
	connection, _ := compatibilityConnection(t)
	requests := []codec.Packet{
		{Header: inredeem.Header},
		{Header: inprogress.Header, Payload: []byte{0}},
		{Header: inconcurrent.Header, Payload: []byte{0}},
		{Header: inearned.Header, Payload: []byte{0}},
		{Header: invote.Header},
		{Header: inreward.Header, Payload: []byte{0}},
	}
	for _, packet := range requests {
		if err := communityGoals(connection, packet); err == nil {
			t.Fatalf("expected request %d rejection", packet.Header)
		}
	}
}

// TestHandlersReturnExplicitEmptySnapshots verifies every optional-system response.
func TestHandlersReturnExplicitEmptySnapshots(t *testing.T) {
	connection, sent := compatibilityConnection(t)
	promo := codec.Packet{Header: inpromos.Header}
	aliases := codec.Packet{Header: inaliases.Header}
	hall, _ := codec.NewPacket(inhall.Header, inhall.Definition, codec.String("habboFameComp"))
	songs, _ := codec.NewPacket(insongs.Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(-1))
	requests := []struct {
		handle   netconn.Handler
		packet   codec.Packet
		expected uint16
	}{
		{handle: promoArticles, packet: promo, expected: outpromos.Header},
		{handle: furnitureAliases, packet: aliases, expected: outaliases.Header},
		{handle: communityHallOfFame, packet: hall, expected: outhall.Header},
		{handle: songInfo, packet: songs, expected: outsongs.Header},
	}
	for _, request := range requests {
		if err := request.handle(connection, request.packet); err != nil {
			t.Fatalf("handle %d: %v", request.packet.Header, err)
		}
		if (*sent)[len(*sent)-1].Header != request.expected {
			t.Fatalf("request %d sent header %d", request.packet.Header, (*sent)[len(*sent)-1].Header)
		}
	}
}

// TestCommunityGoalsReturnNeutralSnapshots verifies wire-only responses without domain state.
func TestCommunityGoalsReturnNeutralSnapshots(t *testing.T) {
	connection, sent := compatibilityConnection(t)
	redeem, _ := codec.NewPacket(inredeem.Header, inredeem.Definition, codec.Int32(7))
	vote, _ := codec.NewPacket(invote.Header, invote.Definition, codec.Int32(2))
	requests := []struct {
		packet   codec.Packet
		expected uint16
	}{
		{packet: redeem, expected: outearned.Header},
		{packet: codec.Packet{Header: inprogress.Header}, expected: outprogress.Header},
		{packet: codec.Packet{Header: inconcurrent.Header}, expected: outconcurrent.Header},
		{packet: codec.Packet{Header: inearned.Header}, expected: outearned.Header},
		{packet: vote, expected: outvote.Header},
		{packet: codec.Packet{Header: inreward.Header}, expected: outconcurrent.Header},
	}
	for _, request := range requests {
		before := len(*sent)
		if err := communityGoals(connection, request.packet); err != nil {
			t.Fatalf("handle %d: %v", request.packet.Header, err)
		}
		if len(*sent) != before+1 || (*sent)[before].Header != request.expected {
			t.Fatalf("request %d sent packets=%+v", request.packet.Header, (*sent)[before:])
		}
	}
}

// compatibilityConnection creates a handler context with observable sends.
func compatibilityConnection(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { connection = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 4)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "compatibility", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}
	return connection, &sent
}
