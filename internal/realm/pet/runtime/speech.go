package runtime

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	outtalk "github.com/niflaot/pixels/networking/outbound/chat/talk"
	"github.com/niflaot/pixels/pkg/i18n"
)

// FindNamed returns one active pet whose name prefixes a chat command.
func (service *Service) FindNamed(roomID int64, message string) (petrecord.Pet, string, bool) {
	message = strings.TrimSpace(message)
	for _, pet := range service.roomPets(roomID) {
		pet.mutex.Lock()
		value := pet.record
		pet.mutex.Unlock()
		if len(message) <= len(value.Name) || !strings.EqualFold(message[:len(value.Name)], value.Name) || message[len(value.Name)] != ' ' {
			continue
		}
		return value, strings.TrimSpace(message[len(value.Name)+1:]), true
	}
	return petrecord.Pet{}, "", false
}

// VocalizeByPet delivers one explicit localized vocal for a visible pet.
func (service *Service) VocalizeByPet(ctx context.Context, roomID int64, petID int64) error {
	active, found := service.rooms.Find(roomID)
	if !found {
		return petrecord.ErrInvalidState
	}
	pet, found := service.Snapshot(roomID, petID)
	if !found {
		return petrecord.ErrPetNotFound
	}
	_, _, err := service.vocalize(ctx, active, pet)
	return err
}

// vocalize selects and delivers one weighted localized species vocal.
func (service *Service) vocalize(ctx context.Context, active *roomlive.Room, pet petrecord.Pet) (time.Duration, bool, error) {
	if service.references == nil {
		return 0, false, petrecord.ErrInvalidState
	}
	references, err := service.references.Current(ctx)
	if err != nil || pet.TypeID < 0 || pet.TypeID >= int32(len(references.Vocals)) {
		return 0, false, err
	}
	if references.SpeciesPresent[pet.TypeID] && references.Species[pet.TypeID].Plant && pet.DerivePlantState(service.Now(), references.Species[pet.TypeID]).Dead {
		return 0, false, nil
	}
	vocals := references.Vocals[pet.TypeID]
	total := uint64(0)
	for _, vocal := range vocals {
		if vocal.Enabled && vocal.Weight > 0 {
			total += uint64(vocal.Weight)
		}
	}
	if total == 0 {
		return 0, false, nil
	}
	ticket := service.source.Uint64() % total
	selected := petrecord.Vocal{}
	for _, vocal := range vocals {
		if !vocal.Enabled || vocal.Weight <= 0 {
			continue
		}
		if ticket < uint64(vocal.Weight) {
			selected = vocal
			break
		}
		ticket -= uint64(vocal.Weight)
	}
	message := selected.TextKey
	if service.translations != nil {
		message = service.translations.Default(i18n.Key(selected.TextKey))
	}
	unit, found := active.Unit(EntityKey(pet.ID))
	if !found || message == "" {
		return selected.Cooldown, false, nil
	}
	if service.speech != nil {
		consumed, interceptErr := service.speech.InterceptPet(ctx, active.ID(), EntityKey(pet.ID), message)
		if interceptErr != nil || consumed {
			return selected.Cooldown, consumed, interceptErr
		}
	}
	packet, err := outtalk.Encode(int32(unit.UnitID), message, 0, 0, int32(utf8.RuneCountInString(message)))
	if err != nil {
		return selected.Cooldown, false, err
	}
	return selected.Cooldown, true, broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// SpeakLocalized delivers one exact localized pet response through room automation.
func (service *Service) SpeakLocalized(ctx context.Context, active *roomlive.Room, pet petrecord.Pet, key i18n.Key) error {
	message := string(key)
	if service.translations != nil {
		message = service.translations.Default(key)
	}
	unit, found := active.Unit(EntityKey(pet.ID))
	if !found || message == "" {
		return petrecord.ErrInvalidState
	}
	if service.speech != nil {
		consumed, err := service.speech.InterceptPet(ctx, active.ID(), EntityKey(pet.ID), message)
		if err != nil || consumed {
			return err
		}
	}
	packet, err := outtalk.Encode(int32(unit.UnitID), message, 0, 0, int32(utf8.RuneCountInString(message)))
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}
