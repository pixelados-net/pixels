package chat

import (
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
	messagecmd "github.com/niflaot/pixels/internal/realm/chat/commands/message"
	settingscmd "github.com/niflaot/pixels/internal/realm/chat/commands/settings"
	stylecmd "github.com/niflaot/pixels/internal/realm/chat/commands/style"
	typingcmd "github.com/niflaot/pixels/internal/realm/chat/commands/typing"
	settingshandler "github.com/niflaot/pixels/internal/realm/chat/handlers/settings"
	shouthandler "github.com/niflaot/pixels/internal/realm/chat/handlers/shout"
	stylehandler "github.com/niflaot/pixels/internal/realm/chat/handlers/style"
	talkhandler "github.com/niflaot/pixels/internal/realm/chat/handlers/talk"
	typinghandler "github.com/niflaot/pixels/internal/realm/chat/handlers/typing"
	whisperhandler "github.com/niflaot/pixels/internal/realm/chat/handlers/whisper"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	intypingstart "github.com/niflaot/pixels/networking/inbound/chat/typing/start"
	intypingstop "github.com/niflaot/pixels/networking/inbound/chat/typing/stop"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains chat packet handler dependencies.
type HandlerDeps struct {
	fx.In

	// Handlers stores connection realm registries.
	Handlers *realmconn.Handlers
	// Chat validates and delivers room communication.
	Chat *chatsend.Service
	// Bubbles validates style selections.
	Bubbles *bubble.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated sessions.
	Bindings *binding.Registry
	// Translations resolves expected user feedback.
	Translations i18n.Translator
	// Log records command dispatch.
	Log *zap.Logger
}

// RegisterConnectionHandlers registers every protocol-backed chat handler.
func RegisterConnectionHandlers(deps HandlerDeps) {
	if deps.Handlers == nil || deps.Handlers.Inbound == nil {
		return
	}
	message := messagecmd.Handler{Chat: deps.Chat}
	talkhandler.Register(deps.Handlers.Inbound, talkhandler.New(message, deps.Log))
	shouthandler.Register(deps.Handlers.Inbound, shouthandler.New(message, deps.Log))
	whisperhandler.Register(deps.Handlers.Inbound, whisperhandler.New(message, deps.Log))
	settingshandler.Register(deps.Handlers.Inbound, settingshandler.New(settingscmd.Handler{Players: deps.Players, Bindings: deps.Bindings}, deps.Log))
	stylehandler.Register(deps.Handlers.Inbound, stylehandler.New(stylecmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Bubbles: deps.Bubbles}, deps.Translations, deps.Log))
	typing := typingcmd.Handler{Chat: deps.Chat}
	_ = deps.Handlers.Inbound.Register(intypingstart.Header, typinghandler.New(typing, true, intypingstart.Decode, deps.Log))
	_ = deps.Handlers.Inbound.Register(intypingstop.Header, typinghandler.New(typing, false, intypingstop.Decode, deps.Log))
}
