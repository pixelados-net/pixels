# Environment Variables

This is the complete runtime environment reference for Pixels. Every setting has a default so local development can start from `.env.example`. Published development credentials are not safe for production. Operational guidance lives in [[PRODUCTION-SETUP]].

Durations use Go duration syntax such as `250ms`, `5m`, `24h`, and `168h`. Empty means the default value is an empty string.

## Process, logging, localization, and extensions

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_ENV` | `development` | Runtime environment and security posture |
| `PIXELS_HOST` | `127.0.0.1` | HTTP and protocol bind host |
| `PIXELS_PORT` | `3000` | HTTP and protocol bind port |
| `PIXELS_ACCESS_KEY` | `pixels-development-access-key-change-me` | Key required in `X-API-Key` for private routes |
| `LOG_LEVEL` | `info` | Zap minimum log level |
| `LOG_FORMAT` | `console` | Zap encoder, either console or JSON |
| `TOON_CONSOLE` | `false` | Compact protocol tracing format |
| `PIXELS_I18N_PATH` | `i18n/translations.json` | Translation catalog path |
| `PIXELS_I18N_DEFAULT_LOCALE` | `es` | Locale used without a player preference |
| `PIXELS_I18N_FALLBACK_LOCALE` | `en` | Locale used for missing keys |
| `PIXELS_I18N_MISSING_MODE` | `key` | Missing key behavior, either key or empty |
| `PIXELS_CURRENCY_TYPES` | `-1:credits,0:duckets,5:diamonds` | Currency ids and localization suffixes |
| `PIXELS_CURRENCY_LEDGER_TYPES` | `-1` | Currency ids that write audit ledger rows |
| `PIXELS_PLUGIN_DIRECTORY` | `plugins` | Root containing one folder per native plugin |
| `PIXELS_PLUGIN_CALLBACK_TIMEOUT` | `2s` | Deadline for plugin callbacks |
| `PIXELS_COMMAND_PREFIX` | `:` | Room chat prefix for plugin commands |

## PostgreSQL, Redis, SSO, and storage

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `PIXELS_POSTGRES_PORT` | `5432` | PostgreSQL port |
| `PIXELS_POSTGRES_DATABASE` | `pixels` | PostgreSQL database |
| `PIXELS_POSTGRES_USER` | `pixels` | PostgreSQL user |
| `PIXELS_POSTGRES_PASSWORD` | `pixels` | PostgreSQL password |
| `PIXELS_POSTGRES_SSL_MODE` | `disable` | PostgreSQL SSL mode |
| `PIXELS_POSTGRES_MAX_CONNS` | `10` | Maximum pool connections |
| `PIXELS_POSTGRES_MIN_CONNS` | `1` | Minimum pool connections |
| `PIXELS_POSTGRES_CONNECT_TIMEOUT` | `5s` | Connection creation timeout |
| `PIXELS_POSTGRES_STATEMENT_TIMEOUT` | `5s` | Per connection statement timeout |
| `PIXELS_POSTGRES_HEALTH_TIMEOUT` | `2s` | Startup health ping timeout |
| `REDIS_ADDRESS` | `127.0.0.1:6379` | Redis host and port |
| `REDIS_USERNAME` | empty | Redis ACL username |
| `REDIS_PASSWORD` | empty | Redis ACL password |
| `REDIS_DATABASE` | `0` | Redis database number |
| `SSO_DEFAULT_TTL` | `5m` | Default one time ticket lifetime |
| `SSO_KEY` | `pixels-development-sso-key-change-me` | HMAC key used to derive ticket storage keys |
| `SSO_PREFIX` | `pixels:sso` | Redis namespace for tickets |
| `STORAGE_ENDPOINT` | `127.0.0.1:9000` | S3 compatible endpoint without scheme |
| `STORAGE_PUBLIC_BASE_URL` | empty | Permanent public bucket origin override |
| `STORAGE_ACCESS_KEY` | empty | S3 access key |
| `STORAGE_SECRET_KEY` | empty | S3 secret key |
| `STORAGE_BUCKET` | `pixels-camera` | Camera object bucket |
| `STORAGE_USE_SSL` | `true` | Use HTTPS for S3 operations |
| `STORAGE_PUBLIC_READ` | `true` | Apply public read bucket policy |
| `STORAGE_UPLOAD_TIMEOUT` | `10s` | Upload and delete request timeout |

## WebSocket and connection behavior

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_WS_QUEUE_SIZE` | `256` | Outbound packet queue capacity per connection |
| `PIXELS_WS_WRITE_TIMEOUT` | `5s` | WebSocket write deadline |
| `PIXELS_WS_READ_TIMEOUT` | `75s` | WebSocket read deadline |
| `PIXELS_WS_PING_INTERVAL` | `30s` | Protocol ping cadence |
| `PIXELS_WS_PONG_TIMEOUT` | `60s` | Maximum time without a pong |
| `PIXELS_WS_CLOSE_GRACE` | `2s` | Graceful close queue deadline |

## Rooms, entry, moderation, and avatars

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_ROOM_IDLE_TIMEOUT` | `5m` | Inactivity before an avatar becomes AFK |
| `PIXELS_ROOM_IDLE_SWEEP_INTERVAL` | `1s` | Global AFK reconciliation cadence |
| `PIXELS_ROOM_ACTION_TRANSITION_DELAY` | `100ms` | Delay between replacing avatar actions |
| `PIXELS_ROOM_ENTRY_HANGOUT_TIMEOUT` | `5m` | Doorbell approval lifetime |
| `PIXELS_ROOM_ENTRY_MAX_PASSWORD_ATTEMPTS` | `5` | Failed password attempts before lockout |
| `PIXELS_ROOM_ENTRY_ATTEMPT_WINDOW` | `5m` | Password failure accumulation window |
| `PIXELS_ROOM_ENTRY_LOCKOUT_SECONDS` | `600` | Password lockout duration in seconds |
| `PIXELS_ROOM_ENTRY_PASSWORD_COST` | `10` | Bcrypt room password cost |
| `PIXELS_ROOM_ENTRY_TRUSTED_TTL` | `10s` | One time admin entry bypass lifetime |
| `PIXELS_ROOM_MODERATION_MIN_MUTE_MINUTES` | `1` | Minimum room mute duration |
| `PIXELS_ROOM_MODERATION_MAX_MUTE_MINUTES` | `1440` | Maximum room mute duration |
| `PIXELS_ROOM_FLOORPLAN_REJECT_ZERO_HEIGHT` | `true` | Reject custom maps without usable tiles |
| `PIXELS_ROOM_FLOORPLAN_SAVE_COOLDOWN` | `3s` | Per player floor plan save cooldown |
| `PIXELS_ROOM_PROMOTION_DURATION` | `2h` | Purchased room advertisement duration |
| `PIXELS_ROOM_SPECTATOR_ENABLED` | `true` | Allow spectator room entry mode |
| `PIXELS_ROOM_BUNDLE_BOTS_ENABLED` | `true` | Clone bots from purchased room templates |
| `PIXELS_BUILDERS_CLUB_FURNITURE_LIMIT` | `0` | Builders Club room furniture allowance |

## Chat and messenger

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_CHAT_MAX_MESSAGE_RUNES` | `256` | Maximum room chat length |
| `PIXELS_CHAT_FLOOD_TIER0_MAX_MESSAGES` | `100` | Loose flood budget |
| `PIXELS_CHAT_FLOOD_TIER0_WINDOW` | `1s` | Loose flood window |
| `PIXELS_CHAT_FLOOD_TIER1_MAX_MESSAGES` | `10` | Normal flood budget |
| `PIXELS_CHAT_FLOOD_TIER1_WINDOW` | `5s` | Normal flood window |
| `PIXELS_CHAT_FLOOD_TIER2_MAX_MESSAGES` | `6` | Strict flood budget |
| `PIXELS_CHAT_FLOOD_TIER2_WINDOW` | `5s` | Strict flood window |
| `PIXELS_CHAT_LOG_WHISPERS` | `false` | Persist whispers in moderation history |
| `PIXELS_CHAT_LOG_RETENTION_DAYS` | `14` | Chat history partition retention |
| `PIXELS_CHAT_LOG_FLUSH_INTERVAL` | `2s` | History write batch cadence |
| `PIXELS_CHAT_LOG_FLUSH_BATCH_SIZE` | `200` | Maximum rows per history batch |
| `PIXELS_CHAT_LOG_QUEUE_SIZE` | `4096` | Nonblocking history queue capacity |
| `PIXELS_MESSENGER_MAX_FRIENDS` | `200` | Normal friend list capacity |
| `PIXELS_MESSENGER_MAX_FRIENDS_CLUB` | `500` | Active club friend list capacity |
| `PIXELS_MESSENGER_SEARCH_MAX_RESULTS` | `50` | Maximum username search results |
| `PIXELS_MESSENGER_SEARCH_CACHE_TTL` | `30s` | Shared search result cache lifetime |
| `PIXELS_MESSENGER_FRIEND_CACHE_TTL` | `30s` | Durable friend card cache lifetime |
| `PIXELS_MESSENGER_SEARCH_THROTTLE` | `3s` | Per player search interval |
| `PIXELS_MESSENGER_CHAT_THROTTLE` | `750ms` | Per sender private message interval |
| `PIXELS_MESSENGER_CHAT_FILTER_ENABLED` | `false` | Apply hotel word filter to private messages |
| `PIXELS_MESSENGER_CHAT_LOG_ENABLED` | `false` | Persist accepted private messages |

## Players and Navigator

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_PLAYER_USERNAME_MIN_LENGTH` | `3` | Minimum username length |
| `PIXELS_PLAYER_USERNAME_MAX_LENGTH` | `15` | Maximum username length |
| `PIXELS_PLAYER_USERNAME_ALLOWED_SYMBOLS` | `_-=!?@:,.\'` | Extra username symbols |
| `PIXELS_PLAYER_USERNAME_RESERVATION_TTL` | `2m` | Rename reservation lifetime |
| `PIXELS_PLAYER_USERNAME_RESERVED` | `admin,moderator,staff,system` | Reserved usernames |
| `PIXELS_PLAYER_MOTTO_MAX_RUNES` | `38` | Maximum motto length |
| `PIXELS_PLAYER_TAG_MAX_COUNT` | `5` | Maximum profile tags |
| `PIXELS_PLAYER_TAG_MAX_RUNES` | `32` | Maximum runes per tag |
| `PIXELS_PLAYER_RESPECT_DAILY_LIMIT` | `3` | Daily player respect budget |
| `PIXELS_PLAYER_PET_RESPECT_DAILY_LIMIT` | `3` | Daily pet respect budget |
| `PIXELS_PLAYER_RESPECT_THROTTLE` | `250ms` | Minimum interval between respect actions |
| `PIXELS_HOTEL_TIMEZONE` | `America/Bogota` | Calendar timezone for hotel days |
| `PIXELS_PLAYER_WARDROBE_MIN_SLOT` | `1` | First valid wardrobe slot |
| `PIXELS_PLAYER_WARDROBE_MAX_SLOT` | `10` | Last valid wardrobe slot |
| `PIXELS_PLAYER_SETTINGS_FLUSH_INTERVAL` | `250ms` | Deferred settings write cadence |
| `PIXELS_PLAYER_SETTINGS_PENDING_LIMIT` | `4096` | Maximum pending settings writes |
| `PIXELS_FIGURE_DATA_URL` | `https://storageapi.pixelados.net/assets-prod/gamedata/FigureData.json` | Nitro-compatible avatar entitlement catalog URL |
| `PIXELS_FIGURE_DATA_PATH` | empty | Optional local JSON or XML override for the catalog URL |
| `PIXELS_FIGURE_DATA_TIMEOUT` | `15s` | Remote figure-data request timeout |
| `PIXELS_FIGURE_DATA_MAX_BYTES` | `16777216` | Maximum remote or local figure-data document size |
| `PIXELS_NAVIGATOR_SEARCH_LIMIT` | `50` | Maximum room search results |
| `PIXELS_NAVIGATOR_HISTORY_LIMIT` | `25` | Visit history size |
| `PIXELS_NAVIGATOR_FAVORITE_LIMIT` | `30` | Favorite room limit |
| `PIXELS_NAVIGATOR_HISTORY_QUEUE_SIZE` | `1024` | Visit telemetry queue capacity |
| `PIXELS_NAVIGATOR_HISTORY_DEDUPE_WINDOW` | `30s` | Repeated visit dedupe window |
| `PIXELS_NAVIGATOR_PREFERENCE_FLUSH_INTERVAL` | `250ms` | Deferred preference write cadence |
| `PIXELS_NAVIGATOR_PREFERENCE_PENDING_LIMIT` | `4096` | Maximum pending preference writes |
| `PIXELS_NAVIGATOR_WINDOW_POSITION_LIMIT` | `32768` | Absolute saved window coordinate bound |
| `PIXELS_NAVIGATOR_WINDOW_MIN_WIDTH` | `320` | Minimum saved window width |
| `PIXELS_NAVIGATOR_WINDOW_MAX_WIDTH` | `4096` | Maximum saved window width |
| `PIXELS_NAVIGATOR_WINDOW_MIN_HEIGHT` | `240` | Minimum saved window height |
| `PIXELS_NAVIGATOR_WINDOW_MAX_HEIGHT` | `2160` | Maximum saved window height |

## Subscription, Marketplace, and trade

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_SUBSCRIPTION_TICK_INTERVAL` | `1m` | Entitlement and payday scheduler cadence |
| `PIXELS_SUBSCRIPTION_PAYDAY_INTERVAL` | `744h` | One HC accounting cycle |
| `PIXELS_SUBSCRIPTION_KICKBACK_PERCENTAGE` | `0.10` | Eligible catalog spending returned at payday |
| `PIXELS_SUBSCRIPTION_PAYDAY_CURRENCY_TYPE` | `-1` | Payday reward currency id |
| `PIXELS_SUBSCRIPTION_BONUSRARE_CURRENCY_TYPE` | `5` | Bonus Rare progress currency id |
| `PIXELS_SUBSCRIPTION_BONUSRARE_THRESHOLD` | `120` | Balance needed to fill Bonus Rare progress |
| `PIXELS_SUBSCRIPTION_BONUSRARE_PRODUCT_ID` | `0` | Bonus Rare furniture definition |
| `PIXELS_MARKETPLACE_ENABLED` | `true` | Enable Marketplace operations |
| `PIXELS_MARKETPLACE_COMMISSION_PERCENT` | `1` | Buyer commission percentage |
| `PIXELS_MARKETPLACE_TOKEN_COST` | `1` | Credit cost per token package |
| `PIXELS_MARKETPLACE_TOKEN_PACKAGE_SIZE` | `5` | Tokens per package |
| `PIXELS_MARKETPLACE_ADVERTISEMENT_COST` | `0` | Promoted listing price projection |
| `PIXELS_MARKETPLACE_MIN_PRICE` | `1` | Minimum seller price |
| `PIXELS_MARKETPLACE_MAX_PRICE` | `1000000` | Maximum seller price |
| `PIXELS_MARKETPLACE_OFFER_DURATION` | `48h` | Open listing lifetime |
| `PIXELS_MARKETPLACE_DISPLAY_DURATION` | `168h` | Historical listing display lifetime |
| `PIXELS_MARKETPLACE_SEARCH_CACHE_TTL` | `30s` | Shared Marketplace search cache lifetime |
| `PIXELS_MARKETPLACE_EXPIRY_INTERVAL` | `1m` | Listing expiry scheduler cadence |
| `PIXELS_TRADE_ENABLED` | `true` | Enable direct player trading |
| `PIXELS_TRADE_START_THROTTLE` | `10s` | Minimum interval between trade starts |
| `PIXELS_TRADE_MAX_ITEMS` | `12` | Maximum offered items per participant |
| `PIXELS_TRADE_AUDIT_ENABLED` | `true` | Persist completed trade audits |

## Moderation and guardians

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_MODERATION_ENABLED` | `true` | Enable CFH, staff, guide, and guardian flows |
| `PIXELS_MODERATION_REPORT_LIMIT` | `3` | Reports per throttle window |
| `PIXELS_MODERATION_REPORT_WINDOW` | `10m` | Distributed report throttle window |
| `PIXELS_MODERATION_CONTEXT_WINDOW` | `50` | Maximum history rows frozen into an issue |
| `PIXELS_GUARDIAN_COUNT` | `3` | Reviewers assigned to one guardian case |
| `PIXELS_GUARDIAN_VOTE_WINDOW` | `1m` | Guardian voting duration |
| `PIXELS_GUARDIAN_IGNORE_LIMIT` | `3` | Ignored offers before exclusion |
| `PIXELS_GUARDIAN_EXCLUSION` | `30m` | Guardian exclusion duration |

## Bots and pets

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_BOT_MAX_PER_ROOM` | `25` | Ordinary player bot room limit |
| `PIXELS_BOT_MAX_INVENTORY` | `25` | Bot inventory limit |
| `PIXELS_BOT_WALK_RADIUS` | `5` | Autonomous bot walk radius |
| `PIXELS_BOT_LIMIT_WALK_RADIUS` | `true` | Enforce configured walk radius |
| `PIXELS_BOT_BARTENDER_COMMAND_DISTANCE` | `6` | Bartender hearing distance |
| `PIXELS_BOT_BARTENDER_REACH_DISTANCE` | `3` | Bartender delivery distance |
| `PIXELS_BOT_PLACEMENT_MESSAGES` | `bots.placement.hello;bots.placement.party;bots.placement.welcome` | Semicolon separated placement message keys |
| `PIXELS_BOT_POSITION_FLUSH_INTERVAL` | `5s` | Durable bot position write cadence |
| `PIXELS_PET_ENABLED` | `true` | Enable pet behavior and protocol flows |
| `PIXELS_PET_MAX_PER_ROOM` | `25` | Total pets allowed in a room |
| `PIXELS_PET_MAX_PER_OWNER_ROOM` | `10` | One owner's pets allowed in a room |
| `PIXELS_PET_MAX_INVENTORY` | `25` | Pet inventory limit |
| `PIXELS_PET_INVENTORY_FRAGMENT_SIZE` | `100` | Records per Nitro pet packet fragment |
| `PIXELS_PET_WALK_RADIUS` | `5` | Autonomous pet walk radius |
| `PIXELS_PET_DECISION_MINIMUM` | `5s` | Minimum autonomous decision delay |
| `PIXELS_PET_DECISION_MAXIMUM` | `15s` | Maximum autonomous decision delay |
| `PIXELS_PET_POSITION_FLUSH_INTERVAL` | `5s` | Durable pet position write cadence |
| `PIXELS_PET_STAT_DECAY_INTERVAL` | `30m` | Need decay materialization cadence |
| `PIXELS_PET_ENERGY_DECAY` | `1` | Energy lost per decay interval |
| `PIXELS_PET_HAPPINESS_DECAY` | `1` | Happiness lost per decay interval |
| `PIXELS_PET_RESPECT_MINIMUM_AGE` | `72h` | Minimum age before respect |
| `PIXELS_PET_RESPECT_EXPERIENCE` | `10` | Experience gained from respect |
| `PIXELS_PET_RESPECT_DAILY_LIMIT` | `3` | Global daily respect budget |
| `PIXELS_PET_ALLOW_RESPECT_OWN` | `false` | Allow respecting an owned pet |
| `PIXELS_PET_PLANT_REWARD_DEFINITION_ID` | `4582` | Monsterplant growth reward definition |
| `PIXELS_PET_PLANT_COMPOST_DEFINITION_ID` | `4830` | Monsterplant compost definition |
| `PIXELS_PET_PLANT_GROW_DURATION` | `168h` | Monsterplant growth duration |
| `PIXELS_PET_PLANT_LIFE_DURATION` | `168h` | Monsterplant mature lifetime |
| `PIXELS_PET_PACKAGE_TIMEOUT` | `2m` | Pet package interaction deadline |
| `PIXELS_PET_BREEDING_TIMEOUT` | `2m` | Breeding interaction deadline |
| `PIXELS_PET_BREEDING_MINIMUM_AGE` | `72h` | Minimum parent age |
| `PIXELS_PET_UNLOAD_FLUSH_TIMEOUT` | `3s` | Pet unload persistence deadline |

## Furniture, rollers, and WIRED

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED` | `false` | Let paired teleports bypass locked room modes |
| `PIXELS_FIREWORK_DEFAULT_RECHARGE` | `5s` | Default firework recharge duration |
| `PIXELS_RENTABLE_DURATION` | `24h` | Furniture rental duration |
| `PIXELS_RENTABLE_PRICE_CREDITS` | `10` | Rental credit price |
| `PIXELS_RENTABLE_BUYOUT_CREDITS` | `50` | Rental buyout credit price |
| `PIXELS_MYSTERYBOX_WAIT` | `3s` | Mystery box reveal delay |
| `PIXELS_MYSTERYBOX_PRIZE_DEFINITION_ID` | `1` | Mystery box reward definition |
| `PIXELS_ROLLER_HOOK_DELAY` | `400ms` | Delay between roller animation and walk hooks |
| `PIXELS_ROLLER_MAX_AVATARS` | `1` | Stationary avatars moved per roller cycle |
| `PIXELS_ROLLER_NO_RULES` | `false` | Disable roller placement and chain rules |
| `PIXELS_WIRED_ENABLED` | `true` | Enable WIRED configuration and execution |
| `PIXELS_WIRED_MAX_SELECTION` | `20` | Maximum selected furniture per node |
| `PIXELS_WIRED_MAX_DELAY_PULSES` | `7200` | Maximum configured delayed pulses |
| `PIXELS_WIRED_MAX_EVENTS_PER_TRACE` | `128` | Derived event budget per trace |
| `PIXELS_WIRED_MAX_STACKS_PER_TRACE` | `64` | Stack budget per trace |
| `PIXELS_WIRED_MAX_EFFECTS_PER_TRACE` | `128` | Effect budget per trace |
| `PIXELS_WIRED_MAX_CALL_DEPTH` | `10` | Recursive call depth limit |
| `PIXELS_WIRED_MAX_DELAYED_PER_ROOM` | `512` | Outstanding delayed actions per room |
| `PIXELS_WIRED_HIGHSCORE_TOP` | `50` | Rows projected by a highscore board |

## Social groups and crafting

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_GROUP_CREATION_COST` | `10` | Credit price for creating a social group |
| `PIXELS_GROUP_REQUIRE_CLUB` | `true` | Require active club for group creation |
| `PIXELS_GROUP_OWNED_LIMIT` | `100` | Groups one player may own |
| `PIXELS_GROUP_MEMBERSHIP_LIMIT` | `100` | Groups one player may join |
| `PIXELS_GROUP_MEMBER_LIMIT` | `50000` | Members allowed in one group |
| `PIXELS_GROUP_PENDING_LIMIT` | `100` | Pending requests allowed per group |
| `PIXELS_GROUP_MEMBER_PAGE_SIZE` | `14` | Member rows per Nitro page |
| `PIXELS_GROUP_MAX_SEARCH_LENGTH` | `64` | Maximum group search text length |
| `PIXELS_GROUP_BULK_APPROVE_LIMIT` | `100` | Requests approved per bulk operation |
| `PIXELS_GROUP_FURNITURE_CLEANUP_LIMIT` | `1000` | Furniture rows handled by cleanup |
| `PIXELS_GROUP_FORUM_PAGE_SIZE` | `50` | Forum rows per page |
| `PIXELS_GROUP_FORUM_SUBJECT_LIMIT` | `120` | Forum subject length |
| `PIXELS_GROUP_FORUM_MESSAGE_LIMIT` | `4000` | Forum message length |
| `PIXELS_GROUP_FORUM_POST_COOLDOWN` | `3s` | Per player forum posting interval |
| `PIXELS_GROUP_FORUM_ACTIVE_WINDOW` | `168h` | Forum active thread window |
| `PIXELS_GROUP_CACHE_TTL` | `10m` | Shared social group cache lifetime |
| `PIXELS_GROUP_FORUM_CURSOR_TTL` | `5m` | Forum cursor lifetime |
| `PIXELS_GROUP_DEACTIVATION_RETENTION` | `8760h` | Soft deleted group retention |
| `PIXELS_CRAFTING_ENABLED` | `true` | Enable altar crafting |
| `PIXELS_CRAFTING_RECYCLER_ENABLED` | `true` | Enable Ecotron recycling |
| `PIXELS_CRAFTING_RECYCLER_BATCH_SIZE` | `8` | Exact recycler input size |
| `PIXELS_CRAFTING_RECYCLER_RARITY_CHANCES` | `5=1000,4=100,3=20,2=5` | Rarity tier denominators |

## Camera and progression

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_CAMERA_ENABLED` | `true` | Enable camera operations |
| `PIXELS_CAMERA_CAPTURE_COOLDOWN` | `3s` | Minimum interval between captures |
| `PIXELS_CAMERA_MAX_PHOTO_BYTES` | `2097152` | Maximum full PNG size |
| `PIXELS_CAMERA_MAX_THUMBNAIL_BYTES` | `1048576` | Maximum thumbnail PNG size |
| `PIXELS_CAMERA_PENDING_RETENTION` | `24h` | Unreferenced pending photo lifetime |
| `PIXELS_CAMERA_SUPERSEDED_RETENTION` | `1h` | Replaced photo grace period |
| `PIXELS_CAMERA_CLEANUP_INTERVAL` | `5m` | Orphan cleanup cadence |
| `PIXELS_CAMERA_CLEANUP_RETRY` | `5m` | Failed deletion retry delay |
| `PIXELS_CAMERA_CLEANUP_BATCH_SIZE` | `100` | Cleanup rows per pass |
| `PIXELS_PROGRESSION_ENABLED` | `true` | Enable achievements, talents, quests, and quizzes |
| `PIXELS_PROGRESSION_TRADE_REQUIRES_PERK` | `false` | Require citizenship trade perk |
| `PIXELS_PROGRESSION_GUIDE_MIN_TRACK_LEVEL` | `0` | Helpers talent level required for guide duty |
| `PIXELS_PROGRESSION_PRESENCE_INTERVAL` | `5m` | Online presence progress cadence |
| `PIXELS_PROGRESSION_FLUSH_INTERVAL` | `2s` | Progress write batch cadence |
| `PIXELS_PROGRESSION_DAILY_POOL_SEED` | empty | Stable salt for daily quest selection |

## Games and Game Center

| Variable | Default | Purpose |
|---|---|---|
| `PIXELS_GAMES_ENABLED` | `true` | Enable server authored room games |
| `PIXELS_GAMES_FREEZE_POINTS_FREEZE` | `10` | Freeze score for freezing a player |
| `PIXELS_GAMES_FREEZE_POINTS_BLOCK` | `1` | Freeze score for breaking a block |
| `PIXELS_GAMES_FREEZE_POINTS_EFFECT` | `3` | Freeze score for revealing an effect |
| `PIXELS_GAMES_FREEZE_POWERUP_CHANCE` | `33` | Freeze powerup reveal percentage |
| `PIXELS_GAMES_FREEZE_MAX_SNOWBALLS` | `5` | Maximum Freeze snowball capacity |
| `PIXELS_GAMES_FREEZE_MAX_LIVES` | `3` | Maximum Freeze lives |
| `PIXELS_GAMES_FREEZE_LOOSE_SNOWBALLS` | `5` | Snowballs granted by loose powerup |
| `PIXELS_GAMES_FREEZE_LOOSE_BOOST` | `3` | Range boost granted by loose powerup |
| `PIXELS_GAMES_FREEZE_FROZEN_SECONDS` | `5` | Freeze immobilization duration |
| `PIXELS_GAMES_FREEZE_PROTECTION_SECONDS` | `10` | Freeze protection duration |
| `PIXELS_GAMES_FREEZE_PROTECTION_STACK` | `true` | Let protection duration stack |
| `PIXELS_GAMES_BANZAI_POINTS_STEAL` | `0` | Battle Banzai tile steal score |
| `PIXELS_GAMES_BANZAI_POINTS_FILL` | `0` | Battle Banzai area fill score |
| `PIXELS_GAMES_BANZAI_POINTS_LOCK` | `1` | Battle Banzai tile lock score |
| `PIXELS_GAMECENTER_ENABLED` | `true` | Enable external Game Center launcher |

## Source of truth

`.env.example` remains the executable copy and paste template. This page explains every current variable, while the config holders beside each component define parsing and normalization. A change that adds, removes, or renames a variable must update all three in the same commit.
