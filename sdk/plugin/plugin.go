// Package plugin contains the controlled dynamic-plugin SDK facade.
package plugin

import (
	"context"

	"github.com/gofiber/fiber/v2"
	sdkcommand "github.com/niflaot/pixels/sdk/command"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
	sdkpriority "github.com/niflaot/pixels/sdk/priority"
	"go.minekube.com/brigodier"
)

const (
	// SDKVersion is the semantic version implemented by this host SDK.
	SDKVersion = "1.0.0"
	// PriorityLowest runs before no lower-priority callback.
	PriorityLowest Priority = sdkpriority.Lowest
	// PriorityLow runs below normal callbacks.
	PriorityLow Priority = sdkpriority.Low
	// PriorityNormal is the default callback priority.
	PriorityNormal Priority = sdkpriority.Normal
	// PriorityHigh runs above normal callbacks.
	PriorityHigh Priority = sdkpriority.High
	// PriorityHighest runs before ordinary high-priority callbacks.
	PriorityHighest Priority = sdkpriority.Highest
	// PriorityMonitor conventionally observes after every mutation callback.
	PriorityMonitor Priority = sdkpriority.Monitor
)

// Player aliases the immutable plugin-facing player snapshot.
type Player = sdkplayer.Player

// Priority aliases the shared callback execution order.
type Priority = sdkpriority.Priority

// Metadata describes one plugin's embedded manifest.
type Metadata struct {
	// Name identifies the plugin and its route namespace.
	Name string
	// Version identifies the plugin release.
	Version string
	// Author identifies the plugin publisher.
	Author string
	// SDKVersion identifies the SDK major used to compile the plugin.
	SDKVersion string
	// Dependencies lists plugins that must register first.
	Dependencies []string
}

// Valid reports whether metadata contains every required identity field.
func (metadata Metadata) Valid() bool {
	return metadata.Name != "" && metadata.Version != "" && metadata.Author != "" && metadata.SDKVersion != ""
}

// Plugin is the entrypoint every dynamic object exports as Plugin.
type Plugin interface {
	// Metadata returns the embedded plugin manifest.
	Metadata() Metadata
	// Register wires the plugin against capability-scoped host services.
	Register(Host) error
}

// Host is the capability-scoped facade supplied during plugin registration.
type Host interface {
	// Players returns bounded live-player access.
	Players() PlayerAccess
	// Routes returns isolated HTTP route registration.
	Routes() RouteRegistrar
	// Events returns plugin event registration.
	Events() EventHub
	// Commands returns the shared Brigadier command tree.
	Commands() CommandTree
	// Permissions returns namespaced permission-node registration.
	Permissions() PermissionRegistrar
}

// PlayerAccess exposes bounded live-player read and action operations.
type PlayerAccess interface {
	// All returns immutable snapshots of connected players.
	All() []Player
	// Find returns one immutable connected-player snapshot.
	Find(int64) (Player, bool)
	// Message sends one system alert to a connected player.
	Message(int64, string) error
	// Disconnect ends one live player session.
	Disconnect(int64, string) error
	// HasPermission resolves a real permission node for one player.
	HasPermission(int64, string) (bool, error)
	// Intercept registers one inbound packet middleware.
	Intercept(PacketInterceptor, InterceptOptions) error
}

// InterceptContext carries one inbound packet through plugin middleware.
type InterceptContext struct {
	// Player stores the immutable source player when authenticated.
	Player Player
	// Header identifies the inbound packet.
	Header uint16
	// Payload stores a private copy of the encoded packet body.
	Payload []byte
}

// Next invokes the next interceptor and eventually the native handler.
type Next func(context.Context) error

// PacketInterceptor observes or cancels one inbound packet.
type PacketInterceptor func(context.Context, InterceptContext, Next) error

// InterceptOptions configures packet filtering and priority.
type InterceptOptions struct {
	// Priority controls global execution order.
	Priority Priority
	// Header limits interception to one header when non-nil.
	Header *uint16
}

// RouteRegistrar mounts routes inside one plugin-owned prefix.
type RouteRegistrar interface {
	// Mount registers handlers below /plugins/<pluginName>.
	Mount(string, func(fiber.Router)) error
	// Describe publishes a separate plugin-owned OpenAPI document.
	Describe(string, []byte) error
}

// EventHub registers prioritized listeners for typed plugin events.
type EventHub interface {
	// Listen registers one listener by stable event name.
	Listen(string, sdkevent.ListenerOptions, sdkevent.Listener) error
}

// CommandTree registers root commands in the shared Brigadier dispatcher.
type CommandTree interface {
	// Register adds one unique root command.
	Register(brigodier.LiteralNodeBuilder) error
}

// PermissionRegistrar declares namespaced plugin permission nodes.
type PermissionRegistrar interface {
	// Register declares plugin.<pluginName>.<node> in the host permission catalog.
	Register(string, string) error
}

// Sender aliases the shared command sender contract for convenient plugin imports.
type Sender = sdkcommand.Sender
