# Plugin SDK

This page is the compact capability reference. The public contracts live below
`sdk/`; host implementations follow the same capability first navigation as
realms, under `internal/plugin/{command,event,player,route,permission,...}`.
Use [[PLUGINS-CREATING]], [[PLUGINS-LISTENERS]], and [[PLUGINS-COMMANDS]] for
the complete tutorials.

## Registration and permissions

`Register(plugin.Host) error` is called once at startup. Register all
capabilities there and return the first error: a partial or failed plugin is
disabled. Permission names passed to `Permissions().Register` are local, while
checks and command requirements use their full name:

```go
host.Permissions().Register("hello.use", "Use the hello command")
// Full node: plugin.example-plugin.hello.use
```

The node joins Pixels' normal permission catalog and can be granted through the
existing group or direct-player administration API. Plugins do not maintain a
parallel rank system.

## Players and packet interceptors

`host.Players().All()` and `Find(id)` return copied `sdk/player.Player` values.
`Message`, `Disconnect` and `HasPermission` are the only player actions in SDK
1.x. `Message` renders as a Nitro system alert.

`Intercept` registers global middleware when `Header` is nil, or middleware for
one inbound packet header otherwise. Larger priorities execute first; equal
priorities preserve registration order. Calling `next(ctx)` advances to the
next plugin and eventually the native handler. Returning without calling it
cancels the packet. Payload bytes are copied before entering plugin code.

```go
host.Players().Intercept(func(ctx context.Context, packet plugin.InterceptContext, next plugin.Next) error {
	log.Printf("inbound header=%d player=%d", packet.Header, packet.Player.ID)
	return next(ctx)
}, plugin.InterceptOptions{Priority: plugin.PriorityLow})
```

## Events

The plugin event hub is separate from Pixels' post-commit internal bus. A plugin
can subscribe but cannot publish arbitrary realm events. SDK 1.x exposes:

| Event | Behavior |
|---|---|
| `player.connected` | Notification after authentication |
| `chat.send` | Mutable and cancellable event after native filtering and before WIRED or room delivery |

Listeners run from larger to smaller priority. `IgnoreCancelled` skips a
listener when a previous one already vetoed a cancellable event. A failure in
one listener is logged and does not stop healthy listeners.

## Chat commands

Commands use `go.minekube.com/brigodier`. Messages beginning with
`PIXELS_COMMAND_PREFIX` are consumed by the command tree before normal room
chat. They never appear as literal speech, including unknown, denied and
malformed commands.

```go
root := brigodier.Literal("hello").
	Requires(command.RequiresPermission("plugin.example-plugin.hello.use")).
	Then(brigodier.Argument("name", brigodier.StringPhrase).
		Executes(brigodier.CommandFunc(func(call *brigodier.CommandContext) error {
			sender, _ := command.SenderFrom(call.Context)
			return sender.Reply(call.Context, "Hello, "+call.String("name"))
		})))

err := host.Commands().Register(root)
```

Only one plugin can own a root literal. Command feedback is localized by the
host; `Sender` deliberately supports players and future console/system callers.

## HTTP routes and OpenAPI

Routes are mounted only below `/plugins/<manifest-name>`, after the same global
`X-API-Key` middleware as every private Pixels route. A plugin cannot claim
another namespace. `Describe` accepts a valid JSON document and serves it at
`/plugins/<name>/openapi.json`; plugin paths never mutate Pixels' central
OpenAPI document.

```go
host.Routes().Mount("example-plugin", func(router fiber.Router) {
	router.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok"})
	})
})
```
