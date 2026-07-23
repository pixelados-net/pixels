package send

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	outtalk "github.com/niflaot/pixels/networking/outbound/chat/talk"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// commandExecutorForTest records and optionally consumes plugin commands.
type commandExecutorForTest struct {
	// handled controls whether the message leaves normal chat.
	handled bool
	// message stores the observed sanitized text.
	message string
}

// Execute records and returns the configured command decision.
func (executor *commandExecutorForTest) Execute(_ context.Context, _ sdkplayer.Player, message string) (bool, error) {
	executor.message = message
	return executor.handled, nil
}

// eventDispatcherForTest replaces or cancels one chat message.
type eventDispatcherForTest struct {
	// replacement stores the event-mutated text.
	replacement string
	// cancelled controls final native delivery.
	cancelled bool
}

// DispatchChat returns the configured event outcome.
func (dispatcher eventDispatcherForTest) DispatchChat(context.Context, sdkplayer.Player, int64, string) (string, bool) {
	return dispatcher.replacement, dispatcher.cancelled
}

// TestPluginCommandIsConsumedBeforeChat verifies command text never broadcasts.
func TestPluginCommandIsConsumedBeforeChat(t *testing.T) {
	fixture := newFixture(t)
	executor := &commandExecutorForTest{handled: true}
	fixture.service.SetPluginRuntime(executor, nil)
	fixture.request.Message = ":hello"

	if err := fixture.source.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive command: %v", err)
	}
	if executor.message != ":hello" || len(*fixture.sourcePackets) != 0 || len(*fixture.targetPackets) != 0 {
		t.Fatalf("expected consumed command, message=%q source=%d target=%d", executor.message, len(*fixture.sourcePackets), len(*fixture.targetPackets))
	}
}

// TestPluginChatEventCancelsDelivery verifies a plugin veto reaches the native action.
func TestPluginChatEventCancelsDelivery(t *testing.T) {
	fixture := newFixture(t)
	fixture.service.SetPluginRuntime(nil, eventDispatcherForTest{replacement: "blocked", cancelled: true})

	if err := fixture.source.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive cancelled chat: %v", err)
	}
	if len(*fixture.sourcePackets) != 0 || len(*fixture.targetPackets) != 0 {
		t.Fatalf("expected cancelled delivery, source=%d target=%d", len(*fixture.sourcePackets), len(*fixture.targetPackets))
	}
}

// TestPluginChatEventRewritesDelivery verifies event text reaches recipients.
func TestPluginChatEventRewritesDelivery(t *testing.T) {
	fixture := newFixture(t)
	fixture.service.SetPluginRuntime(nil, eventDispatcherForTest{replacement: "rewritten"})

	if err := fixture.source.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive rewritten chat: %v", err)
	}
	values, err := codec.DecodePacketExact((*fixture.targetPackets)[0], outtalk.Definition)
	if err != nil || values[1].String != "rewritten" {
		t.Fatalf("expected rewritten message, values=%#v err=%v", values, err)
	}
}
