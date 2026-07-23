package request

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcredits "github.com/niflaot/pixels/networking/outbound/user/currency/credits"
	outwallet "github.com/niflaot/pixels/networking/outbound/user/currency/wallet"
)

// TestCommandNameReturnsStableName verifies command routing identity.
func TestCommandNameReturnsStableName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %q", (Command{}).CommandName())
	}
}

// TestHandleSendsCreditsAndSeasonalWallet verifies complete wallet projection.
func TestHandleSendsCreditsAndSeasonalWallet(t *testing.T) {
	connection, sent := currencyConnection(t)
	handler := &Handler{
		Players:  currencyPlayers(t),
		Bindings: currencyBindings(t),
		Currencies: fakeReader{wallet: []currencymodel.Balance{
			{PlayerID: 7, CurrencyType: -1, Amount: 100},
			{PlayerID: 7, CurrencyType: 0, Amount: 10},
			{PlayerID: 7, CurrencyType: 5, Amount: 2},
		}},
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection}})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outcredits.Header || (*sent)[1].Header != outwallet.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
}

// TestHandleSendsZeroCreditsForEmptyWallet verifies required credits bootstrap.
func TestHandleSendsZeroCreditsForEmptyWallet(t *testing.T) {
	connection, sent := currencyConnection(t)
	handler := &Handler{
		Players: currencyPlayers(t), Bindings: currencyBindings(t), Currencies: fakeReader{},
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Connection: connection}})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	values, err := codec.DecodePacketExact((*sent)[0], outcredits.Definition)
	if err != nil {
		t.Fatalf("decode credits: %v", err)
	}
	if values[0].String != "0.0" {
		t.Fatalf("unexpected credits %q", values[0].String)
	}
}

// TestHandleRejectsMissingLiveSessionState verifies capability ownership resolution.
func TestHandleRejectsMissingLiveSessionState(t *testing.T) {
	handler := &Handler{Players: playerlive.NewRegistry(), Bindings: binding.NewRegistry()}
	_, err := handler.player(netconn.Context{ConnectionID: "missing", ConnectionKind: "websocket"})
	if !errors.Is(err, binding.ErrBindingNotFound) {
		t.Fatalf("expected binding error, got %v", err)
	}

	bindings := currencyBindings(t)
	handler.Bindings = bindings
	_, err = handler.player(netconn.Context{ConnectionID: "currency-connection", ConnectionKind: "websocket"})
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf("expected player error, got %v", err)
	}
}

// currencyPlayers creates one composed live player.
func currencyPlayers(t *testing.T) *playerlive.Registry {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("currency-connection", "websocket", time.Now())
	if err != nil {
		t.Fatalf("new peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("new player: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return players
}

// currencyBindings creates one authenticated connection binding.
func currencyBindings(t *testing.T) *binding.Registry {
	t.Helper()
	bindings := binding.NewRegistry()
	err := bindings.Add(binding.Binding{
		PlayerID: 7, ConnectionID: "currency-connection", ConnectionKind: "websocket",
	})
	if err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return bindings
}

// currencyConnection creates a connection context that records outbound packets.
func currencyConnection(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil },
		netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())

	var captured netconn.Context
	if err := inbound.Register(1, func(context netconn.Context, _ codec.Packet) error {
		captured = context
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register capture: %v", err)
	}

	sent := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "currency-connection", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}

	return captured, &sent
}

// fakeReader returns command test balances.
type fakeReader struct {
	// wallet stores fake balances.
	wallet []currencymodel.Balance
}

// Wallet returns fake balances.
func (reader fakeReader) Wallet(context.Context, int64) ([]currencymodel.Balance, error) {
	return reader.wallet, nil
}

// Balance returns a fake balance.
func (fakeReader) Balance(context.Context, int64, int32) (int64, error) { return 0, nil }

// Types returns no fake definitions.
func (fakeReader) Types(context.Context) ([]currencymodel.Definition, error) { return nil, nil }
