# Building and Deploying Plugins

Final page of the Plugins section. Native compatibility is exact: compile the
host and plugin from the same Pixels checkout with the same toolchain and
dependency graph.

## Create a separate module

Keep plugin code outside the main server packages. During local development a
replace directive can target the exact checkout:

```go
module example.net/hello-plugin

go 1.26

require github.com/niflaot/pixels v0.0.0

replace github.com/niflaot/pixels => /absolute/path/to/pixels
```

Import only the public contracts below `github.com/niflaot/pixels/sdk` and the
types those contracts explicitly expose, such as Fiber and Brigodier builders.
Do not import `internal/`: Go rejects that boundary for an external module.

## Verify compatibility and build

Run these commands in both build environments and compare their output:

```bash
go version
go env GOOS GOARCH CGO_ENABLED
go list -m all
sha256sum go.sum
```

For the current documented development build the toolchain is Go 1.26.1. The
authoritative `go.sum` is always the file shipped with the exact Pixels release;
release source and build artifacts must be retained together.

Build into the manifest-owned folder:

```bash
mkdir -p /path/to/pixels/plugins/example-plugin
CGO_ENABLED=1 go build -buildmode=plugin \
  -o /path/to/pixels/plugins/example-plugin/plugin.so .
```

Recompile after any host SDK, Go, Fiber, Brigodier or transitive shared-package
change. A `.so` compiled for macOS cannot be mounted into the Alpine container;
compile it with the same Linux/musl build image used for the host.

## Runtime configuration

```dotenv
PIXELS_PLUGIN_DIRECTORY=plugins
PIXELS_PLUGIN_CALLBACK_TIMEOUT=2s
PIXELS_COMMAND_PREFIX=:
```

A missing directory is allowed and loads zero plugins. Plugins load only during
startup; adding or replacing a `.so` requires a server restart. The startup log
records each discovered path, loaded manifest and isolated rejection reason.

For containers, mount a directory at `/app/plugins` and keep the default
relative setting, or set an explicit container path. The runtime image includes
the C runtime required by native Go plugin objects.

## Smoke test

1. Start Pixels and verify `plugin loaded` includes the manifest name and
   version.
2. Grant every declared `plugin.<name>.*` permission through the normal admin
   permission routes.
3. Test command success, denial and malformed input from Nitro.
4. Exercise event mutation/cancellation with two clients.
5. Exercise intercepted packets and ensure native handling continues when
   `next` is called.
6. Request plugin HTTP routes once without `X-API-Key` (expect `401`) and once
   with it (expect the plugin response).
7. Request `/plugins/<name>/openapi.json` if the plugin published a document.
8. Trigger any plugin-specific panic/timeout QA hooks only in development and
   verify the plugin is isolated without disconnecting unrelated clients.
