# Production Setup

This page turns a local Pixels checkout into an operable deployment. It covers the current production blocker, release images, infrastructure, migrations, secrets, proxying, storage, monitoring, backups, and rollback. [[ENVIRONMENT-VARIABLES]] is the complete environment reference.

## Current production readiness

Pixels can build and run a production container, but the normal Nitro login flow has one explicit blocker. `PIXELS_ENV=production` requires a session `SecureChannel` before accepting an SSO ticket, while the in protocol Diffie Hellman provider is not implemented yet. TLS at a reverse proxy protects the WebSocket transport but does not mark that session channel ready.

Do not weaken the policy silently or claim a public deployment is complete. Before exposing a hotel to users, implement and test the Diffie provider plus channel activation, or make a reviewed architectural change that lets an authenticated TLS transport satisfy the policy. [[AUTH-SECURITY]] documents the exact boundary.

Everything below remains the required operational setup once that blocker is resolved. It is also useful for a private staging environment where the security limitation is understood.

## Recommended topology

```text
Nitro and CMS
  HTTPS and WSS
  reverse proxy or load balancer
  Pixels container
    PostgreSQL
    Redis
    S3 compatible object storage
```

Expose only the reverse proxy publicly. Keep PostgreSQL and Redis on a private network. Restrict the S3 API to Pixels and expose only the durable public object URL required for camera content.

Pixels currently owns live rooms, players, timers, and connection bindings in one process. PostgreSQL and Redis are shared infrastructure, but they do not coordinate a room simulation across several Pixels replicas. Start with one active application instance. Treat horizontal application scaling as an architectural change, not a deployment toggle.

## Release image contract

GitHub Actions publishes `ghcr.io/niflaot/pixels` only for a valid semantic tag such as `v0.0.1` whose commit belongs to `main`. The tag must match `pkg/build.Version`. Validation, migrations, tests, binary builds, and the multi architecture image must all succeed before the Discord webhook runs.

The image receives both release version and commit hash at build time. `GET /status` exposes the semantic version and the startup log includes both version and short commit.

Release steps:

1. Change `pkg/build.Version` to the intended `vMAJOR.MINOR.PATCH` value.
2. Merge the release commit into `main` and wait for CI.
3. Create the tag on that exact main commit.
4. Push the tag.
5. Verify the Package workflow published the expected immutable version tag before deploying.

```sh
git switch main
git pull --ff-only
git tag -a v0.0.1 -m "Pixels v0.0.1"
git push origin v0.0.1
```

Deploy an immutable tag such as `ghcr.io/niflaot/pixels:v0.0.1`. Do not deploy `latest` when rollback accuracy matters.

## Secrets and environment

Store secrets in the deployment platform, not in `.env` committed to Git. At minimum replace these published development values:

| Variable | Production requirement |
|---|---|
| `PIXELS_ACCESS_KEY` | Long random value shared only with trusted CMS and administration clients |
| `SSO_KEY` | Independent long random HMAC key |
| `PIXELS_POSTGRES_PASSWORD` | Dedicated database credential |
| `REDIS_PASSWORD` | Redis ACL credential on a private network |
| `STORAGE_ACCESS_KEY` | Dedicated bucket access identity |
| `STORAGE_SECRET_KEY` | Dedicated bucket secret |
| `PIXELS_PROGRESSION_DAILY_POOL_SEED` | Stable secret salt if deterministic quest selection must not be predictable |

Use different values per environment. Rotate the API and SSO keys with a coordinated CMS and Pixels deployment. Rotating `SSO_KEY` invalidates outstanding tickets, which is normally desirable during a security rotation.

Generate secrets with an operating system cryptographic generator:

```sh
openssl rand -base64 48
```

## PostgreSQL and migrations

Create a dedicated database and role. Set `PIXELS_POSTGRES_SSL_MODE` to the verification mode required by the provider instead of `disable`. Tune pool limits against the database connection budget and the number of application instances.

Run Liquibase as a separate deployment step before starting the new image:

```sh
docker run --rm \
  -v "$PWD:/workspace" \
  -w /workspace \
  liquibase/liquibase:4.31 \
  --defaults-file=database/liquibase.example.properties validate

docker run --rm \
  -v "$PWD:/workspace" \
  -w /workspace \
  liquibase/liquibase:4.31 \
  --defaults-file=database/liquibase.example.properties update
```

Create a production specific Liquibase properties file or pass its connection values as secret arguments. Never run `database/liquibase.seed.example.properties` or the `development` context against production.

Take a tested backup before every schema change. A successful Liquibase rollback definition is not a replacement for a recoverable database backup.

## Redis

Redis stores one time SSO tickets, shared throttles, and cache fragments. Use a dedicated ACL user and key prefix. Configure persistence and high availability according to how much temporary state loss the hotel can tolerate.

The current Redis client accepts an address, username, password, and database number but does not expose TLS settings. Keep it on a trusted private network until Redis TLS configuration is implemented. Do not publish port 6379 to the internet.

## Object storage

The camera subsystem validates its bucket during startup and fails fast when storage is unusable. Use HTTPS with `STORAGE_USE_SSL=true`. `STORAGE_PUBLIC_BASE_URL` must be a permanent public origin because photo furniture stores durable URLs.

`STORAGE_PUBLIC_READ=true` applies a public read policy to the bucket. If the provider manages policy outside the application, verify the resulting objects are still permanently readable. Use a dedicated bucket and lifecycle monitoring. Do not apply an object expiration rule that deletes referenced photos.

## Reverse proxy and WebSocket

Terminate TLS and expose `wss://` for `/ws`. Preserve binary frames, disable response buffering for upgraded connections, and set proxy idle timeouts above the Pixels read and pong windows. Forward ordinary HTTP routes to the same process.

The current Fiber app does not configure trusted proxy ranges. `RemoteAddr` may therefore represent the proxy rather than the original client. Do not bind SSO tickets to client IP in production until trusted proxy handling is explicitly configured and tested. The ticket itself remains single use without IP binding.

Keep private API routes behind network policy in addition to `X-API-Key`. `/status`, `/ws`, and client configuration resources are public. `/docs` is hidden when `PIXELS_ENV` is not `development`.

## Runtime configuration

Use `LOG_FORMAT=json` for structured collection and keep `TOON_CONSOLE=false`. Size WebSocket queues and PostgreSQL pools from measured load. Feature defaults are starting values, not capacity guarantees.

Native plugins execute in the server process. Build every `.so` with the same release source, Go toolchain, operating system, architecture, and shared dependencies as the container. Mount plugin objects read only and restart the process to load changes. A plugin deployment must be rolled back together with its compatible host image.

## Health and observability

Probe `GET /status` and require both HTTP success and the expected semantic version. The endpoint proves the HTTP process is serving, while startup already fails on PostgreSQL and object storage connectivity. Add external probes for Redis and the public camera origin.

Collect structured logs, container restarts, memory, CPU, open WebSocket count, database pool saturation, Redis latency, object storage failures, room tick delays, and disconnect reasons. Alert on repeated protocol errors separately from transport disconnects so client incompatibility is not mistaken for infrastructure loss.

## Backup and rollback

Back up PostgreSQL and version the object storage bucket. Redis backup is useful for operational continuity but never replaces PostgreSQL. Test restoration on an isolated environment.

A rollback should use the previous immutable image and its compatible plugin set. Database rollback requires an explicit reviewed plan because a prior binary may not understand a newer schema. Prefer forward compatible migrations and a forward fix when data has already been written by the new version.

## Production checklist

| Check | Required state |
|---|---|
| Session security | Diffie or an approved secure transport integration is implemented and tested |
| Release | Semantic tag matches source version and points to `main` |
| Image | Immutable GHCR tag verified before deployment |
| Secrets | Every published development value replaced |
| Database | TLS, backup, migration validation, and update completed |
| Seeds | Development seed changelog not executed |
| Redis | Private network, ACL, persistence, and monitoring configured |
| Storage | HTTPS, permanent public URL, bucket policy, and backup configured |
| Proxy | HTTPS, WSS, binary upgrade, and timeouts tested |
| Scale | One active Pixels simulation instance unless clustering is implemented |
| Observability | Status, logs, metrics, disconnect reasons, and alerts connected |
| Rollback | Previous image, plugins, database plan, and restore procedure tested |
