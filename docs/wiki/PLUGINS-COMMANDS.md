# Plugin Commands and Permissions

Pixels uses [Brigodier](https://github.com/minekube/brigodier), a Go implementation of Brigadier's command tree model. Plugins register command roots during startup. Room chat beginning with `PIXELS_COMMAND_PREFIX` is parsed by that shared tree before it can become visible speech.

## Why a tree instead of string splitting

A command is a graph of literal and typed argument nodes. Brigodier chooses a valid path, enforces node requirements, converts values, and runs the executor attached to the terminal node. This keeps parsing rules beside the command structure and avoids a different manual parser in every plugin.

```text
hello
  executes hello for the sender
  name as StringPhrase
    executes hello for name
```

The equivalent builder is:

```go
root := brigodier.Literal("hello").
	Executes(brigodier.CommandFunc(func(call *brigodier.CommandContext) error {
		return replyToSender(call, senderName(call))
	})).
	Then(brigodier.Argument("name", brigodier.StringPhrase).
		Executes(brigodier.CommandFunc(func(call *brigodier.CommandContext) error {
			return replyToSender(call, call.String("name"))
		})))
```

`StringPhrase` consumes the remaining text as one value. Brigodier also provides typed arguments such as `Int`. A malformed or incomplete path is consumed as a command and produces localized feedback rather than appearing in room chat.

## The command sender

Pixels places an `sdk/command.Sender` in the command context. It represents the origin without forcing command code to depend on a live player type.

```go
func replyToSender(call *brigodier.CommandContext, name string) error {
	sender, found := sdkcommand.SenderFrom(call.Context)
	if !found {
		return nil
	}

	return sender.Reply(call.Context, "Hello, "+name)
}

func senderName(call *brigodier.CommandContext) string {
	sender, found := sdkcommand.SenderFrom(call.Context)
	if !found {
		return "world"
	}

	return sender.Name()
}
```

`Name` returns a display identity. `Kind` distinguishes `player` and future trusted `console` callers. `Reply` sends feedback through the originating channel. `HasPermission` resolves the real Pixels permission system.

## Declaring plugin permissions

A plugin registers only its local suffix:

```go
err := host.Permissions().Register(
	"hello.use",
	"Use the hello command",
)
```

Pixels expands that suffix using the manifest name:

```text
plugin.hello-plugin.hello.use
```

The expanded node joins the normal permission catalog. Operators grant or deny it to permission groups or directly to a player through the same administration API used for native capabilities. There is no plugin rank table and no independent permission resolver.

Local names must be concrete dotted nodes. They cannot begin with `plugin.`, contain a wildcard, or use another plugin's namespace. The description is required because it is displayed with the runtime node catalog.

## Guarding a command node

Requirements belong on the earliest node that should be hidden from a sender. Placing the requirement on the root protects every descendant.

```go
const helloPermission = "plugin.hello-plugin.hello.use"

root := brigodier.Literal("hello").
	Requires(sdkcommand.RequiresPermission(helloPermission)).
	Executes(brigodier.CommandFunc(runHello))

err := host.Commands().Register(root)
```

`RequiresPermission` reads the sender from context and calls its `HasPermission`. A player without the node receives the localized denied response. The executor is never called.

Only one plugin may own a root literal. Registering another `hello` root fails startup registration for the second plugin. Child literals and arguments belong to the root owner.

## Chat handling rules

| Input | Result |
|---|---|
| Does not begin with the configured prefix | Continues as normal room chat |
| Prefix with no command | Consumed with incomplete command feedback |
| Unknown root | Consumed with unknown command feedback |
| Known root without permission | Consumed with denied feedback |
| Malformed argument | Consumed with invalid command feedback |
| Valid command | Executor runs within the plugin callback deadline |
| Panic | Plugin is disabled and the sender receives failure feedback |
| Timeout | Sender receives timeout feedback |

Commands never echo the raw prefixed text to the room. This is important for administrative arguments that may contain player identifiers or operational data.

## Permission resolution

The command helper does not decide authorization itself. It delegates to the hotel wide algorithm documented in [[USERS-PERMISSIONS]]. Direct player overrides, group weights, inheritance, wildcard specificity, and deny precedence all apply to plugin nodes exactly as they apply to native nodes.
