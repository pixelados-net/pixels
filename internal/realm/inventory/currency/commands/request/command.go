// Package request sends one player's current currency wallet.
package request

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/command"
	currencyholder "github.com/niflaot/pixels/internal/realm/inventory/currency/holder"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcredits "github.com/niflaot/pixels/networking/outbound/user/currency/credits"
	outwallet "github.com/niflaot/pixels/networking/outbound/user/currency/wallet"
)

const (
	// Name identifies the currency wallet request command.
	Name command.Name = "inventory.currency.request"

	// creditsType identifies the protocol credits sentinel.
	creditsType int32 = -1
)

// Command requests the authenticated player's wallet.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
}

// Handler handles currency wallet request commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry

	// Bindings maps authenticated connections to players.
	Bindings *binding.Registry

	// Currencies reads durable currency balances.
	Currencies currencyservice.Reader
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle sends the requesting player's current wallet.
func (handler *Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := handler.player(envelope.Command.Connection)
	if err != nil {
		return err
	}

	return handler.Send(ctx, envelope.Command.Connection, player.Currencies())
}

// Send sends a currency holder's current wallet to one connection.
func (handler *Handler) Send(ctx context.Context, connection netconn.Context, holder *currencyholder.Holder) error {
	balances, err := holder.Wallet(ctx, handler.Currencies)
	if err != nil {
		return fmt.Errorf("load player %d currency wallet: %w", holder.PlayerID(), err)
	}

	creditsAmount := int64(0)
	seasonal := make([]outwallet.Entry, 0, len(balances))
	for _, balance := range balances {
		if balance.CurrencyType == creditsType {
			creditsAmount = balance.Amount
			continue
		}
		seasonal = append(seasonal, outwallet.Entry{Type: balance.CurrencyType, Amount: balance.Amount})
	}

	creditsPacket, err := outcredits.Encode(creditsAmount)
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, creditsPacket); err != nil {
		return err
	}

	walletPacket, err := outwallet.Encode(seasonal)
	if err != nil {
		return err
	}

	return connection.Send(ctx, walletPacket)
}

// player resolves the authenticated live player.
func (handler *Handler) player(connection netconn.Context) (*playerlive.Player, error) {
	key := binding.ConnectionKey{ID: connection.ConnectionID, Kind: connection.ConnectionKind}
	sessionBinding, found := handler.Bindings.FindByConnection(key)
	if !found {
		return nil, binding.ErrBindingNotFound
	}
	player, found := handler.Players.Find(sessionBinding.PlayerID)
	if !found {
		return nil, ErrPlayerNotFound
	}

	return player, nil
}
