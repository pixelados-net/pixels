package host

import (
	"testing"
	"time"

	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	"go.uber.org/zap"
)

// TestBackendBuildsEveryScopedCapability verifies host composition boundaries.
func TestBackendBuildsEveryScopedCapability(t *testing.T) {
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	events := pluginevent.NewHub(time.Second, zap.NewNop())
	commands := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	backend := NewBackend(
		players, bindings, connections, netconn.NewHandlerRegistry(), nil,
		pluginroutes.New(), pluginevent.NewHub(time.Second, zap.NewNop()), plugincommand.NewTree(":", time.Second, nil, zap.NewNop()), time.Second, zap.NewNop(),
	)
	host := backend.HostFor(pluginruntime.NewScope("demo"))
	if host.Players() == nil || host.Routes() == nil || host.Events() == nil || host.Commands() == nil || host.Permissions() == nil {
		t.Fatal("expected every scoped capability")
	}
	backend.events = events
	backend.commands = commands
	local := bus.New()
	chat := chatsend.New(chatconfig.Config{}, players, bindings, roomlive.NewRegistry(nil), connections, nil, nil, nil, nil, local, nil, chatsend.Nodes{})
	if err := backend.Connect(local, chat); err != nil {
		t.Fatalf("connect plugin bridges: %v", err)
	}
}
