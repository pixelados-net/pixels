# Infrastructure

`pkg/` holds the components with no game logic of their own: the pieces realms build on. This page walks through each one, how it's wired into the application, and the conventions you're expected to follow when you use it.

## fx modules: how everything is wired

Pixels has no global state and no init-order guesswork. Every component, from the logger to the deepest realm service, is constructed by [`go.uber.org/fx`](https://github.com/uber-go/fx) from an explicit dependency graph. The convention is uniform: each package that participates exposes a `Module`:

```go
// Module provides Redis storage to an Fx dependency graph.
var Module = fx.Module(
	"redis",
	fx.Provide(New),
)
```

`fx.Provide` registers constructors; `fx.Invoke` runs startup work such as registering packet handlers; `fx.Lifecycle` hooks tie a component to process start and stop. The binary in `cmd/` is little more than the list of modules to assemble. Two practical consequences: any constructor can ask for any provided type just by declaring a parameter, and tests never need the graph at all, since every constructor is an ordinary function you can call with fakes.

## Logging: zap

Logging is structured, through [`zap`](https://github.com/uber-go/zap), configured by `pkg/logger`:

```go
type Config struct {
	// Level is the minimum enabled zap level.
	Level string `env:"LOG_LEVEL" envDefault:"info"`

	// Format is the zap encoder format.
	Format Format `env:"LOG_FORMAT" envDefault:"console"`

	// ToonConsole enables compact agent-friendly console logs.
	ToonConsole bool `env:"TOON_CONSOLE" envDefault:"false"`
}
```

`console` gives human-readable development output, `json` gives structured production output, and `TOON_CONSOLE=true` switches to a compact single-line encoder built for protocol tracing: timestamps and callers are stripped and packet fields normalized, so you can watch the wire conversation scroll by while debugging a client. The `*zap.Logger` is provided once through fx and injected wherever it's needed; components log with typed fields (`zap.Int64("player_id", …)`), never `fmt.Sprintf` into a message. The command dispatcher, the bus, and the connection layer already log dispatches, events, and packet traffic, so a feature usually needs to add logging only for its own domain decisions.

## PostgreSQL: pool, scoped transactions, health

`pkg/postgres` wraps [`pgx`](https://github.com/jackc/pgx) with the project's lifecycle conventions. `NewPool` parses the DSN, applies pool limits and a per-connection statement timeout, and registers fx hooks so the process pings the database on start (failing the boot loudly if it's unreachable) and closes the pool on stop.

The part you'll actually interact with is the **scoped transaction** mechanism. Repositories embed the pattern:

```go
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, found := postgres.ScopedExecutor(ctx); found {
		return work(ctx) // already inside a transaction: join it
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}
```

`WithinScope` opens a transaction and stashes its executor in the context; every repository method resolves its executor through `ExecutorFor`, which returns the transaction's executor if one is in flight and the plain pool otherwise. The payoff is composability: a service can wrap calls into *several* repositories in one `WithinTransaction`, and each nested call transparently joins the same transaction. That's how a craft consumes ingredients, mints the reward, and debits currency atomically across three realms' repositories without any of them knowing about the others.

Queries themselves are plain SQL string constants with positional parameters and hand-rolled `rows.Scan`, no ORM. Concurrency-sensitive updates encode their guard in the SQL (`update … set remaining = remaining - 1 where … and remaining > 0 returning …`), so a race resolves in the database rather than in application locks.

## Redis

`pkg/redis` wraps the go-redis client with a deliberately small, generic surface: `Find` (get with a `found bool` instead of a sentinel error), `Set`, `SetIfAbsent`, `Take` (atomic get-and-delete, the primitive behind one-time SSO tickets), `Increment` (atomic counter with expiry-on-first-use, the primitive behind flood control), `Delete`, `Expire`. There's no per-realm Redis wrapper layer; call sites use these operations directly with their own namespaced keys. If you find yourself wanting a Redis method that isn't there, add a generic operation, not a feature-specific one.

## Object storage

`pkg/storage` wraps an S3-compatible client (MinIO's Go SDK, so any S3-compatible provider works) behind three operations: `Put` with an internally enforced upload timeout, `Delete`, and `PublicURL`. It verifies the bucket on process start, the same fail-fast philosophy as the PostgreSQL pool: a misconfigured bucket fails the deploy, not the first photo upload three hours later. One rule worth knowing from the camera realm's design: content that ends up referenced durably (a photo hanging on a wall) always uses permanent public URLs, never presigned ones, because a presigned URL expiring weeks later would silently break every furniture item pointing at it.

## Event bus

`pkg/bus` is the in-process publish/subscribe fabric described in depth in [[EVENTS]]. From an infrastructure standpoint: it's provided as separate `Publisher` and `Subscriber` interfaces so components declare only the capability they use, and every publish is logged with its event name for traceability.

## i18n

`pkg/i18n` loads an immutable translation catalog from JSON at startup and exposes a `Translator`:

```go
// Translator resolves localized text.
type Translator interface {
	T(Locale, Key, ...Params) string
	Default(Key, ...Params) string
	Entries(Locale) map[Key]string
}
```

Keys are stable and namespaced (`moderation.report.received`), params interpolate into the resolved string, and lookups fall back from the requested locale to the configured fallback locale, with missing-key behavior configurable between echoing the key (development-friendly) and returning empty. `Entries` exists so the HTTP layer can serve whole localized text bundles to the Nitro client. The rule that makes this layer work is in [[PROJECTIONS]]: text is resolved before encoding, so no packet or realm re-implements translation.

## HTTP

`pkg/http` hosts the [Fiber](https://github.com/gofiber/fiber)-based HTTP server that serves everything that isn't the game protocol: the public `/status` and `/ws` endpoints, Nitro client configuration resources, and the private administrative API. Private routes sit behind a single `X-API-Key` middleware checked against `PIXELS_ACCESS_KEY`, and each realm contributes its own route group (`pkg/http/<realm>/routes`) with a uniform pattern: an fx-injected `Dependencies` struct, per-feature handler files, permission-node authorization for the acting staff member, and an audit trail (`actorPlayerId` + `reason`) on every mutation. Every route is described in `pkg/http/openapi`, which is what `GET /docs` renders interactively in development. If you add an admin route without documenting it there, review will send you back.

## Build metadata

`pkg/build` carries the project name, semantic version, and commit hash injected at build time through `-ldflags`. The running server reports the semantic version, such as `v0.0.1`, in `/status`. The startup log reports that version and the short commit separately, so operators can compare the public release identity and the exact source revision without producing a noncanonical version string.
