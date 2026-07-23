package cfh

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	issuecreated "github.com/niflaot/pixels/internal/realm/moderation/events/issuecreated"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inreport "github.com/niflaot/pixels/networking/inbound/moderation/cfh/report"
	outresult "github.com/niflaot/pixels/networking/outbound/moderation/cfh/result"
	"github.com/niflaot/pixels/pkg/bus"
)

// reportCaptureStore captures the issue creation used by the packet workflow.
type reportCaptureStore struct {
	// Store supplies unused persistence methods.
	moderationrecord.Store
	// params stores the received report.
	params moderationrecord.ReportParams
}

// Topics returns the enabled queue topic used by the captured packet.
func (*reportCaptureStore) Topics(context.Context, bool) ([]moderationrecord.Topic, error) {
	return []moderationrecord.Topic{{ID: 3, Action: "queue", Enabled: true}}, nil
}

// CreateIssue captures one report and returns its durable projection.
func (store *reportCaptureStore) CreateIssue(_ context.Context, params moderationrecord.ReportParams) (moderationrecord.Issue, error) {
	store.params = params
	return moderationrecord.Issue{ID: 9, ReporterPlayerID: params.ReporterPlayerID, ReportedPlayerID: params.ReportedPlayerID, RoomID: params.RoomID, TopicID: params.TopicID, Kind: params.Kind, Message: params.Message, State: "open", Chatlog: params.Chatlog, CreatedAt: time.Now()}, nil
}

// TestCallForHelpAcceptsCapturedNitroPacket verifies persistence, room recovery, and success response.
func TestCallForHelpAcceptsCapturedNitroPacket(t *testing.T) {
	store := &reportCaptureStore{}
	events := bus.New()
	service := moderationcore.New(moderationconfig.Config{Enabled: true, ContextWindow: 10}, store, nil, nil, nil, nil, events)
	if err := service.RefreshTopics(context.Background()); err != nil {
		t.Fatal(err)
	}
	created := false
	if _, err := events.Subscribe(issuecreated.Name, bus.PriorityNormal, func(context.Context, bus.Event) error {
		created = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	players := reporterRegistry(t, 4, 2)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 4, ConnectionID: "reporter", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	runtime := &moderationruntime.Context{Moderation: service, Players: players, Bindings: bindings}
	inbound := netconn.NewHandlerRegistry()
	if err := inbound.Register(inreport.Header, (Handler{Context: runtime}).callForHelp(false), netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatal(err)
	}
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var sent codec.Packet
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "reporter", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error {
		sent = packet
		return nil
	}, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := base64.StdEncoding.DecodeString("ABNtZSBlc3RhIGhhcnJhc2VhbmRvAAAAAwAAAAMAAAAAAAEAAAADAAJYRA==")
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: inreport.Header, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	assertCapturedReport(t, store.params, created)
	assertReportResponse(t, sent)
}

// reporterRegistry creates one live reporter inside the supplied room.
func reporterRegistry(t *testing.T, playerID int64, roomID int64) *playerlive.Registry {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("reporter", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: playerID, Username: "Carol"}, peer)
	if err != nil {
		t.Fatal(err)
	}
	if err = player.EnterRoom(roomID); err != nil {
		t.Fatal(err)
	}
	players := playerlive.NewRegistry()
	if err = players.Add(player); err != nil {
		t.Fatal(err)
	}
	return players
}

// assertCapturedReport verifies the domain data recovered from Nitro's wire payload.
func assertCapturedReport(t *testing.T, params moderationrecord.ReportParams, created bool) {
	t.Helper()
	if !created || params.ReporterPlayerID != 4 || params.ReportedPlayerID == nil || *params.ReportedPlayerID != 3 || params.RoomID == nil || *params.RoomID != 2 || params.TopicID != 3 || params.Message != "me esta harraseando" {
		t.Fatalf("created=%t params=%+v", created, params)
	}
	if len(params.Chatlog) != 1 || params.Chatlog[0].PlayerID == nil || *params.Chatlog[0].PlayerID != 3 || params.Chatlog[0].PatternID != "ROOM" || params.Chatlog[0].Message != "XD" {
		t.Fatalf("chatlog=%+v", params.Chatlog)
	}
}

// assertReportResponse verifies Nitro receives the normal successful report result.
func assertReportResponse(t *testing.T, packet codec.Packet) {
	t.Helper()
	if packet.Header != outresult.Header {
		t.Fatalf("header=%d", packet.Header)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, packet.Payload)
	if err != nil || len(rest) != 0 || values[0].Int32 != 0 || values[1].String != "moderation.report.received" {
		t.Fatalf("values=%+v rest=%d err=%v", values, len(rest), err)
	}
}
