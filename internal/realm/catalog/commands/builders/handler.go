package builders

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	placecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/place"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	infloor "github.com/niflaot/pixels/networking/inbound/catalog/builders/floor"
	inwall "github.com/niflaot/pixels/networking/inbound/catalog/builders/wall"
)

var errSingleFurnitureRequired = errors.New("builders club offer must grant exactly one furniture item")

// Handler atomically purchases and places a Builders Club offer.
type Handler struct {
	// Config stores the normalized discontinued-tier policy.
	Config Config
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps connections to authenticated players.
	Bindings *binding.Registry
	// Runtime stores active rooms and placement counts.
	Runtime *roomlive.Registry
	// Catalog performs the shared purchase transaction.
	Catalog *catalogservice.Service
	// Place reuses authoritative room placement validation and projection.
	Place placecmd.Handler
}

// Register adds both Builders Club placement packet handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(infloor.Header, handler.floor)
	_ = registry.Register(inwall.Header, handler.wall)
}

// floor handles one catalog purchase with floor placement.
func (handler Handler) floor(connection netconn.Context, packet codec.Packet) error {
	payload, err := infloor.Decode(packet)
	if err != nil {
		return err
	}
	return handler.purchase(connection, int64(payload.PageID), int64(payload.OfferID), payload.ExtraData, placecmd.Command{Handler: connection, X: int(payload.X), Y: int(payload.Y), Rotation: int(payload.Direction)})
}

// wall handles one catalog purchase with wall placement.
func (handler Handler) wall(connection netconn.Context, packet codec.Packet) error {
	payload, err := inwall.Decode(packet)
	if err != nil {
		return err
	}
	return handler.purchase(connection, int64(payload.PageID), int64(payload.OfferID), payload.ExtraData, placecmd.Command{Handler: connection, WallPosition: payload.WallPosition})
}

// purchase enforces policy and combines catalog and placement persistence.
func (handler Handler) purchase(connection netconn.Context, pageID int64, offerID int64, extraData string, placement placecmd.Command) error {
	config := handler.Config.Normalize()
	if config.FurnitureLimit == 0 {
		return nil
	}
	player, err := catalogsession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found || len(active.FurnitureItems()) >= config.FurnitureLimit {
		return nil
	}
	ctx := context.Background()
	_, offers, err := handler.Catalog.Page(ctx, pageID, player.ID(), catalogsession.HasClub(player))
	if err != nil || !containsOffer(offers, offerID) {
		return err
	}
	var projectPlacement func(context.Context) error
	_, err = handler.Catalog.PurchaseAndMutate(ctx, catalogservice.PurchaseParams{PlayerID: player.ID(), CatalogItemID: offerID, HasClub: catalogsession.HasClub(player), Amount: 1, ExtraData: extraData}, func(txCtx context.Context, result catalogservice.PurchaseResult) error {
		if len(result.GrantedItems) != 1 {
			return errSingleFurnitureRequired
		}
		placement.ItemID = result.GrantedItems[0].ID
		deferredCtx, complete := placecmd.DeferProjections(txCtx)
		projectPlacement = complete
		return handler.Place.Handle(deferredCtx, command.Envelope[placecmd.Command]{Command: placement})
	})
	if err != nil || projectPlacement == nil {
		return err
	}
	return projectPlacement(ctx)
}

// containsOffer reports whether the requested page owns one offer.
func containsOffer(items []catalogmodel.Item, offerID int64) bool {
	for _, item := range items {
		if item.ID == offerID {
			return true
		}
	}
	return false
}
