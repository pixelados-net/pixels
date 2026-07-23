// Package command implements the shared dynamic-plugin chat command tree.
package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkcommand "github.com/niflaot/pixels/sdk/command"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.minekube.com/brigodier"
	"go.uber.org/zap"
)

var (
	// ErrInvalidCommand reports an invalid root command declaration.
	ErrInvalidCommand = errors.New("invalid plugin command")
	// ErrCommandExists reports a root literal already owned by another plugin.
	ErrCommandExists = errors.New("plugin command exists")
)

// Tree owns the shared Brigadier dispatcher and root ownership.
type Tree struct {
	// mutex protects startup command registration.
	mutex sync.RWMutex
	// dispatcher parses and executes the shared command tree.
	dispatcher brigodier.Dispatcher
	// owners stores plugin scope by root literal.
	owners map[string]*pluginruntime.Scope
	// players supplies sender permission and feedback operations.
	players sdkcommand.PlayerAccess
	// prefix marks command chat messages.
	prefix string
	// timeout bounds command callbacks.
	timeout time.Duration
	// translations resolves command feedback.
	translations i18n.Translator
	// log records isolated command failures.
	log *zap.Logger
}

// NewTree creates an empty shared command dispatcher.
func NewTree(prefix string, timeout time.Duration, translations i18n.Translator, log *zap.Logger) *Tree {
	if log == nil {
		log = zap.NewNop()
	}
	return &Tree{owners: make(map[string]*pluginruntime.Scope), prefix: prefix, timeout: timeout, translations: translations, log: log}
}

// SetPlayers installs bounded sender operations after backend composition.
func (tree *Tree) SetPlayers(players sdkcommand.PlayerAccess) { tree.players = players }

// Access scopes root ownership to one plugin.
type Access struct {
	// tree stores the shared dispatcher.
	tree *Tree
	// scope stores plugin identity and health.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped command registrar.
func NewAccess(tree *Tree, scope *pluginruntime.Scope) *Access {
	return &Access{tree: tree, scope: scope}
}

// Register adds one unique root command for this plugin.
func (access *Access) Register(command brigodier.LiteralNodeBuilder) error {
	return access.tree.register(access.scope, command)
}

// register validates ownership before modifying Brigadier's root.
func (tree *Tree) register(scope *pluginruntime.Scope, command brigodier.LiteralNodeBuilder) error {
	if scope == nil || command == nil {
		return ErrInvalidCommand
	}
	built := command.BuildLiteral()
	name := strings.TrimSpace(built.Name())
	if name == "" || strings.ContainsAny(name, " \t\r\n") {
		return ErrInvalidCommand
	}
	tree.mutex.Lock()
	defer tree.mutex.Unlock()
	if _, exists := tree.owners[name]; exists {
		return fmt.Errorf("%w: %s", ErrCommandExists, name)
	}
	tree.dispatcher.Register(command)
	tree.owners[name] = scope
	return nil
}

// Execute resolves a command-prefixed chat message.
func (tree *Tree) Execute(ctx context.Context, player sdkplugin.Player, message string) (bool, error) {
	if !strings.HasPrefix(message, tree.prefix) {
		return false, nil
	}
	input := strings.TrimSpace(strings.TrimPrefix(message, tree.prefix))
	sender := sdkcommand.NewPlayerSender(player, tree.players)
	if input == "" {
		return true, sender.Reply(ctx, tree.message("plugin.command.invalid", "Comando incompleto."))
	}
	rootName := strings.Fields(input)[0]
	tree.mutex.RLock()
	scope, found := tree.owners[rootName]
	root := tree.dispatcher.Root.Literals()[rootName]
	tree.mutex.RUnlock()
	if !found || root == nil {
		return true, sender.Reply(ctx, tree.message("plugin.command.unknown", "Comando desconocido."))
	}
	commandContext := sdkcommand.WithSender(ctx, sender)
	if !scope.Enabled() {
		return true, sender.Reply(ctx, tree.message("plugin.command.disabled", "Ese plugin está desactivado."))
	}
	if !root.CanUse(commandContext) {
		return true, sender.Reply(ctx, tree.message("plugin.command.denied", "No tienes permiso para usar ese comando."))
	}
	err := pluginruntime.InvokeCallback(commandContext, tree.timeout, scope, "command "+rootName, tree.log, func(callbackContext context.Context) error {
		return tree.dispatcher.Do(callbackContext, input)
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pluginruntime.ErrCallbackTimeout) {
		return true, sender.Reply(ctx, tree.message("plugin.command.timeout", "El comando tardó demasiado."))
	}
	if errors.Is(err, pluginruntime.ErrCallbackPanic) || errors.Is(err, pluginruntime.ErrPluginDisabled) {
		return true, sender.Reply(ctx, tree.message("plugin.command.failed", "El plugin falló y fue desactivado."))
	}
	tree.log.Debug("plugin command rejected", zap.String("plugin", scope.Name()), zap.String("command", input), zap.Error(err))
	return true, sender.Reply(ctx, tree.message("plugin.command.invalid", "Comando incompleto o inválido."))
}

// message resolves one localized command feedback value.
func (tree *Tree) message(key string, fallback string) string {
	if tree.translations == nil {
		return fallback
	}
	translated := tree.translations.Default(i18n.Key(key))
	if translated == key {
		return fallback
	}
	return translated
}
