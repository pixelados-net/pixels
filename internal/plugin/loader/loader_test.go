package loader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginconfig "github.com/niflaot/pixels/internal/plugin/config"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	pluginhost "github.com/niflaot/pixels/internal/plugin/host"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// fixturePlugin implements the dynamic entrypoint for loader tests.
type fixturePlugin struct {
	// metadata stores the embedded test manifest.
	metadata sdkplugin.Metadata
	// metadataPanic makes manifest evaluation panic.
	metadataPanic bool
	// registerPanic makes registration panic.
	registerPanic bool
	// registerErr stores an ordinary registration failure.
	registerErr error
	// registered records successful registration attempts.
	registered bool
}

// Metadata returns or panics with the configured manifest behavior.
func (plugin *fixturePlugin) Metadata() sdkplugin.Metadata {
	if plugin.metadataPanic {
		panic("metadata panic")
	}
	return plugin.metadata
}

// Register records, rejects, or panics during registration.
func (plugin *fixturePlugin) Register(sdkplugin.Host) error {
	if plugin.registerPanic {
		panic("register panic")
	}
	if plugin.registerErr != nil {
		return plugin.registerErr
	}
	plugin.registered = true
	return nil
}

// fixtureObject stores one exported Plugin symbol.
type fixtureObject struct {
	// symbol stores the lookup result.
	symbol any
	// err stores a lookup error.
	err error
}

// Lookup resolves the required fixed symbol.
func (object fixtureObject) Lookup(string) (any, error) { return object.symbol, object.err }

// fixtureOpener maps object basenames to native object substitutes.
type fixtureOpener struct {
	// objects stores open results by shared-object basename.
	objects map[string]Object
}

// panickingOpener simulates unsafe package initialization.
type panickingOpener struct{}

// Open panics like a broken native initializer.
func (panickingOpener) Open(string) (Object, error) { panic("open panic") }

// Open returns one configured object.
func (opener fixtureOpener) Open(path string) (Object, error) {
	object, found := opener.objects[filepath.Base(path)]
	if !found {
		return nil, os.ErrNotExist
	}
	return object, nil
}

// TestLoadIsolatesInvalidAndPanickingPlugins verifies two-phase startup isolation.
func TestLoadIsolatesInvalidAndPanickingPlugins(t *testing.T) {
	directory := t.TempDir()
	plugins := map[string]*fixturePlugin{
		"healthy.so":  {metadata: fixtureMetadata("healthy")},
		"metadata.so": {metadataPanic: true},
		"register.so": {metadata: fixtureMetadata("register"), registerPanic: true},
		"invalid.so":  {metadata: sdkplugin.Metadata{Name: "invalid"}},
	}
	opener := fixtureOpener{objects: make(map[string]Object, len(plugins))}
	for filename, plugin := range plugins {
		createObjectPath(t, directory, filename)
		opener.objects[filename] = fixtureObject{symbol: sdkplugin.Plugin(plugin)}
	}
	backend := pluginhost.NewBackend(nil, nil, nil, nil, nil, nil, pluginevent.NewHub(time.Second, zap.NewNop()), plugincommand.NewTree(":", time.Second, nil, zap.NewNop()), time.Second, zap.NewNop())
	loader := New(pluginconfig.Config{Directory: directory, CallbackTimeout: time.Second}, backend, opener, zap.NewNop())

	if err := loader.Load(context.Background()); err != nil {
		t.Fatalf("load plugins: %v", err)
	}
	report := loader.Report()
	if !plugins["healthy.so"].registered || len(report.Loaded) != 1 || report.Loaded[0] != "healthy" {
		t.Fatalf("expected healthy plugin load, report=%+v", report)
	}
	if len(report.Skipped) != 3 {
		t.Fatalf("expected isolated failures, report=%+v", report)
	}
}

// TestLoadLogsEmptySummary verifies an empty plugin directory remains observable.
func TestLoadLogsEmptySummary(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core)
	backend := pluginhost.NewBackend(nil, nil, nil, nil, nil, nil, pluginevent.NewHub(time.Second, log), plugincommand.NewTree(":", time.Second, nil, log), time.Second, log)
	loader := New(pluginconfig.Config{Directory: t.TempDir(), CallbackTimeout: time.Second}, backend, fixtureOpener{}, log)

	if err := loader.Load(context.Background()); err != nil {
		t.Fatalf("load empty plugin directory: %v", err)
	}
	entries := logs.FilterMessage("plugin loading completed").All()
	if len(entries) != 1 {
		t.Fatalf("expected one loading summary, got %d", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["discovered"] != int64(0) || fields["loaded"] != int64(0) || fields["skipped"] != int64(0) {
		t.Fatalf("unexpected loading summary %#v", fields)
	}
}

// TestLookupAcceptsInterfaceAndPointerExports verifies supported native symbol shapes.
func TestLookupAcceptsInterfaceAndPointerExports(t *testing.T) {
	entry := &fixturePlugin{metadata: fixtureMetadata("fixture")}
	contract := sdkplugin.Plugin(entry)
	tests := []any{contract, &contract}
	for _, symbol := range tests {
		resolved, err := lookup(fixtureObject{symbol: symbol})
		if err != nil || resolved != entry {
			t.Fatalf("lookup %T: resolved=%v err=%v", symbol, resolved, err)
		}
	}
	if _, err := lookup(fixtureObject{err: errors.New("missing")}); !errors.Is(err, ErrMissingSymbol) {
		t.Fatalf("expected missing symbol, got %v", err)
	}
}

// TestOpenAndLookupRecoverPanics verifies native object boundaries are isolated.
func TestOpenAndLookupRecoverPanics(t *testing.T) {
	if _, err := openObject(panickingOpener{}, "broken.so"); !errors.Is(err, ErrObjectPanic) {
		t.Fatalf("expected open panic recovery, got %v", err)
	}
	object := fixtureObject{symbol: sdkplugin.Plugin(&fixturePlugin{metadata: fixtureMetadata("fixture")})}
	panicking := panicObject{delegate: object}
	if _, err := lookup(panicking); !errors.Is(err, ErrObjectPanic) {
		t.Fatalf("expected lookup panic recovery, got %v", err)
	}
}

// panicObject raises a panic from symbol lookup.
type panicObject struct {
	// delegate documents the otherwise valid underlying object.
	delegate Object
}

// Lookup panics before reaching its delegate.
func (panicObject) Lookup(string) (any, error) { panic("lookup panic") }

// TestReadMetadataRejectsIncompatibleSDK verifies explicit SDK major validation.
func TestReadMetadataRejectsIncompatibleSDK(t *testing.T) {
	entry := &fixturePlugin{metadata: fixtureMetadata("fixture")}
	entry.metadata.SDKVersion = "2.0.0"
	_, err := readMetadata(context.Background(), time.Second, entry, "fixture.so", zap.NewNop())
	if !errors.Is(err, ErrIncompatibleSDK) {
		t.Fatalf("expected incompatible SDK, got %v", err)
	}
}

// createObjectPath creates one discoverable placeholder shared-object path.
func createObjectPath(t *testing.T, root string, filename string) {
	t.Helper()
	directory := filepath.Join(root, filename[:len(filename)-3])
	if err := os.MkdirAll(directory, 0o750); err != nil {
		t.Fatalf("create plugin directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, filename), nil, 0o600); err != nil {
		t.Fatalf("create plugin object: %v", err)
	}
}

// fixtureMetadata creates one complete manifest.
func fixtureMetadata(name string) sdkplugin.Metadata {
	return sdkplugin.Metadata{Name: name, Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion}
}
