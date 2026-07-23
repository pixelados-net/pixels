package avatar

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	chatwhisper "github.com/niflaot/pixels/networking/outbound/chat/whisper"
)

// messageConnection captures private WIRED messages for one room player.
type messageConnection struct {
	netconn.Connection
	// id identifies the captured connection.
	id netconn.ID
	// packets stores delivered packets.
	packets []codec.Packet
}

// ID returns the configured connection identifier.
func (connection *messageConnection) ID() netconn.ID { return connection.id }

// Kind returns the test transport kind.
func (*messageConnection) Kind() netconn.Kind { return "test" }

// Send captures one outbound packet.
func (connection *messageConnection) Send(_ context.Context, packet codec.Packet) error {
	connection.packets = append(connection.packets, packet)
	return nil
}

// TestActorlessShowMessageDeliversToEveryPlayer verifies timer messages remain observable.
func TestActorlessShowMessageDeliversToEveryPlayer(t *testing.T) {
	rooms, active := avatarRoom(t)
	if _, err := rooms.Join(context.Background(), active.ID(), live.Occupant{PlayerID: 8, Username: "alice", ConnectionID: "other", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	demo := &messageConnection{id: "test"}
	alice := &messageConnection{id: "other"}
	if err := connections.Register(demo); err != nil {
		t.Fatal(err)
	}
	if err := connections.Register(alice); err != nil {
		t.Fatal(err)
	}
	service := New(rooms, nil, connections, nil, nil)
	node := &configuration.Node{Parameters: configuration.Parameters{Text: "timer %username%"}}
	result, err := service.ExecuteAvatar(context.Background(), effect.ShowMessage, node, trigger.Event{RoomID: active.ID()})
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	assertTimerMessage(t, demo.packets, "timer demo")
	assertTimerMessage(t, alice.packets, "timer alice")
}

// assertTimerMessage verifies one captured WIRED whisper payload.
func assertTimerMessage(t *testing.T, packets []codec.Packet, want string) {
	t.Helper()
	if len(packets) != 1 || packets[0].Header != chatwhisper.Header {
		t.Fatalf("packets=%#v", packets)
	}
	values, err := codec.DecodePacketExact(packets[0], chatwhisper.Definition)
	if err != nil || values[1].String != want {
		t.Fatalf("message=%q want=%q err=%v", values[1].String, want, err)
	}
}
