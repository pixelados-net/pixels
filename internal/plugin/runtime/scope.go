// Package runtime implements the host-side dynamic plugin capabilities.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

var (
	// ErrPluginDisabled reports a callback from a disabled plugin.
	ErrPluginDisabled = errors.New("plugin disabled")
	// ErrCallbackTimeout reports a callback that exceeded its deadline.
	ErrCallbackTimeout = errors.New("plugin callback timeout")
	// ErrCallbackPanic reports a recovered plugin panic.
	ErrCallbackPanic = errors.New("plugin callback panic")
	// ErrNextUnavailable reports an interceptor next call outside its callback.
	ErrNextUnavailable = errors.New("plugin interceptor next unavailable")
	// ErrWrongNamespace reports a plugin attempting to claim another namespace.
	ErrWrongNamespace = errors.New("plugin namespace does not match manifest")
	// ErrInvalidPermission reports an invalid plugin-local permission node.
	ErrInvalidPermission = errors.New("invalid plugin permission")
)

// Scope tracks one plugin's global enabled state.
type Scope struct {
	// name stores the plugin namespace.
	name string
	// enabled reports whether callbacks may still execute.
	enabled atomic.Bool
}

// NewScope creates one enabled plugin scope.
func NewScope(name string) *Scope {
	scope := &Scope{name: name}
	scope.enabled.Store(true)

	return scope
}

// Name returns the plugin namespace.
func (scope *Scope) Name() string { return scope.name }

// Enabled reports whether plugin callbacks may execute.
func (scope *Scope) Enabled() bool { return scope != nil && scope.enabled.Load() }

// Disable prevents future callbacks from this plugin.
func (scope *Scope) Disable() {
	if scope != nil {
		scope.enabled.Store(false)
	}
}

// callbackResult stores one protected asynchronous callback outcome.
type callbackResult struct {
	// err stores the callback result.
	err error
	// panicValue stores one recovered panic.
	panicValue any
	// stack stores the recovered panic stack.
	stack []byte
}

// InvokeCallback executes one plugin callback with panic recovery and timeout.
func InvokeCallback(ctx context.Context, timeout time.Duration, scope *Scope, kind string, log *zap.Logger, callback func(context.Context) error) error {
	if !scope.Enabled() {
		return ErrPluginDisabled
	}
	callbackContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	done := make(chan callbackResult, 1)
	go func() {
		result := callbackResult{}
		defer func() {
			if recovered := recover(); recovered != nil {
				result.panicValue = recovered
				result.stack = debug.Stack()
			}
			done <- result
		}()
		result.err = callback(callbackContext)
	}()

	select {
	case result := <-done:
		if result.panicValue == nil {
			return result.err
		}
		scope.Disable()
		if log != nil {
			log.Error("plugin callback panicked", zap.String("plugin", scope.Name()), zap.String("callback", kind), zap.Any("panic", result.panicValue), zap.ByteString("stack", result.stack))
		}
		return fmt.Errorf("%w: %v", ErrCallbackPanic, result.panicValue)
	case <-callbackContext.Done():
		if log != nil {
			log.Error("plugin callback timed out", zap.String("plugin", scope.Name()), zap.String("callback", kind), zap.Duration("timeout", timeout))
		}
		return ErrCallbackTimeout
	}
}
