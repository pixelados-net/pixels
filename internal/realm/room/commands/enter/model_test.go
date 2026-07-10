package enter

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentrytile "github.com/niflaot/pixels/networking/outbound/room/entrytile"
	outmodel "github.com/niflaot/pixels/networking/outbound/room/model"
	outmodelname "github.com/niflaot/pixels/networking/outbound/room/modelname"
)

// TestSendModelSendsEntryTileBeforeScaledHeightmap verifies room model bootstrapping order.
func TestSendModelSendsEntryTileBeforeScaledHeightmap(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)

	err := SendModel(context.Background(), connection, roomForTest(), layoutForTest())
	if err != nil {
		t.Fatalf("send model: %v", err)
	}
	if len(*sent) != 3 {
		t.Fatalf("expected model packets, got %#v", *sent)
	}
	if (*sent)[0].Header != outmodelname.Header || (*sent)[1].Header != outentrytile.Header || (*sent)[2].Header != outmodel.Header {
		t.Fatalf("unexpected model packet order %#v", *sent)
	}

	entryValues, err := codec.DecodePacketExact((*sent)[1], outentrytile.Definition)
	if err != nil {
		t.Fatalf("decode entry tile packet: %v", err)
	}
	if entryValues[0].Int32 != 0 || entryValues[1].Int32 != 0 || entryValues[2].String != "0.0" || entryValues[3].Int32 != 2 {
		t.Fatalf("unexpected entry tile values %#v", entryValues)
	}

	values, err := codec.DecodePacketExact((*sent)[2], outmodel.Definition)
	if err != nil {
		t.Fatalf("decode model packet: %v", err)
	}
	if !values[0].Boolean {
		t.Fatalf("expected scaled heightmap packet, got %#v", values)
	}
}

// TestSendGeometryOmitsModelName verifies model requests cannot retrigger Nitro's request loop.
func TestSendGeometryOmitsModelName(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	if err := SendGeometry(context.Background(), connection, layoutForTest()); err != nil {
		t.Fatalf("send geometry: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outentrytile.Header || (*sent)[1].Header != outmodel.Header {
		t.Fatalf("expected entry tile and heightmap only, got %#v", *sent)
	}
}

// TestSendModelReturnsTransportError verifies model packet send failures.
func TestSendModelReturnsTransportError(t *testing.T) {
	connection, sendErr := failingModelConnectionForTest(t)

	err := SendModel(context.Background(), connection, roomForTest(), layoutForTest())
	if !errors.Is(err, sendErr) {
		t.Fatalf("expected send error, got %v", err)
	}
}

// failingModelConnectionForTest creates a connection that fails the second send.
func failingModelConnectionForTest(t *testing.T) (netconn.Context, error) {
	t.Helper()

	sendErr := errors.New("send failed")
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	inbound := netconn.NewHandlerRegistry()
	var captured netconn.Context
	if err := inbound.Register(1, func(context netconn.Context, _ codec.Packet) error {
		captured = context
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register inbound: %v", err)
	}

	sendCount := 0
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       netconn.ID("conn"),
		Kind:     netconn.Kind("websocket"),
		Inbound:  inbound,
		Outbound: outbound,
		Sender: func(context.Context, codec.Packet) error {
			sendCount++
			if sendCount == 2 {
				return sendErr
			}
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive context packet: %v", err)
	}

	return captured, sendErr
}
