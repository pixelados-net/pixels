# Configuration

Every tunable in Pixels comes from environment variables, every variable has a working default, and every variable is documented in [`.env.example`](https://github.com/niflaot/pixels/blob/main/.env.example) plus [[ENVIRONMENT-VARIABLES]]. This page explains how configuration is structured in code, the two loading patterns you will encounter, and the rules for adding new settings.

## Principles

- **Configuration lives next to what it configures.** Each component has its own small `config.go` (`pkg/redis/config.go`, `internal/realm/crafting/config/config.go`). There is no god-struct listing every setting in the project; small configs are composed into the application only where startup actually needs a single value.
- **Every field has a safe default.** The server must boot with an empty environment for local development. Defaults that are unsafe for production (the published dev API key, the dev SSO key) are called out in [[GETTING-STARTED]].
- **Environment variables are the source of truth.** A local `.env` file is loaded for convenience in development, but variables always win.
- **`.env.example` is maintained in the same change** that adds, renames or removes a variable, with a comment stating the purpose and the default.

## Pattern one: struct tags

Infrastructure packages and several realms declare their config as a struct with `env` tags, parsed by [`caarlos0/env`](https://github.com/caarlos0/env). Redis is the smallest real example:

```go
// Config contains Redis connection settings.
type Config struct {
	// Address is the Redis server address.
	Address string `env:"REDIS_ADDRESS" envDefault:"127.0.0.1:6379"`

	// Username is the Redis ACL username.
	Username string `env:"REDIS_USERNAME" envDefault:""`

	// Password is the Redis password.
	Password string `env:"REDIS_PASSWORD" envDefault:""`

	// Database is the selected Redis database.
	Database int `env:"REDIS_DATABASE" envDefault:"0"`
}

// LoadConfig reads Redis configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
```

When values need cross-field validation or clamping, the struct adds a `Normalize()` method applied after parsing, so an out-of-range value degrades to something sane instead of crashing the boot. The pet realm's policy config works this way: `MaxPerRoom` parses from `PIXELS_PET_MAX_PER_ROOM`, and `Normalize` clamps nonsense back to the default.

## Pattern two: explicit loaders

Some realms (crafting, moderation, group) read their handful of settings with small explicit helpers instead of tags. Same guarantees, different mechanics:

```go
// Load reads crafting environment settings with safe defaults.
func Load() Config {
	return Config{
		Enabled:              envBool("PIXELS_CRAFTING_ENABLED", true),
		RecyclerEnabled:      envBool("PIXELS_CRAFTING_RECYCLER_ENABLED", true),
		RecyclerBatchSize:    envInt("PIXELS_CRAFTING_RECYCLER_BATCH_SIZE", 8),
		RecyclerRarityChance: envChances("PIXELS_CRAFTING_RECYCLER_RARITY_CHANCES", map[int32]int{5: 1000, 4: 100, 3: 20, 2: 5}),
	}
}
```

Each `env*` helper parses and falls back on any error, never panicking and never requiring the variable to be set. This pattern earns its keep when a value has a custom shape; `envChances` above parses a `tier=chance,tier=chance` string into a map and rejects partial garbage by falling back wholesale.

Which pattern a new component should use follows the neighborhood: match `pkg/`-style struct tags for infrastructure clients, and match whichever pattern the sibling realms of your feature already use.

## Naming conventions

Domain settings are prefixed with the project and realm: `PIXELS_CRAFTING_*`, `PIXELS_PET_*`, `PIXELS_CAMERA_*`, `PIXELS_WS_*`. Infrastructure clients keep their upstream-conventional, unprefixed names: `REDIS_ADDRESS`, `SSO_KEY`, `STORAGE_ENDPOINT`, `LOG_LEVEL`. PostgreSQL is the historical exception (`PIXELS_POSTGRES_*`). When you add a block to `.env.example`, follow the file's format: a `#` comment per variable (or tight group) stating what it does and `# Default: <value>`, then the assignment.

## Environment config versus database config

Not everything tunable is an environment variable, and the split is deliberate:

- **Environment config** is for values that change how the process runs and are acceptable to require a restart for: batch sizes, timeouts, feature kill-switches, connection settings.
- **Database-backed settings** are for operational values a hotel operator adjusts while the server runs, such as photo prices or the recycler's enabled flag, stored in small settings tables, edited through the admin API with optimistic locking, and cached in memory with an explicit reload path.

A good test when adding a knob: if support staff might legitimately change it at 2 a.m. without a deploy, it belongs in the database behind an admin route. If only an operator doing capacity planning would touch it, it's an environment variable.

## How config reaches components

Configs are constructed once and injected through fx like any other dependency. A realm's `module.go` provides its loader, and constructors receive the typed struct:

```go
var Module = fx.Module("realm-crafting",
	fx.Provide(craftingconfig.Load, craftingdb.New, NewStore, craftingrecipe.New, …),
	fx.Invoke(RegisterConnectionHandlers),
)
```

Nothing reads `os.Getenv` at request time, and nothing reaches into a global config object. If a service needs a setting, the setting is a field on the config struct that was handed to its constructor, which also makes every test trivially able to run with whatever configuration the test needs.
