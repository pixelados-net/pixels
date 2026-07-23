// Package loader discovers and registers native Go plugins at process startup.
package loader

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pluginconfig "github.com/niflaot/pixels/internal/plugin/config"
	pluginhost "github.com/niflaot/pixels/internal/plugin/host"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

var (
	// ErrMissingSymbol reports a native object without the required Plugin export.
	ErrMissingSymbol = errors.New("plugin symbol missing")
	// ErrInvalidSymbol reports a Plugin export that does not implement the SDK contract.
	ErrInvalidSymbol = errors.New("invalid plugin symbol")
	// ErrInvalidMetadata reports an incomplete or unsafe embedded manifest.
	ErrInvalidMetadata = errors.New("invalid plugin metadata")
	// ErrIncompatibleSDK reports a plugin compiled for another SDK major.
	ErrIncompatibleSDK = errors.New("incompatible plugin sdk")
	// ErrDependency reports missing, cyclic, or failed plugin dependencies.
	ErrDependency = errors.New("plugin dependency unavailable")
	// ErrObjectPanic reports a recovered native open or symbol lookup panic.
	ErrObjectPanic = errors.New("plugin object panicked")
)

// Object resolves exported symbols from one native plugin object.
type Object interface {
	// Lookup returns one exported symbol.
	Lookup(string) (any, error)
}

// Opener opens one native plugin object.
type Opener interface {
	// Open loads one .so and executes its package initialization.
	Open(string) (Object, error)
}

// candidate stores one opened and validated plugin before registration.
type candidate struct {
	// path stores the native object location.
	path string
	// plugin stores the SDK entrypoint.
	plugin sdkplugin.Plugin
	// metadata stores the embedded manifest.
	metadata sdkplugin.Metadata
}

// Report summarizes one plugin discovery and registration run.
type Report struct {
	// Discovered counts .so candidates found on disk.
	Discovered int
	// Loaded lists successfully registered plugin names.
	Loaded []string
	// Skipped maps skipped paths or plugin names to explanatory errors.
	Skipped map[string]string
}

// Loader coordinates two-phase native discovery and registration.
type Loader struct {
	// config stores directory and callback bounds.
	config pluginconfig.Config
	// backend creates scoped plugin host facades.
	backend *pluginhost.Backend
	// opener loads native objects.
	opener Opener
	// log records isolated discovery failures.
	log *zap.Logger
	// mutex protects the latest report.
	mutex sync.RWMutex
	// report stores the latest immutable run summary.
	report Report
}

// New creates a native plugin loader.
func New(config pluginconfig.Config, backend *pluginhost.Backend, opener Opener, log *zap.Logger) *Loader {
	if log == nil {
		log = zap.NewNop()
	}
	return &Loader{config: config.Normalize(), backend: backend, opener: opener, log: log, report: Report{Skipped: make(map[string]string)}}
}

// Load discovers manifests first and then registers plugins in dependency order.
func (loader *Loader) Load(ctx context.Context) error {
	paths, err := discover(loader.config.Directory)
	if err != nil {
		loader.log.Error("plugin directory scan failed", zap.String("directory", loader.config.Directory), zap.Error(err))
		return nil
	}
	report := Report{Discovered: len(paths), Skipped: make(map[string]string)}
	candidates := loader.openCandidates(ctx, paths, report.Skipped)
	ordered, blocked := resolve(candidates)
	for name, resolveErr := range blocked {
		report.Skipped[name] = resolveErr.Error()
		loader.log.Error("plugin dependency resolution failed", zap.String("plugin", name), zap.Error(resolveErr))
	}
	loaded := make(map[string]bool, len(ordered))
	for _, current := range ordered {
		if dependency := failedDependency(current.metadata.Dependencies, loaded); dependency != "" {
			err = fmt.Errorf("%w: %s", ErrDependency, dependency)
			report.Skipped[current.metadata.Name] = err.Error()
			loader.log.Error("plugin skipped after dependency failed", zap.String("plugin", current.metadata.Name), zap.String("dependency", dependency))
			continue
		}
		scope := pluginruntime.NewScope(current.metadata.Name)
		host := loader.backend.HostFor(scope)
		err = invokeRegister(ctx, loader.config.CallbackTimeout, scope, current.plugin, host, loader.log)
		if err != nil {
			scope.Disable()
			report.Skipped[current.metadata.Name] = err.Error()
			loader.log.Error("plugin registration failed", zap.String("plugin", current.metadata.Name), zap.String("path", current.path), zap.Error(err))
			continue
		}
		loaded[current.metadata.Name] = true
		report.Loaded = append(report.Loaded, current.metadata.Name)
		loader.log.Info("plugin loaded", zap.String("plugin", current.metadata.Name), zap.String("version", current.metadata.Version), zap.String("author", current.metadata.Author), zap.String("path", current.path))
	}
	loader.mutex.Lock()
	loader.report = report
	loader.mutex.Unlock()
	loader.log.Info(
		"plugin loading completed",
		zap.String("directory", loader.config.Directory),
		zap.Int("discovered", report.Discovered),
		zap.Int("loaded", len(report.Loaded)),
		zap.Int("skipped", len(report.Skipped)),
	)

	return nil
}

// Report returns a stable copy of the latest load summary.
func (loader *Loader) Report() Report {
	loader.mutex.RLock()
	defer loader.mutex.RUnlock()
	report := loader.report
	report.Loaded = append([]string(nil), report.Loaded...)
	report.Skipped = make(map[string]string, len(loader.report.Skipped))
	for name, reason := range loader.report.Skipped {
		report.Skipped[name] = reason
	}

	return report
}

// openCandidates opens every object and validates embedded manifests.
func (loader *Loader) openCandidates(ctx context.Context, paths []string, skipped map[string]string) []candidate {
	candidates := make([]candidate, 0, len(paths))
	for _, path := range paths {
		object, err := openObject(loader.opener, path)
		if err != nil {
			skipped[path] = err.Error()
			loader.log.Error("plugin object open failed", zap.String("path", path), zap.Error(err))
			continue
		}
		entry, err := lookup(object)
		if err != nil {
			skipped[path] = err.Error()
			loader.log.Error("plugin symbol lookup failed", zap.String("path", path), zap.Error(err))
			continue
		}
		metadata, err := readMetadata(ctx, loader.config.CallbackTimeout, entry, path, loader.log)
		if err != nil {
			skipped[path] = err.Error()
			loader.log.Error("plugin metadata rejected", zap.String("path", path), zap.Error(err))
			continue
		}
		candidates = append(candidates, candidate{path: path, plugin: entry, metadata: metadata})
	}

	return candidates
}
