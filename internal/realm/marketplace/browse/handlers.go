// Package browse adapts Marketplace browsing packets.
package browse

import (
	"context"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inconfig "github.com/niflaot/pixels/networking/inbound/marketplace/config"
	insearch "github.com/niflaot/pixels/networking/inbound/marketplace/search"
	instats "github.com/niflaot/pixels/networking/inbound/marketplace/stats"
	outconfig "github.com/niflaot/pixels/networking/outbound/marketplace/config"
	outoffers "github.com/niflaot/pixels/networking/outbound/marketplace/offers"
	outstats "github.com/niflaot/pixels/networking/outbound/marketplace/stats"
)

// Handler owns Marketplace browsing adapters.
type Handler struct {
	// Service executes Marketplace browsing behavior.
	Service *marketcore.Service
}

// Register installs Marketplace browsing handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inconfig.Header, handler.config)
	_ = registry.Register(insearch.Header, handler.search)
	_ = registry.Register(instats.Header, handler.stats)
}

// config sends Marketplace policy.
func (handler Handler) config(connection netconn.Context, packet codec.Packet) error {
	if err := inconfig.Decode(packet); err != nil {
		return err
	}
	response, err := outconfig.Encode(handler.Service.Config())
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// search sends filtered offers.
func (handler Handler) search(connection netconn.Context, packet codec.Packet) error {
	payload, err := insearch.Decode(packet)
	if err != nil {
		return err
	}
	result, err := handler.Service.Search(context.Background(), marketcore.SearchParams{MinimumPrice: int64(payload.MinimumPrice), MaximumPrice: int64(payload.MaximumPrice), Query: payload.Query, SortType: payload.SortType})
	if err != nil {
		return err
	}
	response, err := outoffers.Encode(result.Offers, result.Total)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// stats sends one furniture definition's recent sales.
func (handler Handler) stats(connection netconn.Context, packet codec.Packet) error {
	payload, err := instats.Decode(packet)
	if err != nil {
		return err
	}
	definitions, found, err := handler.Service.DefinitionIDBySprite(context.Background(), payload.SpriteID)
	if err != nil {
		return err
	}
	if !found {
		return binding.ErrBindingNotFound
	}
	value, err := handler.Service.ItemStats(context.Background(), definitions)
	if err != nil {
		return err
	}
	response, err := outstats.Encode(value, payload.Category, payload.SpriteID)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}
