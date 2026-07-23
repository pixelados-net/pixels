package decor

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	outinfo "github.com/niflaot/pixels/networking/outbound/room/entities/info"
)

// mannequinState stores map-style mannequin data durably as JSON.
type mannequinState struct {
	// Gender stores the outfit gender.
	Gender string `json:"gender"`
	// Figure stores clothing-only figure parts.
	Figure string `json:"figure"`
	// Name stores the visible outfit name.
	Name string `json:"name"`
}

// handleMannequin saves an outfit look or display name.
func (handler Handler) handleMannequin(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil || !allowed {
		return err
	}
	item, definition, state, found, err := handler.mannequin(ctx, roomID, command.ItemID)
	if err != nil || !found {
		return err
	}
	if command.Kind == KindMannequinLook {
		snapshot := player.Snapshot()
		state.Gender = string(snapshot.Gender)
		state.Figure = mannequinFigure(snapshot.Look)
	} else {
		name := strings.TrimSpace(command.Text)
		if name == "" || utf8.RuneCountInString(name) > 30 {
			return nil
		}
		if handler.GlobalFilter != nil {
			name, _ = handler.GlobalFilter.Censor(name)
		}
		if handler.WordFilters != nil {
			name, _, err = handler.WordFilters.Censor(ctx, roomID, name)
			if err != nil {
				return err
			}
		}
		state.Name = name
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}
	updated, err := handler.furnitureState(ctx, item, roomID, string(encoded))
	if err != nil {
		return err
	}
	active.SetFurnitureExtraData(updated.ID, updated.ExtraData)
	return handler.broadcastFloorUpdate(ctx, active, updated, definition)
}

// Use handles mannequin outfit application from the generic furniture-use packet.
func (handler Handler) Use(ctx context.Context, player *playerlive.Player, active *roomlive.Room, item worldfurniture.Item) (bool, error) {
	if item.Definition.InteractionType == "background_toner" {
		return true, handler.toggleToner(ctx, active, item)
	}
	if item.Definition.InteractionType != "mannequin" {
		return false, nil
	}
	_, _, state, found, err := handler.mannequin(ctx, active.ID(), item.ID)
	if err != nil || !found {
		return true, err
	}
	snapshot := player.Snapshot()
	if state.Gender == "" || state.Gender != string(snapshot.Gender) || state.Figure == "" {
		return true, nil
	}
	look := mergeMannequinFigure(snapshot.Look, state.Figure)
	record, err := handler.PlayerAdmin.Update(ctx, player.ID(), playerservice.UpdateParams{Look: &look})
	if err != nil {
		return true, err
	}
	if err = player.ReplaceSnapshot(playerlive.SnapshotFromRecord(record)); err != nil {
		return true, err
	}
	active.UpdateOccupantProfile(player.ID(), look, string(record.Profile.Gender), record.Profile.Motto)
	unit, found := active.Unit(player.ID())
	if !found {
		return true, nil
	}
	packet, err := outinfo.Encode(unit.UnitID, look, string(record.Profile.Gender), record.Profile.Motto, 0)
	if err != nil {
		return true, err
	}
	if err = handler.broadcast(ctx, active, packet); err != nil {
		return true, err
	}
	return true, nil
}

// mannequin resolves one placed mannequin and its durable state.
func (handler Handler) mannequin(ctx context.Context, roomID int64, itemID int64) (furnituremodel.Item, furnituremodel.Definition, mannequinState, bool, error) {
	item, found, err := handler.Furniture.FindItemByID(ctx, itemID)
	if err != nil || !found || item.RoomID == nil || *item.RoomID != roomID {
		return item, furnituremodel.Definition{}, mannequinState{}, false, err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || definition.InteractionType != "mannequin" {
		return item, definition, mannequinState{}, false, err
	}
	state := mannequinState{}
	if item.ExtraData != "" {
		_ = json.Unmarshal([]byte(item.ExtraData), &state)
	}
	return item, definition, state, true, nil
}

// mannequinFigure retains only clothing parts and adds Nitro's mannequin head.
func mannequinFigure(figure string) string {
	parts := []string{"hd-99999-99998"}
	for _, part := range strings.Split(figure, ".") {
		if clothingPart(part) {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ".")
}

// mergeMannequinFigure replaces clothing while retaining the actor's identity parts.
func mergeMannequinFigure(current string, outfit string) string {
	parts := make([]string, 0, 12)
	for _, part := range strings.Split(current, ".") {
		if !clothingPart(part) {
			parts = append(parts, part)
		}
	}
	for _, part := range strings.Split(outfit, ".") {
		if clothingPart(part) {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ".")
}

// clothingPart reports whether one figure segment is mannequin clothing.
func clothingPart(part string) bool {
	prefix, _, _ := strings.Cut(part, "-")
	return prefix == "ca" || prefix == "cc" || prefix == "ch" || prefix == "lg" || prefix == "sh" || prefix == "wa"
}
