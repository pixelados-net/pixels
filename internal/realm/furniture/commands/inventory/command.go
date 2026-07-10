// Package inventory lists a player's furniture inventory.
package inventory

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/furniture/list"
)

const (
	// Name identifies the furniture inventory list command.
	Name command.Name = "furniture.inventory.list"

	// fragmentSize is the maximum number of items sent per inventory fragment.
	fragmentSize = 1000
)

// Command requests a player's furniture inventory.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles furniture inventory list commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Furniture reads placed and inventory furniture records.
	Furniture furnitureservice.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a furniture inventory list command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := furnituresession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	player.OpenInventory()

	items, err := handler.Furniture.ListInventory(ctx, player.ID())
	if err != nil {
		return err
	}

	definitions, err := definitionsByID(ctx, handler.Furniture)
	if err != nil {
		return err
	}

	return handler.sendFragments(ctx, envelope.Command.Handler, items, definitions)
}

// definitionsByID indexes furniture definitions by id.
func definitionsByID(ctx context.Context, manager furnitureservice.DefinitionFinder) (map[int64]furnituremodel.Definition, error) {
	definitions, err := manager.ListDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	indexed := make(map[int64]furnituremodel.Definition, len(definitions))
	for _, definition := range definitions {
		indexed[definition.ID] = definition
	}

	return indexed, nil
}

// sendFragments sends inventory items as one or more fragments, matching Arcturus's 1000-item pages.
func (handler Handler) sendFragments(ctx context.Context, connection netconn.Context, items []furnituremodel.Item, definitions map[int64]furnituremodel.Definition) error {
	if len(items) == 0 {
		packet, err := outlist.Encode(1, 1, nil)
		if err != nil {
			return err
		}

		return connection.Send(ctx, packet)
	}

	totalFragments := (len(items) + fragmentSize - 1) / fragmentSize
	for fragment := 0; fragment < totalFragments; fragment++ {
		packet, err := outlist.Encode(fragment+1, totalFragments, fragmentRecords(items, definitions, fragment))
		if err != nil {
			return err
		}
		if err := connection.Send(ctx, packet); err != nil {
			return err
		}
	}

	return nil
}

// fragmentRecords maps one fragment slice of items into inventory records.
func fragmentRecords(items []furnituremodel.Item, definitions map[int64]furnituremodel.Definition, fragment int) []outlist.Item {
	start := fragment * fragmentSize
	end := start + fragmentSize
	if end > len(items) {
		end = len(items)
	}

	records := make([]outlist.Item, 0, end-start)
	for _, item := range items[start:end] {
		definition, ok := definitions[item.DefinitionID]
		if !ok {
			continue
		}
		records = append(records, outlist.Item{
			ID:                  item.ID,
			SpriteID:            definition.SpriteID,
			Kind:                inventoryKind(definition.Kind),
			Category:            inventoryCategory(definition.Name),
			ExtraData:           item.ExtraData,
			AllowInventoryStack: definition.AllowInventoryStack,
		})
	}

	return records
}

// inventoryCategory maps room-effect definitions to Nitro inventory categories.
func inventoryCategory(name string) outlist.Category {
	switch name {
	case "wallpaper":
		return outlist.CategoryWallpaper
	case "floor":
		return outlist.CategoryFloor
	case "landscape":
		return outlist.CategoryLandscape
	default:
		return outlist.CategoryDefault
	}
}

// inventoryKind maps persistent furniture kinds to inventory packet kinds.
func inventoryKind(kind furnituremodel.Kind) outlist.Kind {
	if kind == furnituremodel.KindWall {
		return outlist.KindWall
	}

	return outlist.KindFloor
}
