# Plugin Listeners and Interceptors

Pixels exposes two callback pipelines to plugins. Event listeners react to typed domain moments chosen by the host. Packet interceptors sit before native inbound packet handlers. They share priority, timeout, panic recovery, and plugin scope rules, but they are not interchangeable.

## Event listeners

A listener is registered by stable event name and receives an `sdk/event.Event`. SDK 1.x exposes these events:

| Event | Type | Moment |
|---|---|---|
| `player.connected` | Notification | After authentication has produced a live player |
| `chat.send` | Cancellable and mutable | After the word filter and before WIRED or room delivery |

The event name selects the stream. The concrete type provides the data.

```go
err := host.Events().Listen(
	sdkevent.PlayerConnectedName,
	sdkevent.ListenerOptions{Priority: sdkplugin.PriorityNormal},
	func(ctx context.Context, current sdkevent.Event) error {
		connected, valid := current.(*sdkevent.PlayerConnected)
		if !valid {
			return nil
		}

		return host.Players().Message(connected.Player.ID, "Plugin ready")
	},
)
```

The internal realm bus is never exposed. A plugin can subscribe only to events the SDK deliberately projects and cannot publish arbitrary realm events.

## Mutable and cancellable chat

`chat.send` carries sanitized text. A listener may replace `Text`, cancel delivery, or do both.

```go
err := host.Events().Listen(
	sdkevent.ChatSendName,
	sdkevent.ListenerOptions{Priority: sdkplugin.PriorityHigh},
	func(_ context.Context, current sdkevent.Event) error {
		chat, valid := current.(*sdkevent.ChatSend)
		if !valid {
			return nil
		}

		if strings.EqualFold(chat.Text, "blocked phrase") {
			chat.SetCancelled(true)
			return nil
		}

		chat.Text = strings.ReplaceAll(chat.Text, "hello", "hi")
		return nil
	},
)
```

Each callback receives its own event copy. Pixels applies that copy to the shared result only when the listener returns successfully. This prevents a timed out callback from changing chat later from a goroutine that outlived its deadline.

`IgnoreCancelled` skips a listener when an earlier callback has already cancelled the event. Leave it false when the listener must observe cancellation or may restore it.

## Priority order

Larger numeric values execute first. Registrations with the same priority keep registration order.

| Constant | Value | Intended use |
|---|---:|---|
| `PriorityHighest` | `200` | Earliest policy guard |
| `PriorityHigh` | `100` | Early mutation or validation |
| `PriorityNormal` | `0` | Ordinary behavior |
| `PriorityLow` | `-100` | Late behavior |
| `PriorityLowest` | `-200` | Final mutation |
| `PriorityMonitor` | `-1000` | Last, observation only |

A monitor should not mutate or uncancel an event. Its low value exists so it sees the final state produced by ordinary listeners.

## Packet interceptors

An interceptor receives an immutable player snapshot when authenticated, the inbound header, and a private copy of the encoded payload. A nil `Header` observes every inbound packet. A header pointer limits it to that one packet.

```go
err := host.Players().Intercept(
	func(ctx context.Context, packet sdkplugin.InterceptContext, next sdkplugin.Next) error {
		log.Printf("header=%d player=%d", packet.Header, packet.Player.ID)
		return next(ctx)
	},
	sdkplugin.InterceptOptions{Priority: sdkplugin.PriorityLow},
)
```

Calling `next(ctx)` advances to the next interceptor and eventually the native handler. Returning `nil` without calling `next` deliberately consumes the packet. Calling `next` more than once, after the callback has returned, or with a cancelled context is rejected.

An interceptor should not return an ordinary error to express a harmless cancellation. That error reaches the transport's protocol failure path. Consume intentionally with `return nil`, or continue with `return next(ctx)`.

## Isolation and deadlines

Every callback receives a context bounded by `PIXELS_PLUGIN_CALLBACK_TIMEOUT`. Code should stop work when `ctx.Done()` closes. Go cannot forcibly terminate a goroutine, so a plugin that ignores cancellation can still waste resources after the caller has moved on.

A recovered panic disables the complete plugin scope. Future listeners, commands, interceptors, and routes from that plugin stop running. A normal returned error is logged for listeners and does not disable the scope. A timed out listener has its late mutations discarded. A panicking or timed out interceptor that has not advanced the chain falls through to native handling so infrastructure failure does not silently swallow the player's packet.

Plugins execute inside the Pixels process. Recovery and deadlines isolate callbacks operationally, but they do not form a security sandbox. Only install binaries you trust.
