package websocket

import (
	"context"
	"testing"

	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap"
)

// TestHandleWriteActivatesSecurity verifies queued activation handling.
func TestHandleWriteActivatesSecurity(t *testing.T) {
	socket := testQueuedSocket(t, 1)
	socket.log = zap.NewNop()

	if !socket.handleWrite(writeItem{kind: writeActivate, channel: testSecureChannel{state: netconn.SecurityReady}}) {
		t.Fatal("expected activation to keep writer alive")
	}
	if socket.session.SecurityState() != netconn.SecurityReady {
		t.Fatalf("expected ready security, got %d", socket.session.SecurityState())
	}
}

// TestHandleWriteIgnoresUnknownItems verifies defensive queue item handling.
func TestHandleWriteIgnoresUnknownItems(t *testing.T) {
	socket := testQueuedSocket(t, 1)
	if !socket.handleWrite(writeItem{}) {
		t.Fatal("expected unknown item to be ignored")
	}
}

// TestActivateSecurityRejectsInvalidChannel verifies activation failure cleanup.
func TestActivateSecurityRejectsInvalidChannel(t *testing.T) {
	socket := testQueuedSocket(t, 1)
	if socket.activateSecurity(nil) {
		t.Fatal("expected activation failure")
	}

	select {
	case <-socket.done:
	default:
		t.Fatal("expected failed activation to finish")
	}
}

// TestEnqueueCloseForLocalClose verifies close-only disposal enqueueing.
func TestEnqueueCloseForLocalClose(t *testing.T) {
	socket := testQueuedSocket(t, 2)
	socket.enqueueClose(context.Background(), netconn.Reason{Code: netconn.DisconnectLocalClose})

	if len(socket.queue) != 1 {
		t.Fatalf("expected one close item, got %d", len(socket.queue))
	}
}

// testSecureChannel is a deterministic secure channel fixture.
type testSecureChannel struct {
	// state stores the configured security phase.
	state netconn.SecurityState
}

// State returns the configured security state.
func (channel testSecureChannel) State() netconn.SecurityState {
	if channel.state == 0 {
		return netconn.SecurityReady
	}

	return channel.state
}

// Begin accepts security negotiation.
func (channel testSecureChannel) Begin(context.Context) error {
	return nil
}

// Open returns inbound bytes unchanged.
func (channel testSecureChannel) Open(src []byte) ([]byte, error) {
	return src, nil
}

// Seal returns outbound bytes unchanged.
func (channel testSecureChannel) Seal(src []byte) ([]byte, error) {
	return src, nil
}

// Close accepts security cleanup.
func (channel testSecureChannel) Close(netconn.Reason) error {
	return nil
}
