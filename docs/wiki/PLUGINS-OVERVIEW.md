# Native Plugins

First page of the Plugins section. Pixels can discover native Go shared
objects at startup and give them a deliberately narrow SDK: live-player
snapshots and actions, inbound packet interceptors, isolated private HTTP
routes, typed events, chat commands, and namespaced permissions. Continue with
[[PLUGINS-CREATING]] for the first complete plugin, [[PLUGINS-LISTENERS]] for
events and packet middleware, [[PLUGINS-COMMANDS]] for Brigodier and permissions,
[[PLUGINS-SDK]] for the capability reference, and [[PLUGINS-DEPLOYMENT]] for the
build and operations contract.

## Startup model

The loader scans exactly one level below `PIXELS_PLUGIN_DIRECTORY`:

```text
plugins/
  example-plugin/
    plugin.so
    config-owned-by-the-plugin.json
```

Only `*.so` files are interpreted by Pixels. Configuration and data beside a
plugin belong entirely to that plugin. Identity is embedded in the binary; no
external manifest can drift away from it.

Startup has two phases. Pixels first opens every object and reads every
manifest. It rejects malformed metadata, duplicate names, missing dependencies
and dependency cycles before registration begins. It then registers the
remaining graph in dependency-first order. A plugin whose dependency fails to
register is skipped without blocking unrelated plugins.

Every plugin exports one symbol:

```go
var Plugin plugin.Plugin = &example{}

func (*example) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:       "example-plugin",
		Version:    "1.0.0",
		Author:     "Example author",
		SDKVersion: plugin.SDKVersion,
	}
}
```

Names contain lowercase ASCII letters, digits and internal hyphens. A manifest
may list other manifest names in `Dependencies`; it may not depend on itself or
repeat a dependency.

## Isolation boundary

Plugins run in the server process, so this is a capability boundary rather than
a security sandbox. The host never hands a plugin PostgreSQL, Redis, mutable
realm objects, the internal event publisher, or raw connection registries. It
provides immutable player values and specific actions instead.

Pixels recovers panics at manifest, registration, interceptor, listener,
command, route-mount and route-handler boundaries. A callback panic disables
that plugin scope while other plugins and native behavior keep running.
Interceptor, event and command calls are also bounded by
`PIXELS_PLUGIN_CALLBACK_TIMEOUT`. Timeouts protect the caller, but Go cannot
forcibly terminate the plugin's goroutine; plugin code must honor context
cancellation and must not start unowned background work.

## Native ABI limitation

Go native plugins have no stable cross-build ABI. The host and `.so` must use
the same Go release, `GOOS`, `GOARCH`, SDK source, and versions of every shared
dependency. `SDKVersion` gives an early, understandable rejection for known
major SDK mismatches, but cannot replace that native requirement.

Native loading is supported where Go's standard `plugin` package is supported:
Linux, macOS and FreeBSD. A Windows host reports that native plugins are
unsupported. See [[PLUGINS-DEPLOYMENT]] before compiling or moving a binary.
