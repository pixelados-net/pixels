package websocket

import (
	"testing"
	"time"

	fastws "github.com/fasthttp/websocket"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestLoadConfig verifies WebSocket environment loading.
func TestLoadConfig(t *testing.T) {
	t.Setenv("PIXELS_WS_QUEUE_SIZE", "12")
	t.Setenv("PIXELS_WS_WRITE_TIMEOUT", "3s")
	t.Setenv("PIXELS_WS_READ_TIMEOUT", "4s")
	t.Setenv("PIXELS_WS_PING_INTERVAL", "5s")
	t.Setenv("PIXELS_WS_PONG_TIMEOUT", "6s")
	t.Setenv("PIXELS_WS_CLOSE_GRACE", "7s")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.QueueSize != 12 || config.WriteTimeout != 3*time.Second || config.CloseGrace != 7*time.Second {
		t.Fatalf("unexpected config: %#v", config)
	}
}

// TestConfigNormalize verifies defensive WebSocket defaults.
func TestConfigNormalize(t *testing.T) {
	config := Config{}.Normalize()
	if config.QueueSize == 0 || config.WriteTimeout == 0 || config.PingInterval == 0 {
		t.Fatalf("expected normalized config, got %#v", config)
	}
}

// TestCloseMappings verifies protocol and WebSocket close mappings.
func TestCloseMappings(t *testing.T) {
	if code, ok := pixelErrorCode(netconn.DisconnectProtocolError); !ok || code != 0 {
		t.Fatalf("expected protocol error code, got %d %v", code, ok)
	}

	if _, ok := errorPacket(netconn.Reason{Code: netconn.DisconnectKicked}); !ok {
		t.Fatal("expected kicked error packet")
	}

	if websocketCloseCode(netconn.DisconnectAuthenticationFailed) == 0 {
		t.Fatal("expected websocket close code")
	}

	if code, ok := pixelErrorCode(netconn.DisconnectServerShutdown); !ok || code != 0 {
		t.Fatalf("expected shutdown code, got %d %v", code, ok)
	}

	if _, ok := pixelErrorCode(netconn.DisconnectRemoteClose); ok {
		t.Fatal("expected no remote-close error code")
	}
	if code, ok := pixelDisconnectCode(netconn.DisconnectBanned); !ok || code != 1 {
		t.Fatalf("expected banned disconnect code, got %d %v", code, ok)
	}
	if _, ok := disconnectPacket(netconn.Reason{Code: netconn.DisconnectTransportError}); ok {
		t.Fatal("expected no transport-error disconnect packet")
	}

	if websocketCloseCode(netconn.DisconnectIdleTimeout) != fastws.CloseGoingAway {
		t.Fatal("expected idle timeout close code")
	}
	if readReason(&fastws.CloseError{Code: fastws.CloseNormalClosure}).Code != netconn.DisconnectRemoteClose {
		t.Fatal("expected remote close reason")
	}
	if readReason(assertError{}).Code != netconn.DisconnectTransportError {
		t.Fatal("expected transport error reason")
	}
}

// assertError is a deterministic error fixture.
type assertError struct{}

// Error returns the fixture message.
func (assertError) Error() string {
	return "assert"
}
