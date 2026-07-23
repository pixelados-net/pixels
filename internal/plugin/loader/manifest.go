package loader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// discover returns direct plugins/<name>/*.so paths in stable order.
func discover(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		objects, globErr := filepath.Glob(filepath.Join(directory, entry.Name(), "*.so"))
		if globErr != nil {
			return nil, globErr
		}
		paths = append(paths, objects...)
	}
	sort.Strings(paths)
	return paths, nil
}

// openObject recovers failures raised while the native loader initializes code.
func openObject(opener Opener, path string) (object Object, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			object = nil
			err = fmt.Errorf("%w during open: %v", ErrObjectPanic, recovered)
		}
	}()
	return opener.Open(path)
}

// lookup resolves common exported variable forms into the Plugin interface.
func lookup(object Object) (entry sdkplugin.Plugin, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			entry = nil
			err = fmt.Errorf("%w during lookup: %v", ErrObjectPanic, recovered)
		}
	}()
	symbol, err := object.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMissingSymbol, err)
	}
	if entry, ok := symbol.(sdkplugin.Plugin); ok {
		return entry, nil
	}
	if pointer, ok := symbol.(*sdkplugin.Plugin); ok && pointer != nil && *pointer != nil {
		return *pointer, nil
	}
	value := reflect.ValueOf(symbol)
	if value.IsValid() && value.Kind() == reflect.Pointer && !value.IsNil() {
		if entry, ok := value.Elem().Interface().(sdkplugin.Plugin); ok {
			return entry, nil
		}
	}
	return nil, ErrInvalidSymbol
}

// readMetadata recovers and validates one embedded manifest.
func readMetadata(ctx context.Context, timeout time.Duration, entry sdkplugin.Plugin, path string, log *zap.Logger) (metadata sdkplugin.Metadata, err error) {
	scope := pluginruntime.NewScope(filepath.Base(path))
	err = invokeBoundary(ctx, timeout, scope, "metadata", log, func() error {
		metadata = entry.Metadata()
		return nil
	})
	if err != nil {
		return sdkplugin.Metadata{}, err
	}
	if !validMetadata(metadata) {
		return sdkplugin.Metadata{}, ErrInvalidMetadata
	}
	if major(metadata.SDKVersion) != major(sdkplugin.SDKVersion) {
		return sdkplugin.Metadata{}, fmt.Errorf("%w: plugin=%s host=%s", ErrIncompatibleSDK, metadata.SDKVersion, sdkplugin.SDKVersion)
	}
	return metadata, nil
}

// validMetadata verifies identity, namespace, and dependency declarations.
func validMetadata(metadata sdkplugin.Metadata) bool {
	if !metadata.Valid() || !validName(metadata.Name) {
		return false
	}
	seen := make(map[string]struct{}, len(metadata.Dependencies))
	for _, dependency := range metadata.Dependencies {
		if !validName(dependency) || dependency == metadata.Name {
			return false
		}
		if _, exists := seen[dependency]; exists {
			return false
		}
		seen[dependency] = struct{}{}
	}
	return validVersion(metadata.Version) && validVersion(metadata.SDKVersion)
}

// validName reports whether a manifest name is one safe route segment.
func validName(name string) bool {
	if name == "" || name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	for _, value := range name {
		if (value < 'a' || value > 'z') && (value < '0' || value > '9') && value != '-' {
			return false
		}
	}
	return true
}

// major parses the semantic major component or returns minus one.
func major(version string) int {
	value := strings.TrimPrefix(strings.TrimSpace(version), "v")
	part, _, _ := strings.Cut(value, ".")
	parsed, err := strconv.Atoi(part)
	if err != nil || parsed < 0 {
		return -1
	}
	return parsed
}

// validVersion reports whether a value is a three-part non-negative semantic version.
func validVersion(version string) bool {
	value := strings.TrimPrefix(strings.TrimSpace(version), "v")
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < 0 {
			return false
		}
	}
	return true
}

// failedDependency returns the first dependency that did not register.
func failedDependency(dependencies []string, loaded map[string]bool) string {
	for _, dependency := range dependencies {
		if !loaded[dependency] {
			return dependency
		}
	}
	return ""
}
