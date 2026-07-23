package connection

import (
	"cmp"
	"slices"
	"sync"
	"time"

	"github.com/niflaot/pixels/networking/codec"
)

// Context describes the connection state visible to packet handlers.
type Context struct {
	// ConnectionID is the handled connection identifier.
	ConnectionID ID
	// ConnectionKind is the handled connection kind.
	ConnectionKind Kind
	// Direction is the packet flow direction.
	Direction Direction
	// State is the connection lifecycle phase.
	State State
	// StartedAt is the connection start time.
	StartedAt time.Time
	// RemoteAddr is the transport peer address.
	RemoteAddr string
	// AuthenticatedAt is the authentication time when authenticated.
	AuthenticatedAt time.Time
	// Authenticated reports whether authentication completed.
	Authenticated bool
	// Disconnected reports whether the connection is disposed.
	Disconnected bool
	// DisconnectReason stores the disposal reason when disconnected.
	DisconnectReason Reason
	// session links handler helpers to the active session.
	session *Session
}

// Handler executes realm-owned packet behavior.
type Handler func(Context, codec.Packet) error

// HandlerPolicy controls when a packet handler can run.
type HandlerPolicy struct {
	// AllowedStates contains accepted connection states.
	AllowedStates []State
	// RequiresAuthenticated reports whether authentication is required.
	RequiresAuthenticated bool
	// AllowsDisconnected reports whether disposed sessions can be handled.
	AllowsDisconnected bool
}

// HandlerOption configures a handler policy.
type HandlerOption func(*HandlerPolicy)

// HandlerRegistration stores a handler and its policy.
type HandlerRegistration struct {
	// Handler executes packet behavior.
	Handler Handler
	// Policy controls when Handler can run.
	Policy HandlerPolicy
}

// InterceptorNext advances one packet toward the native handler.
type InterceptorNext func() error

// PacketInterceptor observes or cancels one packet before native dispatch.
type PacketInterceptor func(Context, codec.Packet, InterceptorNext) error

// interceptorRegistration stores ordered packet middleware.
type interceptorRegistration struct {
	// interceptor executes plugin or host middleware.
	interceptor PacketInterceptor
	// priority runs larger values first.
	priority int
	// order preserves registration order across equal priorities.
	order uint64
}

// HandlerRegistry stores packet handlers by header.
type HandlerRegistry struct {
	// mutex protects handler registration state.
	mutex sync.RWMutex
	// handlers stores exact packet handlers by header.
	handlers map[uint16]HandlerRegistration
	// fallback stores the handler used when no exact header matches.
	fallback HandlerRegistration
	// hasFallback reports whether fallback is active.
	hasFallback bool
	// interceptors stores global middleware.
	interceptors []interceptorRegistration
	// headerInterceptors stores header-scoped middleware.
	headerInterceptors map[uint16][]interceptorRegistration
	// interceptorOrder stores the next stable registration sequence.
	interceptorOrder uint64
}

// NewHandlerRegistry creates an empty handler registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{handlers: make(map[uint16]HandlerRegistration), headerInterceptors: make(map[uint16][]interceptorRegistration)}
}

// Register adds a handler for a packet header.
func (registry *HandlerRegistry) Register(header uint16, handler Handler, opts ...HandlerOption) error {
	if handler == nil {
		return ErrInvalidHandler
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if _, exists := registry.handlers[header]; exists {
		return ErrHandlerExists
	}

	registry.handlers[header] = HandlerRegistration{Handler: handler, Policy: NewHandlerPolicy(opts...)}

	return nil
}

// Unregister removes a handler for a packet header.
func (registry *HandlerRegistry) Unregister(header uint16) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if _, exists := registry.handlers[header]; !exists {
		return false
	}

	delete(registry.handlers, header)

	return true
}

// SetFallback changes the handler used when no header handler is registered.
func (registry *HandlerRegistry) SetFallback(handler Handler, opts ...HandlerOption) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.fallback = HandlerRegistration{Handler: handler, Policy: NewHandlerPolicy(opts...)}
	registry.hasFallback = handler != nil
}

// Intercept registers prioritized packet middleware for every or one header.
func (registry *HandlerRegistry) Intercept(header *uint16, priority int, interceptor PacketInterceptor) error {
	if interceptor == nil {
		return ErrInvalidHandler
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	registry.interceptorOrder++
	registration := interceptorRegistration{interceptor: interceptor, priority: priority, order: registry.interceptorOrder}
	if header == nil {
		registry.interceptors = appendSortedInterceptor(registry.interceptors, registration)
		return nil
	}
	registry.headerInterceptors[*header] = appendSortedInterceptor(registry.headerInterceptors[*header], registration)

	return nil
}

// Handle routes a packet to the matching handler.
func (registry *HandlerRegistry) Handle(context Context, packet codec.Packet) error {
	registry.mutex.RLock()
	registration, ok := registry.handlers[packet.Header]
	if !ok && registry.hasFallback {
		registration = registry.fallback
		ok = true
	}
	interceptors := mergeInterceptors(registry.interceptors, registry.headerInterceptors[packet.Header])
	registry.mutex.RUnlock()

	if ok && registration.Handler != nil && !registration.Policy.Allows(context) {
		return ErrHandlerPolicy
	}
	if len(interceptors) == 0 {
		if !ok || registration.Handler == nil {
			return ErrHandlerNotFound
		}
		return registration.Handler(context, packet)
	}

	native := func() error {
		if !ok || registration.Handler == nil {
			return ErrHandlerNotFound
		}
		return registration.Handler(context, packet)
	}

	return runInterceptorChain(interceptors, context, packet, native)
}

// appendSortedInterceptor adds one registration in stable priority order.
func appendSortedInterceptor(entries []interceptorRegistration, registration interceptorRegistration) []interceptorRegistration {
	entries = append(entries, registration)
	slices.SortStableFunc(entries, func(left interceptorRegistration, right interceptorRegistration) int {
		if order := cmp.Compare(right.priority, left.priority); order != 0 {
			return order
		}
		return cmp.Compare(left.order, right.order)
	})

	return entries
}

// mergeInterceptors returns global and scoped middleware in stable priority order.
func mergeInterceptors(global []interceptorRegistration, scoped []interceptorRegistration) []interceptorRegistration {
	if len(global) == 0 {
		return append([]interceptorRegistration(nil), scoped...)
	}
	if len(scoped) == 0 {
		return append([]interceptorRegistration(nil), global...)
	}
	merged := make([]interceptorRegistration, 0, len(global)+len(scoped))
	merged = append(merged, global...)
	merged = append(merged, scoped...)
	slices.SortStableFunc(merged, func(left interceptorRegistration, right interceptorRegistration) int {
		if order := cmp.Compare(right.priority, left.priority); order != 0 {
			return order
		}
		return cmp.Compare(left.order, right.order)
	})

	return merged
}

// runInterceptorChain advances middleware recursively toward native dispatch.
func runInterceptorChain(entries []interceptorRegistration, context Context, packet codec.Packet, native InterceptorNext) error {
	var invoke func(int) error
	invoke = func(index int) error {
		if index == len(entries) {
			return native()
		}
		return entries[index].interceptor(context, packet, func() error { return invoke(index + 1) })
	}

	return invoke(0)
}

// Len returns the number of registered header handlers.
func (registry *HandlerRegistry) Len() int {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	return len(registry.handlers)
}

// context returns an immutable handler context snapshot.
func (session *Session) context(direction Direction) Context {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	context := session.contextLocked()
	context.Direction = direction

	return context
}

// contextLocked returns a context snapshot while the session mutex is held.
func (session *Session) contextLocked() Context {
	return Context{
		ConnectionID:     session.id,
		ConnectionKind:   session.kind,
		State:            session.state,
		StartedAt:        session.startedAt,
		RemoteAddr:       session.remoteAddr,
		AuthenticatedAt:  session.authenticatedAt,
		Authenticated:    session.authenticated,
		Disconnected:     session.disconnected,
		DisconnectReason: session.disconnectReason,
		session:          session,
	}
}

// NewHandlerPolicy creates a handler policy from options.
func NewHandlerPolicy(opts ...HandlerOption) HandlerPolicy {
	policy := HandlerPolicy{AllowedStates: []State{StateConnected}, RequiresAuthenticated: true}
	for _, opt := range opts {
		opt(&policy)
	}

	return policy
}

// AllowStates allows a handler in the given connection states.
func AllowStates(states ...State) HandlerOption {
	return func(policy *HandlerPolicy) {
		policy.AllowedStates = append([]State(nil), states...)
	}
}

// AllowAnyActiveState allows a handler in any non-terminal state.
func AllowAnyActiveState() HandlerOption {
	return AllowStates(StateCreated, StateHandshaking, StateSecuring, StateAuthenticating, StateAuthenticated, StateConnected)
}

// AllowUnauthenticated allows a handler before authentication.
func AllowUnauthenticated() HandlerOption {
	return func(policy *HandlerPolicy) {
		policy.RequiresAuthenticated = false
	}
}

// AllowDisconnected allows a handler after disposal starts.
func AllowDisconnected() HandlerOption {
	return func(policy *HandlerPolicy) {
		policy.AllowsDisconnected = true
	}
}

// Allows reports whether a context can run under the policy.
func (policy HandlerPolicy) Allows(context Context) bool {
	if context.Disconnected && !policy.AllowsDisconnected {
		return false
	}
	if policy.RequiresAuthenticated && !context.Authenticated {
		return false
	}
	if slices.Contains(policy.AllowedStates, context.State) {
		return true
	}

	return len(policy.AllowedStates) == 0
}
