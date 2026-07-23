package connection

// State names the lifecycle phase of a connection.
type State uint8

const (
	// StateCreated means the session exists before protocol traffic.
	StateCreated State = iota + 1

	// StateHandshaking means metadata or crypto negotiation is active.
	StateHandshaking

	// StateSecuring means key exchange work is active.
	StateSecuring

	// StateAuthenticating means identity proof is being validated.
	StateAuthenticating

	// StateAuthenticated means identity proof succeeded.
	StateAuthenticated

	// StateConnected means normal session traffic is allowed.
	StateConnected

	// StateClosing means disposal has started.
	StateClosing

	// StateClosed means disposal finished.
	StateClosed

	// StateError means protocol, transport, security, or policy failed.
	StateError
)

// String returns a stable state label.
func (state State) String() string {
	switch state {
	case StateCreated:
		return "created"
	case StateHandshaking:
		return "handshaking"
	case StateSecuring:
		return "securing"
	case StateAuthenticating:
		return "authenticating"
	case StateAuthenticated:
		return "authenticated"
	case StateConnected:
		return "connected"
	case StateClosing:
		return "closing"
	case StateClosed:
		return "closed"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// Event names a lifecycle transition trigger.
type Event string

const (
	// EventPacketReceived begins protocol handling.
	EventPacketReceived Event = "packet_received"

	// EventDiffieRequested begins security negotiation.
	EventDiffieRequested Event = "diffie_requested"

	// EventDiffieCompleted completes security negotiation.
	EventDiffieCompleted Event = "diffie_completed"

	// EventAuthenticationStarted begins authentication.
	EventAuthenticationStarted Event = "authentication_started"

	// EventAuthenticationAccepted completes authentication.
	EventAuthenticationAccepted Event = "authentication_accepted"

	// EventAuthenticationRejected rejects authentication.
	EventAuthenticationRejected Event = "authentication_rejected"

	// EventSessionReady marks the session connected.
	EventSessionReady Event = "session_ready"

	// EventDisconnectRequested starts graceful disposal.
	EventDisconnectRequested Event = "disconnect_requested"

	// EventDisconnectCompleted completes disposal.
	EventDisconnectCompleted Event = "disconnect_completed"

	// EventProtocolFailed records protocol failure.
	EventProtocolFailed Event = "protocol_failed"

	// EventTransportFailed records transport failure.
	EventTransportFailed Event = "transport_failed"
)

// Transition describes one valid lifecycle movement.
type Transition struct {
	// From is the starting state.
	From State
	// Event is the transition trigger.
	Event Event
	// To is the resulting state.
	To State
}

// State returns the connection lifecycle phase.
func (session *Session) State() State {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.state
}

// Transition applies a lifecycle event.
func (session *Session) Transition(event Event) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.disconnected && event != EventDisconnectCompleted {
		return ErrDisposed
	}

	next, ok := nextState(session.state, event)
	if !ok {
		return ErrInvalidTransition
	}

	session.state = next

	return nil
}

// nextState returns the state reached by an event.
func nextState(state State, event Event) (State, bool) {
	switch event {
	case EventPacketReceived:
		return transitionFrom(state, StateCreated, StateHandshaking)
	case EventDiffieRequested:
		return transitionFrom(state, StateHandshaking, StateSecuring)
	case EventDiffieCompleted:
		return transitionFrom(state, StateSecuring, StateHandshaking)
	case EventAuthenticationStarted:
		return transitionFrom(state, StateHandshaking, StateAuthenticating)
	case EventAuthenticationAccepted:
		return transitionFrom(state, StateAuthenticating, StateAuthenticated)
	case EventAuthenticationRejected:
		return transitionFrom(state, StateAuthenticating, StateError)
	case EventSessionReady:
		return transitionFrom(state, StateAuthenticated, StateConnected)
	case EventDisconnectRequested:
		return closeState(state)
	case EventDisconnectCompleted:
		return transitionFrom(state, StateClosing, StateClosed)
	case EventProtocolFailed, EventTransportFailed:
		return failState(state)
	default:
		return 0, false
	}
}

// canTransition reports whether an event can be applied.
func canTransition(state State, event Event) bool {
	_, ok := nextState(state, event)

	return ok
}

// markTraffic records traffic and applies the first packet transition.
func (session *Session) markTraffic(event Event) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.disconnected {
		return ErrDisposed
	}

	session.trafficStarted = true
	if event == "" || session.state != StateCreated {
		return nil
	}

	next, ok := nextState(session.state, event)
	if !ok {
		return ErrInvalidTransition
	}

	session.state = next

	return nil
}

// transitionFrom applies a transition from one exact state.
func transitionFrom(state State, from State, to State) (State, bool) {
	if state != from {
		return 0, false
	}

	return to, true
}

// closeState returns the closing state for active or errored states.
func closeState(state State) (State, bool) {
	switch state {
	case StateCreated, StateHandshaking, StateSecuring, StateAuthenticating, StateAuthenticated, StateConnected, StateError:
		return StateClosing, true
	default:
		return 0, false
	}
}

// failState returns the error state for active states.
func failState(state State) (State, bool) {
	switch state {
	case StateCreated, StateHandshaking, StateSecuring, StateAuthenticating, StateAuthenticated, StateConnected:
		return StateError, true
	default:
		return 0, false
	}
}
