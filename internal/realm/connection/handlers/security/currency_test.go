package security

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/auth/sso"
	currencyrequest "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcredits "github.com/niflaot/pixels/networking/outbound/user/currency/credits"
	outwallet "github.com/niflaot/pixels/networking/outbound/user/currency/wallet"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestTicketBootstrapsComposedCurrencyHolder verifies authentication sends the durable wallet.
func TestTicketBootstrapsComposedCurrencyHolder(t *testing.T) {
	tickets := testSSO(t)
	ticket, err := tickets.Create(context.Background(), sso.CreateRequest{PlayerID: 2, TTL: time.Minute})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	currencies := &currencyrequest.Handler{
		Players: players, Bindings: bindings, Currencies: currencyReader{},
	}
	authenticator := NewAuthenticator(tickets, testFinder{}, players, bindings, bus.New(), currencies)
	session, sent := currencySession(t, authenticator)

	if err := session.Receive(context.Background(), ticketPacket(t, ticket.Value)); err != nil {
		t.Fatalf("receive ticket: %v", err)
	}
	if len(*sent) != 10 || (*sent)[8].Header != outcredits.Header || (*sent)[9].Header != outwallet.Header {
		t.Fatalf("unexpected bootstrap packets %#v", *sent)
	}
}

// currencySession creates a security session with currency bootstrap enabled.
func currencySession(t *testing.T, authenticator *Authenticator) (*netconn.Session, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	Register(inbound, authenticator)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil },
		netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())

	sent := make([]codec.Packet, 0, 10)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "currency-security", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := session.Transition(netconn.EventPacketReceived); err != nil {
		t.Fatalf("packet transition: %v", err)
	}

	return session, &sent
}

// currencyReader returns authentication wallet balances.
type currencyReader struct{}

// Wallet returns authentication wallet balances.
func (currencyReader) Wallet(_ context.Context, playerID int64) ([]currencymodel.Balance, error) {
	return []currencymodel.Balance{
		{PlayerID: playerID, CurrencyType: -1, Amount: 100},
		{PlayerID: playerID, CurrencyType: 5, Amount: 2},
	}, nil
}

// Balance returns one authentication balance.
func (currencyReader) Balance(context.Context, int64, int32) (int64, error) { return 0, nil }

// Types returns no authentication definitions.
func (currencyReader) Types(context.Context) ([]currencymodel.Definition, error) { return nil, nil }
