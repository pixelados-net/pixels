package bus

import (
	"context"
	"sync"
	"time"

	gookitevent "github.com/gookit/event"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	// contextKey stores the publish context inside the third-party event.
	contextKey = "_pixels_context"

	// eventKey stores the Pixels event inside the third-party event.
	eventKey = "_pixels_event"
)

// Module provides the local event bus.
var Module = fx.Module(
	"event-bus",
	fx.Provide(
		NewWithLogger,
		NewPublisher,
		NewSubscriber,
	),
)

// Handler handles one local event.
type Handler func(context.Context, Event) error

// Publisher publishes local events.
type Publisher interface {
	// Publish publishes an event to local subscribers.
	Publish(context.Context, Event) error
}

// Subscriber subscribes to local events.
type Subscriber interface {
	// Subscribe registers a persistent handler with priority.
	Subscribe(Name, int, Handler) (*Subscription, error)

	// SubscribeOnce registers a one-shot handler with priority.
	SubscribeOnce(Name, int, Handler) (*Subscription, error)
}

// Bus publishes events to prioritized local listeners.
type Bus struct {
	// manager stores the underlying event bus implementation.
	manager *gookitevent.Manager
	// log records published events.
	log *zap.Logger
}

// New creates a local event bus.
func New() *Bus {
	return NewWithLogger(zap.NewNop())
}

// NewWithLogger creates a local event bus with structured logging.
func NewWithLogger(log *zap.Logger) *Bus {
	if log == nil {
		log = zap.NewNop()
	}

	return &Bus{manager: gookitevent.NewManager("pixels"), log: log}
}

// NewPublisher exposes the bus through its publish contract.
func NewPublisher(local *Bus) Publisher {
	return local
}

// NewSubscriber exposes the bus through its subscription contract.
func NewSubscriber(local *Bus) Subscriber {
	return local
}

// Publish publishes an event to subscribers.
func (bus *Bus) Publish(ctx context.Context, event Event) error {
	if !event.Valid() {
		return ErrInvalidEvent
	}

	if ctx == nil {
		ctx = context.Background()
	}

	event = event.WithTime(time.Now())
	bus.log.Debug("event published",
		zap.String("event_name", string(event.Name)),
		zap.Time("event_at", event.At),
		zap.Any("event_payload", event.Payload),
	)

	return bus.manager.FireEventCtx(ctx, gookitevent.New(string(event.Name), gookitevent.M{
		contextKey: ctx,
		eventKey:   event,
	}))
}

// Subscribe registers a persistent handler with priority.
func (bus *Bus) Subscribe(name Name, priority int, handler Handler) (*Subscription, error) {
	return bus.subscribe(name, priority, handler, false)
}

// SubscribeOnce registers a one-shot handler with priority.
func (bus *Bus) SubscribeOnce(name Name, priority int, handler Handler) (*Subscription, error) {
	return bus.subscribe(name, priority, handler, true)
}

// Close releases bus resources.
func (bus *Bus) Close() error {
	return bus.manager.Close()
}

// subscriptionListener adapts a Pixels handler to gookit/event.
type subscriptionListener struct {
	// handler handles the local event.
	handler Handler
}

// Handle handles one third-party event.
func (listener subscriptionListener) Handle(raw gookitevent.Event) error {
	ctx, ok := raw.Get(contextKey).(context.Context)
	if !ok || ctx == nil {
		ctx = context.Background()
	}

	event, ok := raw.Get(eventKey).(Event)
	if !ok || !event.Valid() {
		return ErrInvalidEvent
	}

	return listener.handler(ctx, event)
}

// subscribe registers a handler.
func (bus *Bus) subscribe(name Name, priority int, handler Handler, once bool) (*Subscription, error) {
	if name == "" {
		return nil, ErrInvalidEvent
	}
	if handler == nil {
		return nil, ErrInvalidHandler
	}

	listener := &subscriptionListener{handler: handler}
	if once {
		bus.manager.Once(string(name), listener, priority)
	} else {
		bus.manager.On(string(name), listener, priority)
	}

	return &Subscription{manager: bus.manager, name: name, listener: listener}, nil
}

// Subscription describes one event subscription.
type Subscription struct {
	// mutex protects subscription lifecycle.
	mutex sync.Mutex
	// manager removes the listener.
	manager *gookitevent.Manager
	// name stores the subscribed event name.
	name Name
	// listener stores the adapted listener.
	listener *subscriptionListener
	// closed reports whether the listener was removed.
	closed bool
}

// Unsubscribe removes the event subscription.
func (subscription *Subscription) Unsubscribe() {
	subscription.mutex.Lock()
	defer subscription.mutex.Unlock()

	if subscription.closed {
		return
	}

	subscription.manager.RemoveListener(string(subscription.name), subscription.listener)
	subscription.closed = true
}
