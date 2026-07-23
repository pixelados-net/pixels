# Creating a Plugin

This page builds a native Pixels plugin from an empty directory to a loadable shared object. Read [[PLUGINS-OVERVIEW]] first for the trust and compatibility model, then continue with [[PLUGINS-LISTENERS]] and [[PLUGINS-COMMANDS]] for the two callback systems.

## One plugin is one Go module

A plugin should not live inside the Pixels module. Keeping it separate makes the public `sdk/` boundary real and lets the Go compiler reject imports from `internal/`.

```text
hello-plugin/
  go.mod
  go.sum
  main.go
```

For local development, point the dependency at the exact Pixels checkout used to build the server:

```go
module example.net/hello-plugin

go 1.26

require github.com/niflaot/pixels v0.0.0

replace github.com/niflaot/pixels => /absolute/path/to/pixels
```

The replace directive is a development convenience. A published plugin should depend on the exact tagged Pixels SDK release it supports.

## The exported entrypoint

The loader resolves one symbol named `Plugin`. Its value must implement `sdk/plugin.Plugin`, which has only `Metadata` and `Register`.

```go
package main

import sdkplugin "github.com/niflaot/pixels/sdk/plugin"

// Plugin is the fixed symbol resolved by Pixels.
var Plugin sdkplugin.Plugin = &helloPlugin{}

// helloPlugin implements the plugin entrypoint.
type helloPlugin struct{}

// Metadata returns the identity embedded in the shared object.
func (*helloPlugin) Metadata() sdkplugin.Metadata {
	return sdkplugin.Metadata{
		Name:       "hello-plugin",
		Version:    "1.0.0",
		Author:     "Example author",
		SDKVersion: sdkplugin.SDKVersion,
	}
}

// Register declares every capability used by the plugin.
func (*helloPlugin) Register(host sdkplugin.Host) error {
	return nil
}
```

The manifest name is also the permission and HTTP namespace. It accepts lowercase ASCII letters, digits, and internal hyphens. `Dependencies` contains manifest names that must register first. Pixels rejects duplicate names, missing dependencies, dependency cycles, malformed metadata, and an incompatible SDK major before the plugin can handle traffic.

## Registration is the construction phase

`Register` runs once during server startup. It is where the plugin declares permissions, command roots, event listeners, packet interceptors, routes, and an optional OpenAPI document. Registration should not start unmanaged goroutines or perform slow remote work.

The host exposes five bounded capabilities:

| Capability | Purpose |
|---|---|
| `Players()` | Immutable online player snapshots, alerts, disconnects, permission checks, and inbound interceptors |
| `Events()` | Typed listeners for host approved events |
| `Commands()` | Root registration in the shared Brigadier command tree |
| `Permissions()` | Runtime declaration of namespaced permission nodes |
| `Routes()` | Private Fiber routes and a plugin owned OpenAPI document |

Return the first registration error. A failed registration disables that plugin and causes dependants to be skipped, while unrelated plugins continue loading.

```go
func (plugin *helloPlugin) Register(host sdkplugin.Host) error {
	if err := host.Permissions().Register("hello.use", "Use the hello command"); err != nil {
		return err
	}
	if err := plugin.registerCommands(host); err != nil {
		return err
	}
	if err := plugin.registerListeners(host); err != nil {
		return err
	}

	return plugin.registerRoutes(host)
}
```

Splitting registration by capability keeps the plugin navigable as it grows. A practical layout mirrors Pixels realms:

```text
hello-plugin/
  command/
  config/
  event/
  permission/
  player/
  route/
  main.go
```

## Live player access

`All` and `Find` return copied `sdk/player.Player` values. A snapshot contains the player id, username, current room id, and online state. It cannot mutate a live realm object.

```go
current, found := host.Players().Find(42)
if found {
	_ = host.Players().Message(current.ID, "Hello from a plugin")
}
```

`Message` sends a Nitro system alert. `Disconnect` closes the active session with a plugin authored reason. `HasPermission` resolves a concrete node through the same permission engine used by the emulator. Offline players are not returned because SDK 1.x deliberately exposes live operations, not repositories.

## Private HTTP routes

A plugin may mount only its own namespace. The following handler becomes `GET /plugins/hello-plugin/health` and inherits the server's `X-API-Key` protection.

```go
func (plugin *helloPlugin) registerRoutes(host sdkplugin.Host) error {
	return host.Routes().Mount("hello-plugin", func(router fiber.Router) {
		router.Get("/health", func(ctx *fiber.Ctx) error {
			return ctx.JSON(fiber.Map{"status": "ok"})
		})
	})
}
```

`Describe` accepts an independent OpenAPI JSON document served at `/plugins/hello-plugin/openapi.json`. Plugin operations never modify Pixels' central OpenAPI specification.

## Build and install

Create the manifest owned folder and compile with native plugin mode:

```sh
mkdir -p /path/to/pixels/plugins/hello-plugin
CGO_ENABLED=1 go build -buildmode=plugin \
  -o /path/to/pixels/plugins/hello-plugin/plugin.so .
```

Start Pixels and look for both `plugin loaded` and the final `plugin loading completed` summary. Native ABI compatibility is exact, so follow [[PLUGINS-DEPLOYMENT]] whenever the host, SDK, toolchain, platform, or a shared dependency changes.
