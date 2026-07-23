package wiring

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/wired"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
)

// TestModuleConstructorsExposeFocusedBoundaries verifies dependency wiring preserves each domain contract.
func TestModuleConstructorsExposeFocusedBoundaries(t *testing.T) {
	rooms := roomlive.NewRegistry(nil)
	registered, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	compiler := NewCompiler(registered, roomwired.Config{})
	games := NewGames(rooms, nil)
	avatar := NewAvatarEffects(rooms, nil, nil, nil, nil)
	furniture := NewFurnitureEffects(rooms, nil, nil)
	bot := NewBotEffects(rooms, nil)
	effects := NewEffects(furniture, avatar, bot, games, nil)
	engine := NewEngine(roomwired.Config{}, nil, compiler, effects, nil, nil, furniture)
	if compiler == nil || games == nil || avatar == nil || furniture == nil || bot == nil || effects == nil || engine == nil {
		t.Fatal("module constructor returned nil")
	}
	repository := NewRepository(nil)
	if repository == nil || NewStore(repository) == nil || NewRewardStore(repository) == nil || NewHighscoreStore(repository) == nil {
		t.Fatal("repository boundary constructor returned nil")
	}
	handler := NewCommandHandler(roomwired.Config{}, nil, nil, rooms, nil, registered, compiler, engine, nil, nil)
	RegisterHandlers(nil, handler, nil)
	RegisterHandlers(nil, wiredcmd.Handler{}, nil)
}
