package loader

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// invokeRegister protects one plugin registration boundary.
func invokeRegister(ctx context.Context, timeout time.Duration, scope *pluginruntime.Scope, entry sdkplugin.Plugin, host sdkplugin.Host, log *zap.Logger) error {
	return invokeBoundary(ctx, timeout, scope, "register", log, func() error { return entry.Register(host) })
}

// invokeBoundary bounds metadata and registration callbacks without exposing runtime internals.
func invokeBoundary(ctx context.Context, timeout time.Duration, scope *pluginruntime.Scope, kind string, log *zap.Logger, callback func() error) error {
	callbackContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			if recovered := recover(); recovered != nil {
				scope.Disable()
				if log != nil {
					log.Error("plugin loader callback panicked", zap.String("plugin", scope.Name()), zap.String("callback", kind), zap.Any("panic", recovered), zap.ByteString("stack", debug.Stack()))
				}
				err = fmt.Errorf("%w: %v", pluginruntime.ErrCallbackPanic, recovered)
			}
			done <- err
		}()
		err = callback()
	}()
	select {
	case err := <-done:
		return err
	case <-callbackContext.Done():
		return pluginruntime.ErrCallbackTimeout
	}
}
