# Plan: Derechos y Moderación de Rooms (`internal/realm/room/rights`, `internal/realm/room/moderation`, `internal/realm/room/audit`)

Este plan implementa completamente la **Parte 4** de `plan/REMAINING-ROOMS.md` (derechos de construcción y moderación: kick/mute/ban), cierra los dos puntos que `plan/rooms/ENTRY.md` dejó deliberadamente en stub (`entry.RightsChecker`/`entry.BanChecker`), y agrega un sistema de **auditoría/historial completo** que no existe en ninguno de los dos documentos anteriores: quién otorgó/revocó derechos y cuándo, quién muteó/baneó/kickeó a quién, con qué razón y por cuánto tiempo, consultable por room, por jugador afectado (en una room o en todas), y por jugador que ejecutó la acción (en una room o en todas).

**Nota sobre fuentes**: el diseño de duraciones/persistencia de kick/mute/ban se basa exclusivamente en la investigación de Arcturus ya volcada y vetada en `plan/REMAINING-ROOMS.md` Parte 4 (código Java real ya auditado en este proyecto) — no en artículos externos sobre el cliente oficial moderno de Habbo, que corre un protocolo y un cliente distintos a los que Pixels emula. Donde el research existente no confirma algo (ej. un campo de razón en los packets de moderación), este plan lo dice explícitamente en vez de inventarlo.

Estado: implementado completamente en RM1-RM8. Este documento conserva las decisiones y límites del feature.

---

## Parte 0 — Punto de partida real (grounding, confirmado leyendo el código actual)

`plan/rooms/ENTRY.md` ya se implementó de punta a punta (confirmado leyendo `internal/realm/room/entry/*`, `internal/realm/room/doorbell/*`, `internal/realm/room/commands/enter/*`, y `pkg/http/room/routes/teleport.go`) — incluyendo el hangout timeout, el freeze por intentos de contraseña, el `Trusted`/`GrantTrusted` bypass, y el endpoint admin de teleport single-player. Este plan no repite nada de eso; parte de ahí.

| Ya existe | Dónde | Nota |
| --- | --- | --- |
| `entry.RightsChecker` interface — `HasRights(ctx, roomID, playerID) (bool, error)` | `internal/realm/room/entry/service.go` | Ya declarada, ya consultada por `Service.Authorize`/`Service.CanAnswerDoorbell` — pero `service.rights` es `nil` hoy (nadie la implementó todavía). Este plan la implementa. |
| `entry.BanChecker` interface — `IsBanned(ctx, roomID, playerID) (bool, error)` | `internal/realm/room/entry/service.go` | Ya declarada, ya consultada por `Service.checkBan` — `service.bans` es `nil` hoy. Este plan la implementa. |
| `Service.WithRights(RightsChecker) *Service` / `Service.WithBans(BanChecker) *Service` | ídem | Los puntos de enganche exactos donde este plan conecta sus servicios nuevos — no hace falta tocar `entry` para nada más. |
| `permission.Checker`/`permission.RegisterNode` — real, no solo planeado | `internal/permission/*` | Confirmado en `plan/PERMISSIONS.md`/`plan/rooms/ENTRY.md` y re-confirmado ahora: `catalog`, `currency`, `player`, y `room` ya tienen su propio `permissions.go`. |
| `internal/realm/room/permissions.go` — ya existe, 3 nodos | mismo archivo | Hoy tiene `EnterAny`, `EnterFull`, `AnswerAnyDoorbell` (los tres ya consumidos por `entry.Service`). Este plan agrega los nodos de moderación/derechos al MISMO archivo (Parte 4), no crea uno nuevo. |
| Columnas `moderation_mute`, `moderation_kick`, `moderation_ban` (`smallint`) | `internal/realm/room/database/migrations/0002_create_room_records.sql` | Existen en Postgres, **no están ni en `model.Room` ni en `roomColumns`/`scanRoom`** — confirmado, gap real. `REMAINING-ROOMS.md` 4.3 ya señaló que Arcturus las persiste pero nunca las lee; acá se cierra ese hueco definiendo el enum concreto que faltaba decidir (Parte 2.5). |
| Ninguna tabla `room_rights`/`room_bans`/`room_mutes`/tabla de auditoría | grep sin resultados en todo `internal/` | Completamente greenfield — ni siquiera hay una migración a medio empezar. |
| Ningún paquete `internal/realm/room/{rights,moderation,audit}` | `find internal/realm/room/commands` | Tampoco existen los comandos `rights/*`/`moderation/*`. Greenfield. |
| `currency_ledger_entries` — plantilla de auditoría ya establecida | `internal/realm/inventory/currency/model/ledger.go`, `repository/query.go` | `LedgerEntry{PlayerID, Delta, BalanceAfter, Reason, ActorKind, ActorID *int64, CreatedAt}` — append-only, un `insertLedger` por mutación, nunca update/delete. **Hoy solo se escribe, no hay ningún endpoint ni método para leerlo** — este plan sí construye la capa de lectura completa (Parte 3), algo que ni siquiera `currency` tiene todavía; se toma el shape `ActorKind`/`ActorID`/`CreatedAt` como base, no se reinventa — **salvo `Reason`**, que este plan no adopta (ver 2.7: sin campo de razón en moderación de rooms). |
| `leave.Handler` ya reusable para expulsar a un jugador de su room activo | `internal/realm/room/commands/leave/command.go`, ya reusado por `enter/runtime.go:leavePreviousRoom` | El kick de este plan reusa exactamente el mismo mecanismo (construir un `leavecmd.Command{PlayerID: targetID}` y llamarlo), en vez de reimplementar la salida de un jugador. |
| Research ya vetado de Arcturus sobre kick/mute/ban (`plan/REMAINING-ROOMS.md` Parte 4.1-4.3) | — | **Kick**: sin persistencia, efecto inmediato, dueño/rights-holder/staff. **Mute**: `Room.mutedHabbos`, en memoria con **TTL en minutos** (valor libre, no un enum fijo de opciones) — Arcturus lo pierde si el room se descarga, hallazgo que `REMAINING-ROOMS.md` 4.2 ya decidió NO replicar (Pixels persiste). **Ban**: persistido, **tres duraciones fijas** (hora/día/permanente ≈10 años). **Ningún campo de razón/mensaje** aparece en el research de los packets de moderación ejecutados por el dueño de la sala — no se encontró evidencia de esto en el código Java auditado. |

---

## Parte 1 — Derechos de construcción (`internal/realm/room/rights`)

### 1.1 Esquema

```sql
create table room_rights (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    granted_by_player_id bigint not null references players(id),
    created_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_rights_player_id_idx on room_rights (player_id);
```

Diferencia deliberada contra el sketch original de `REMAINING-ROOMS.md` 4.1 (que solo tenía `room_id, player_id`): se agrega `granted_by_player_id`. No es estrictamente necesario para `HasRights` (un simple `exists`), pero evita tener que ir a buscar "quién lo otorgó" a otro lado el día que se quiera mostrar ese dato sin depender pura y exclusivamente de la tabla de auditoría de la Parte 3 (que además sí lo tiene, con más detalle — esta columna es una comodidad de lectura rápida sobre el estado actual, no un reemplazo de la auditoría).

### 1.2 Repository y Service

```go
// internal/realm/room/rights/repository/contract.go

// Store persists room rights membership.
type Store interface {
    Grant(ctx context.Context, roomID int64, playerID int64, grantedByPlayerID int64) error
    Revoke(ctx context.Context, roomID int64, playerID int64) (bool, error)
    RevokeAll(ctx context.Context, roomID int64) (int, error)
    List(ctx context.Context, roomID int64) ([]model.Right, error)
    Exists(ctx context.Context, roomID int64, playerID int64) (bool, error)
}
```

```go
// internal/realm/room/rights/service.go

// Manager grants, revokes, and resolves room build rights.
type Manager interface {
    // GrantRights grants a player build rights, requiring ownership or a staff node.
    GrantRights(ctx context.Context, roomID int64, granterID int64, targetID int64) error
    // RevokeRights revokes one player's rights, requiring ownership or a staff node.
    RevokeRights(ctx context.Context, roomID int64, revokerID int64, targetID int64) error
    // RevokeAllRights revokes every rights holder of a room at once, returning the count removed.
    RevokeAllRights(ctx context.Context, roomID int64, revokerID int64) (int, error)
    // RelinquishRights lets a player drop their own rights, unconditionally.
    RelinquishRights(ctx context.Context, roomID int64, playerID int64) error
    // ListRights lists current rights holders.
    ListRights(ctx context.Context, roomID int64) ([]model.Right, error)
    // HasRights reports whether a player currently holds room rights — satisfies entry.RightsChecker directly.
    HasRights(ctx context.Context, roomID int64, playerID int64) (bool, error)
}
```

`HasRights` tiene exactamente la firma de `entry.RightsChecker` — `*Service` se pasa directo a `entry.Service.WithRights(rightsService)` sin ningún adapter intermedio (Parte 5).

Autorización dentro del propio service (no en el comando — mismo criterio que el resto del proyecto, la regla de negocio vive en el service, el comando solo despacha):
- `GrantRights`/`RevokeRights`: requiere `room.OwnerPlayerID == granterID` **o** `permission.Checker.HasPermission(ctx, granterID, room.RightsAnyGrant/Revoke)` (staff). El sketch original de `REMAINING-ROOMS.md` 4.1 no contemplaba un nodo `.own` separado para esto porque no existía el sistema de permisos todavía — con `plan/PERMISSIONS.md` ya implementado, se agrega `room.RightsOwnGrant/Revoke` (Parte 4) para que, igual que con moderación, se le pueda revocar a un dueño puntual la capacidad de otorgar/revocar derechos en su propia sala sin tocar nada del sistema de staff.
- `RevokeAllRights`: mismo criterio (`OwnerPlayerID` o `RightsAnyRevoke`).
- `RelinquishRights`: **sin ningún chequeo** — cualquiera puede auto-revocarse (confirmado por `REMAINING-ROOMS.md` 4.1, "cualquiera puede mandarlo, sin chequeo — solo afecta al que lo manda").
- Otorgar a un jugador offline funciona igual que a uno presente (confirmado por el research — no depende de estar conectado, la tabla es puro estado persistido).

### 1.3 Comandos

`internal/realm/room/commands/rights/{grant,revoke,revokeall,list,relinquish}` — mismo patrón `command.Command`/`command.Handler[T]` que el resto del proyecto (`enter`, `leave`, etc.).

### 1.4 Packets

| Dirección | Paquete | Contenido |
| --- | --- | --- |
| Inbound | `rights/grant` | `playerId int32` (o `username string`, a confirmar contra el packet real — Arcturus permite otorgar a un amigo offline por nombre) |
| Inbound | `rights/revoke` | `playerIds []int32` (Arcturus manda un array, un solo packet para varios — confirmado por `REMAINING-ROOMS.md` 4.1) |
| Inbound | `rights/revokeall` | sin campos (implícito: "mi room actual") |
| Inbound | `rights/relinquish` | sin campos |
| Inbound | `rights/list` | sin campos |
| Outbound | `rights/level` | `level int32` (`NONE=0/RIGHTS=1/OWNER=2` — simplificado respecto al enum completo de Arcturus, `GUILD_RIGHTS`/`GUILD_ADMIN` se agregan el día que exista un realm de grupos, mismo criterio ya fijado en `REMAINING-ROOMS.md` 4.1), mandado al jugador afectado tras cada cambio |
| Outbound | `rights/list` | lista de `(playerId, username)` actualmente con derechos, respuesta a `rights/list` |

Los headers fueron confirmados contra Nitro real durante RM2 y están enumerados en la Parte 2.7.

### 1.5 Eventos

```go
// internal/realm/room/events/rightsgranted/event.go
type Payload struct {
    RoomID   int64
    PlayerID int64
    ActorID  int64
}

// internal/realm/room/events/rightsrevoked/event.go
type Payload struct {
    RoomID   int64
    PlayerID int64
    ActorID  int64
    Action   RevokeAction // Explicit | RevokedAll | Relinquished — nunca texto libre
}
```

Publicados por `rights.Service` tras cada mutación exitosa, vía `bus.Publisher` (mismo patrón que `leave.Handler`/`enter.Handler` ya usan). Dos consumidores independientes, ninguno sabe del otro:
1. Un broadcaster (`internal/realm/room/rights/broadcast`) que, si el jugador afectado está online, le manda `rights/level` actualizado — mismo patrón que `currency/broadcast`.
2. El subscriber de auditoría de la Parte 3.5, que inserta la fila en `room_rights_audit` — **nunca** el propio `rights.Service` escribiendo ahí directo (desacople deliberado, ver 3.5).

### 1.6 Wiring

`internal/realm/room/module.go` gana: construir `rights.Service`, exponerlo como `roomrights.Manager` vía `fx.Provide`, y encadenar `entry.Service.WithRights(rightsService)` (Parte 5 tiene el detalle completo del wiring conjunto con moderación).

### 1.7 Tests

- Otorgar derechos a un jugador offline funciona igual que a uno presente.
- Otorgar/revocar sin ser dueño ni tener `RightsAnyGrant/Revoke` falla.
- Otorgar/revocar siendo dueño, o teniendo el nodo de staff, funciona.
- Revocar todos los derechos de una sala de una vez remueve a todos y retorna el conteo correcto.
- Auto-revocarse (`relinquish`) nunca falla por permisos, sin importar quién lo mande.
- `HasRights` retorna `false` para un jugador sin fila en `room_rights` y `true` para uno con fila — usado directo como `entry.RightsChecker` en un test de integración con `entry.Service.Authorize` (confirma el acople real, no solo el método aislado).
- Cada mutación exitosa publica el evento correspondiente con el `ActorID` correcto.

---

## Parte 2 — Moderación (`internal/realm/room/moderation`)

### 2.1 Kick — sin estado propio, pero siempre logueado/persistido en la auditoría

**No tiene tabla de estado propia** — a diferencia de mute/ban, un kick no tiene nada que "siga vigente" con el correr del tiempo (no hay un `ends_at` que expire): es un efecto inmediato, el jugador sale del room activo ahora mismo, y puede volver a entrar de inmediato si el room lo permite (mismo comportamiento confirmado en `REMAINING-ROOMS.md` 4.2). `internal/realm/room/commands/moderation/kick` reusa exactamente `leavecmd.Handler` (mismo mecanismo que `enter/runtime.go:leavePreviousRoom` ya usa para cambiar de room) para sacar al objetivo del `roomlive.Registry`, en vez de reimplementar esa lógica.

**Pero SÍ queda persistido/logueado como acción de moderación** — "sin estado propio" no significa "sin rastro": cada kick exitoso inserta, sin excepción, una fila en `room_moderation_actions` (`action_type = 'kick'`, `duration_seconds = null`, `expires_at = null`, Parte 3) exactamente igual que un mute o un ban. Esto es lo que permite responder después "¿quién kickeó a quién, de qué room, y cuándo" con la misma consulta de historial que ya cubre mute/ban (Parte 3.3) — un kick nunca es una acción invisible para la auditoría, aunque no deje ningún estado vigente en `room_bans`/`room_mutes`.

### 2.2 Mute — persistido (decisión ya tomada en `REMAINING-ROOMS.md` 4.2, se ejecuta acá)

```sql
create table room_mutes (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    ends_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_mutes_room_active_idx on room_mutes (room_id, ends_at desc);
create index room_mutes_player_active_idx on room_mutes (player_id, ends_at desc);
```

**Duración**: `durationMinutes int32`, validado server-side contra `Config.MinMuteMinutes`/`Config.MaxMuteMinutes` (ej. 1-1440) — **no** un enum fijo de 2/5/10 minutos. El research ya vetado en `REMAINING-ROOMS.md` 4.2 confirma que Arcturus guarda un TTL en minutos como valor libre en `Room.mutedHabbos`, no tres opciones fijas — cualquier restricción a un puñado de valores fijos sería una decisión de **UI del cliente** (un dropdown puede ofrecer solo 3 valores), no una restricción de protocolo/servidor, y este plan no la impone del lado servidor sin evidencia real de que el protocolo la exija.

A diferencia de Arcturus (mute en memoria, se pierde si el room se descarga — bug ya identificado y decidido NO replicar), `room_mutes` persiste igual que `room_bans`.

### 2.3 Ban — persistido, tres duraciones fijas (confirmado por el research ya existente, no por fuentes externas)

```sql
create table room_bans (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    ends_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (room_id, player_id)
);

create index room_bans_room_active_idx on room_bans (room_id, ends_at desc);
create index room_bans_player_active_idx on room_bans (player_id, ends_at desc);
```

```go
// internal/realm/room/moderation/duration.go

// BanDuration names the fixed ban lengths confirmed against the reference research.
type BanDuration int16

const (
    // BanDurationHour bans for one hour.
    BanDurationHour BanDuration = 1
    // BanDurationDay bans for one day.
    BanDurationDay BanDuration = 2
    // BanDurationPermanent bans for ~10 years, matching the reference implementation's
    // "permanent" convention (an actual permanent ban with no end date would need its
    // own nullable-ends_at handling; this plan keeps parity with the confirmed research
    // instead of introducing a NULL-duration concept the reference never has).
    BanDurationPermanent BanDuration = 3
)

// Seconds returns the concrete duration a BanDuration represents.
func (duration BanDuration) Seconds() (int64, bool) { ... }
```

### 2.4 Unmute / Unban

`UnmuteRoom`/`UnbanRoom` actualizan la fila vigente (`update ... set ends_at = now() where room_id = $1 and player_id = $2 and ends_at > now()`) en vez de borrarla — deja la fila de estado como rastro de "esto existió y se levantó antes de tiempo" sin depender pura y exclusivamente de la tabla de auditoría (misma lógica que `granted_by_player_id` en 1.1: una comodidad de lectura rápida del estado actual, la auditoría de la Parte 3 sigue siendo la fuente completa independientemente).

### 2.5 Quién puede moderar: política por sala + nodos de permiso + inmunidad del objetivo

`model.Room` gana:
```go
// ModerationPolicy describes who may execute a moderation action in a room.
type ModerationPolicy int16

const (
    // ModerationPolicyOwnerOnly allows only the room owner.
    ModerationPolicyOwnerOnly ModerationPolicy = 0
    // ModerationPolicyOwnerAndRights allows the owner and any rights holder.
    ModerationPolicyOwnerAndRights ModerationPolicy = 1
    // ModerationPolicyOwnerRightsAndStaff is recorded for symmetry with the column's
    // full range but has no effect on the authorization formula below — staff with an
    // Any node always bypasses this column regardless of its value (see formula).
    ModerationPolicyOwnerRightsAndStaff ModerationPolicy = 2
)
```
```go
// ModerationMute, ModerationKick, ModerationBan control which room-scoped actors may
// execute each moderation action, independent of the global staff bypass.
ModerationMute ModerationPolicy
ModerationKick ModerationPolicy
ModerationBan  ModerationPolicy
```
Mapeadas 1:1 desde las columnas `moderation_mute/kick/ban` ya existentes en Postgres (`smallint`), cerrando el hueco que `REMAINING-ROOMS.md` 4.3 dejó pendiente ("falta decidir el enum concreto en implementación").

**Fórmula de autorización combinada** (une el research de Arcturus, el split `.own`/`.any` de `plan/PERMISSIONS.md`, y la columna de política por sala en un único chequeo, para cada acción `{kick, mute, ban}`):

```go
func (service *Service) authorize(ctx context.Context, room roommodel.Room, actorID int64, action Action) (bool, error) {
    // El staff con el nodo "any" siempre puede, en cualquier sala, sin importar la
    // política configurada por el dueño — el nodo "any" es intencionalmente más
    // fuerte que cualquier configuración local de la sala.
    if allowed, err := service.hasPermission(ctx, actorID, action.AnyNode); err != nil || allowed {
        return allowed, err
    }

    isOwner := room.OwnerPlayerID == actorID
    if isOwner {
        return service.hasPermission(ctx, actorID, action.OwnNode)
    }

    policy := action.Policy(room)
    if policy < roommodel.ModerationPolicyOwnerAndRights {
        return false, nil
    }
    hasRights, err := service.rights.HasRights(ctx, room.ID, actorID)
    if err != nil || !hasRights {
        return false, err
    }

    return service.hasPermission(ctx, actorID, action.OwnNode)
}
```

- **Staff (`ModerationAnyKick/Mute/Ban`)**: siempre autorizado, en cualquier sala, sin mirar la columna de política — es el equivalente exacto al `Trusted`/bypass de `ENTRY.md`, pero para acciones de moderación en vez de entrada.
- **Dueño**: autorizado solo si además tiene `ModerationOwnKick/Mute/Ban` — este es el punto exacto que motivó el split `.own`/`.any` en `plan/PERMISSIONS.md`: a una cuenta puntual se le puede revocar la capacidad de moderar SU PROPIA sala sin tocar el sistema de staff.
- **Rights holder (no dueño)**: autorizado solo si la columna `moderation_{mute,kick,ban}` de ESA sala es `OwnerAndRights` (o mayor) **y** además tiene `ModerationOwnKick/Mute/Ban` — la columna de política es la palanca que el DUEÑO controla ("¿dejo que mis rights-holders también moderen?"), el nodo `.own` es la palanca que STAFF controla por jugador ("¿esta cuenta puntual conserva esa capacidad, la tenga habilitada la sala o no?") — dos ejes independientes, intencionalmente.

**Inmunidad del objetivo** (`Unkickable`, Parte 4) — chequeo final, siempre, independiente de quién ejecuta:
```go
protected, err := service.hasPermission(ctx, targetID, room.Unkickable)
if err != nil {
    return err
}
if protected {
    return ErrTargetProtected
}
```
Mismo patrón que `acc_unkickable` en Arcturus (confirmado en `REMAINING-ROOMS.md` 1.6/4.2) — se chequea sobre el **objetivo**, nunca sobre quien ejecuta la acción, y gana incluso sobre un actor autorizado por la fórmula de arriba (staff incluido, salvo que el propio actor también tenga algún nodo que explícitamente lo exima — este plan no agrega ese escape hatch por defecto; se deja anotado en "Milestones futuros confirmados" si en la práctica hace falta un "puedo moderar incluso a alguien unkickable").

### 2.6 Comandos

`internal/realm/room/commands/moderation/{kick,mute,unmute,ban,unban,listbans}`. `ListMutes` existe como lectura tipada del service y en HTTP admin; no se inventa un comando de protocolo sin un packet real que lo consuma.

### 2.7 Packets

| Dirección | Paquete | Contenido |
| --- | --- | --- |
| Inbound | `moderation/kick` | `playerId int32` |
| Inbound | `moderation/mute` | `playerId int32`, `durationMinutes int32` |
| Inbound | `moderation/unmute` | `playerId int32` |
| Inbound | `moderation/ban` | `playerId int32`, `duration byte` (`BanDuration`) |
| Inbound | `moderation/unban` | `playerId int32` |
| Inbound | `moderation/listbans` | sin campos |
| Outbound | `moderation/kicked` | mandado al objetivo (motivo de salida) |
| Outbound | `moderation/muted` | mandado al objetivo, `endsAt`/`remainingSeconds` |
| Outbound | `moderation/banlist` | respuesta a `listbans` |

Headers confirmados contra Nitro: rights give `808`, remove `2064`, remove-all `2683`, relinquish `3182`, list request `3385`; kick `1320`, mute/unmute `3485`, ban `1477`, unban `992`, ban-list request `2267`; rights level `780`, owner `339`, rights list `1284`, add `2088`, remove `1327`, clear `2392`; remaining mute `826`, ban list `1869`, unbanned `3429`. **Sin campo de razón/mensaje** — el protocolo auditado no lo incluye y Pixels no lo inventa.

### 2.8 Eventos

```
moderation/kicked   {RoomID, TargetPlayerID, ActorID}
moderation/muted    {RoomID, TargetPlayerID, ActorID, DurationSeconds, ExpiresAt}
moderation/unmuted  {RoomID, TargetPlayerID, ActorID}
moderation/banned   {RoomID, TargetPlayerID, ActorID, DurationSeconds, ExpiresAt}
moderation/unbanned {RoomID, TargetPlayerID, ActorID}
```
Mismos dos consumidores independientes que en 1.5: un broadcaster (avisa al objetivo si está online, y fuerza su salida vía `leavecmd.Handler` en el caso de kick/ban) y el subscriber de auditoría de la Parte 3.5 — **`moderation/kicked` se audita exactamente igual que los otros cuatro**, ver 2.1: el kick no deja estado propio, pero sí deja su fila en `room_moderation_actions` como cualquier otra acción de moderación.

### 2.9 `IsBanned` / `IsMuted`

`IsBanned(ctx, roomID, playerID) (bool, error)` — `exists(select 1 from room_bans where room_id = $1 and player_id = $2 and ends_at > now())`. Tiene exactamente la firma de `entry.BanChecker` — se pasa directo a `entry.Service.WithBans(moderationService)` sin adapter (Parte 5).

`IsMuted(ctx, roomID, playerID) (bool, error)` — mismo shape de query contra `room_mutes`. **Sin ningún consumidor todavía** (el chat no existe en Pixels — mismo criterio ya usado en `plan/PERMISSIONS.md` con el `// TODO(chat): ...` para el color de grupo primario): se implementa igual en este plan porque la tabla/servicio ya están, pero queda como `// TODO(chat): enganchar IsMuted al pipeline de envío de mensajes una vez exista el realm de chat`.

### 2.10 Wiring

`entry.Service.WithBans(moderationService)` — mismo módulo que 1.6, ver Parte 5 para el detalle conjunto.

### 2.11 Tests

- Kick remueve al objetivo del room activo sin dejar ninguna fila de **estado** — pero SÍ inserta una fila persistida en `room_moderation_actions` (`action_type = 'kick'`), verificable consultando el historial inmediatamente después: un kick es tan trazable como un ban o un mute, solo que no tiene contraparte de "estado vigente".
- Mute con `durationMinutes` fuera de rango (`< Min` o `> Max`) rechaza sin persistir nada.
- Mute persiste con el `ends_at` correcto, y `IsMuted` refleja `true` hasta expirar.
- Ban con cada una de las 3 `BanDuration` persiste con el `ends_at` correcto (`Hour`/`Day`/`Permanent`).
- `Unmute`/`Unban` antes de tiempo actualiza `ends_at = now()`, `IsMuted`/`IsBanned` reflejan `false` de inmediato.
- **Fórmula de autorización (2.5), tabla de casos exhaustiva**:
  - Staff con nodo `Any` → autorizado, sin importar política de sala ni ser dueño.
  - Dueño con nodo `Own` → autorizado.
  - Dueño **sin** nodo `Own` (revocado puntualmente) → **no** autorizado, aunque sea el dueño.
  - Rights-holder, sala con política `OwnerAndRights`, con nodo `Own` → autorizado.
  - Rights-holder, sala con política `OwnerOnly` → **no** autorizado, aunque tenga el nodo `Own`.
  - Rights-holder con nodo `Own` revocado, sala con política `OwnerAndRights` → **no** autorizado.
  - Objetivo con `Unkickable` → **nunca** autorizado, incluso para un actor staff con nodo `Any`.
- Cada acción exitosa publica el evento correspondiente con `DurationSeconds`/`ExpiresAt` correctos (mute/ban) — el kick publica el suyo igual, sin duración, y también termina auditado (ver primer bullet de esta sección).
- `IsBanned`/`HasRights` integrados en un test end-to-end contra `entry.Service.Authorize` (confirma el acople real de la Parte 5, no solo el método aislado).

---

## Parte 3 — Auditoría e historial

### 3.1 Filosofía: append-only, nunca update/delete

Mismo criterio que `currency_ledger_entries` (ya establecido en el proyecto): una fila de auditoría, una vez insertada, **nunca** se actualiza ni se borra. Las tablas de **estado actual** (`room_rights`, `room_bans`, `room_mutes`, Partes 1-2) sí se mutan (eso es lo que las hace "estado actual" y no "historial") — las de auditoría son un log estrictamente creciente, independiente de qué pase después con el estado actual.

Una expiración natural por TTL (`ends_at` pasado) **no genera una fila propia** — la fila original de mute/ban ya tiene su `expires_at`, y una consulta de historial puede mostrar "expiró solo" simplemente comparando `expires_at` contra el momento de la consulta. Solo una acción **explícita** (`unmute`/`unban` manual, antes de tiempo) genera su propia fila de auditoría con su propio actor.

### 3.2 Esquema

```sql
create table room_rights_audit (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id),
    actor_kind text not null default 'player',
    actor_id bigint null references players(id),
    action text not null,
    created_at timestamptz not null default now(),
    constraint room_rights_audit_action_chk check (action in ('granted', 'revoked', 'revoked_all', 'relinquished'))
);

create index room_rights_audit_room_id_idx on room_rights_audit (room_id, created_at desc);
create index room_rights_audit_player_id_idx on room_rights_audit (player_id, created_at desc);
create index room_rights_audit_actor_id_idx on room_rights_audit (actor_id, created_at desc);

create table room_moderation_actions (
    id bigint generated always as identity primary key,
    room_id bigint not null references rooms(id) on delete cascade,
    target_player_id bigint not null references players(id),
    actor_kind text not null default 'player',
    actor_id bigint null references players(id),
    action_type text not null,
    duration_seconds integer null,
    expires_at timestamptz null,
    created_at timestamptz not null default now(),
    constraint room_moderation_actions_action_type_chk check (action_type in ('kick', 'mute', 'unmute', 'ban', 'unban'))
);

create index room_moderation_actions_room_id_idx on room_moderation_actions (room_id, created_at desc);
create index room_moderation_actions_target_player_id_idx on room_moderation_actions (target_player_id, created_at desc);
create index room_moderation_actions_actor_id_idx on room_moderation_actions (actor_id, created_at desc);
create index room_moderation_actions_action_type_idx on room_moderation_actions (action_type, created_at desc);
```

Cuatro índices en `room_moderation_actions`, uno por cada eje de consulta que el usuario pidió explícitamente (por room, por objetivo, por actor, por tipo de acción) — cada uno como `(columna, created_at desc)` para que la paginación por fecha use el índice directo sin un sort adicional en cada query.

`actor_kind`/`actor_id` mismo patrón ya establecido por `currency_ledger_entries`. `actor_kind = 'player'` es el único valor real usado hoy (una acción siempre la ejecuta un jugador — dueño, rights-holder, o staff); `'system'` queda reservado sin uso real todavía (las expiraciones de TTL no generan fila propia, 3.1) — la puerta queda abierta para una futura acción automática (ej. un auto-mod) sin cambiar el esquema.

### 3.3 Servicio de consulta

```go
// internal/realm/room/audit/service.go

// Query filters audit records along any combination of axes.
type Query struct {
    // RoomID optionally scopes results to one room.
    RoomID *int64
    // TargetPlayerID optionally scopes results to actions received by one player.
    TargetPlayerID *int64
    // ActorPlayerID optionally scopes results to actions executed by one player.
    ActorPlayerID *int64
    // ActionTypes optionally restricts which action kinds are returned.
    ActionTypes []string
    // Before excludes records at or after this id — keyset pagination cursor.
    Before *int64
    // Limit caps the returned record count.
    Limit int
}

// Manager reads room rights and moderation audit history.
type Manager interface {
    // ModerationHistory lists moderation actions matching query, newest first.
    ModerationHistory(ctx context.Context, query Query) ([]ModerationAction, error)
    // RightsHistory lists rights grants/revocations matching query, newest first.
    RightsHistory(ctx context.Context, query Query) ([]RightsAudit, error)
}
```

Cubre explícitamente cada ángulo que el usuario pidió:

| Necesidad pedida | Cómo se resuelve |
| --- | --- |
| Historial de mute/unmute de una room | `ModerationHistory(ctx, Query{RoomID: &roomID, ActionTypes: []string{"mute", "unmute"}})` |
| Historial de sanciones (kick/ban/unban) de una room | `ModerationHistory(ctx, Query{RoomID: &roomID, ActionTypes: []string{"kick", "ban", "unban"}})` |
| Historial de grant/removal de derechos de una room | `RightsHistory(ctx, Query{RoomID: &roomID})` |
| Acciones relacionadas a un usuario en una room puntual (como sancionado) | `ModerationHistory(ctx, Query{RoomID: &roomID, TargetPlayerID: &playerID})` |
| Acciones relacionadas a un usuario en **todas** las rooms (global, como sancionado) | `ModerationHistory(ctx, Query{TargetPlayerID: &playerID})` (sin `RoomID`) |
| Acciones **ejecutadas por** un usuario (como moderador), en una room o global | `ModerationHistory(ctx, Query{ActorPlayerID: &playerID})` (± `RoomID`) — pensado para responsabilidad de moderadores, no solo el lado del sancionado |

### 3.4 HTTP admin endpoints

```go
// pkg/http/room/routes/routes.go (extendido)

const (
    // playersPath stores the player-scoped admin base path — nuevo, `roomPath`/
    // `navigatorPath` ya existen en este archivo, `playersPath` no.
    playersPath = "/api/admin/players"
)

app.Get(roomPath+"/:id/rights/history", rightsHistoryHandler(audit))
app.Get(roomPath+"/:id/moderation/history", moderationHistoryHandler(audit))
app.Get(roomPath+"/:id/bans", activeBansHandler(moderation))
app.Get(roomPath+"/:id/mutes", activeMutesHandler(moderation))

app.Get(playersPath+"/:playerId/moderation/history", playerModerationTargetHandler(audit))   // ?roomId= opcional
app.Get(playersPath+"/:playerId/moderation/actions", playerModerationActorHandler(audit))    // ?roomId= opcional
```

| Método | Ruta | Efecto |
| --- | --- | --- |
| GET | `/api/admin/rooms/:id/rights/history` | Historial completo de grant/revoke/relinquish de esa room |
| GET | `/api/admin/rooms/:id/moderation/history?type=mute,unmute` | Historial de moderación de esa room, filtrable por `type` (uno o más de `kick,mute,unmute,ban,unban`) |
| GET | `/api/admin/rooms/:id/bans` | Baneos actualmente vigentes (`room_bans` con `ends_at > now()`, estado — no auditoría) |
| GET | `/api/admin/rooms/:id/mutes` | Mutes actualmente vigentes (ídem) |
| GET | `/api/admin/players/:playerId/moderation/history?roomId=` | Acciones donde el jugador es el **objetivo** — global si `roomId` se omite, acotado a una room si se manda |
| GET | `/api/admin/players/:playerId/moderation/actions?roomId=` | Acciones donde el jugador es el **actor/moderador** — global o por room |

Todos paginados con `limit` (default 50, tope 200) + `before` (id, keyset descendente) — mismo criterio simple ya usado por el resto de los endpoints de listado del proyecto (`ctx.QueryInt("limit", 50)` en `pkg/http/room/routes/handler.go`), sin introducir un esquema de paginación nuevo.

### 3.5 Quién escribe en las tablas de auditoría

Un subscriber de bus dedicado, `internal/realm/room/audit/subscriber.go`, suscripto a los 4 eventos de rights (1.5) y los 5 de moderación (2.8, kick incluido) — **nunca** `rights.Service`/`moderation.Service` insertando directo en `room_rights_audit`/`room_moderation_actions`. Mismo principio de desacople que ya usa el resto del proyecto para efectos secundarios (`currency/broadcast` reacciona a mutaciones de currency sin que `currency.Service` sepa que existe): permite agregar mañana OTRO consumidor del mismo evento (ej. un futuro sistema de detección de abuso de moderación, o una integración con un bot de Discord de staff) sin tocar `rights`/`moderation` para nada.

```go
// internal/realm/room/audit/subscriber.go

// RegisterAuditSubscriber wires audit persistence to room rights and moderation events.
func RegisterAuditSubscriber(subscriber bus.Subscriber, store Store) {
    subscriber.Subscribe(rightsgranted.Name, bus.PriorityNormal, handleRightsGranted(store))
    subscriber.Subscribe(rightsrevoked.Name, bus.PriorityNormal, handleRightsRevoked(store))
    subscriber.Subscribe(moderationkicked.Name, bus.PriorityNormal, handleModerationKicked(store))
    subscriber.Subscribe(moderationmuted.Name, bus.PriorityNormal, handleModerationMuted(store))
    subscriber.Subscribe(moderationbanned.Name, bus.PriorityNormal, handleModerationBanned(store))
    // ... unmuted/unbanned, mismo patrón
}
```

---

## Parte 4 — Nodos de permiso (extiende `internal/realm/room/permissions.go`, ya existente)

```go
// internal/realm/room/permissions.go (agregado a los 3 nodos ya existentes)

var (
    // ModerationOwnKick allows removing a player from a room you own.
    ModerationOwnKick = permission.RegisterNode("room.moderation.own.kick", "")
    // ModerationOwnMute allows temporarily muting a player's chat in a room you own.
    ModerationOwnMute = permission.RegisterNode("room.moderation.own.mute", "")
    // ModerationOwnBan allows banning a player from a room you own.
    ModerationOwnBan = permission.RegisterNode("room.moderation.own.ban", "")
    // ModerationAnyKick allows removing a player from any room, owned or not (staff).
    ModerationAnyKick = permission.RegisterNode("room.moderation.any.kick", "")
    // ModerationAnyMute allows muting a player in any room, owned or not (staff).
    ModerationAnyMute = permission.RegisterNode("room.moderation.any.mute", "")
    // ModerationAnyBan allows banning a player from any room, owned or not (staff).
    ModerationAnyBan = permission.RegisterNode("room.moderation.any.ban", "")

    // RightsOwnGrant allows granting build rights in a room you own.
    RightsOwnGrant = permission.RegisterNode("room.rights.own.grant", "")
    // RightsOwnRevoke allows revoking build rights in a room you own.
    RightsOwnRevoke = permission.RegisterNode("room.rights.own.revoke", "")
    // RightsAnyGrant allows granting build rights in any room (staff).
    RightsAnyGrant = permission.RegisterNode("room.rights.any.grant", "")
    // RightsAnyRevoke allows revoking build rights in any room (staff).
    RightsAnyRevoke = permission.RegisterNode("room.rights.any.revoke", "")

    // Unkickable protects a player from any moderation action, regardless of the actor's rights.
    Unkickable = permission.RegisterNode("room.unkickable", "")
)
```

11 nodos nuevos, exactamente los que `plan/PERMISSIONS.md` Parte 3.1 ya había previsto para este dominio — este plan es el que finalmente los declara y los consume de verdad.

---

## Parte 5 — Cierre del acople con `ENTRY.md`

Esto es, específicamente, "terminar la parte de `ENTRY.md`" que quedó abierta:

1. `rights.Service` implementa `entry.RightsChecker` con la firma exacta (`HasRights(ctx, roomID, playerID) (bool, error)`) — cero adapter necesario.
2. `moderation.Service` implementa `entry.BanChecker` con la firma exacta (`IsBanned(ctx, roomID, playerID) (bool, error)`) — cero adapter necesario.
3. `internal/realm/room/module.go` (extendido): tras construir `entry.Service` (ya existente) y los nuevos `rights.Service`/`moderation.Service`, encadena:
   ```go
   entryService.WithRights(rightsService).WithBans(moderationService)
   ```
   antes de exponerlo vía `fx.Provide` — un solo punto de wiring, sin tocar `entry` ni `enter.Handler` para nada.
4. Con esto, `entry.Service.Authorize` deja de degradar silenciosamente (`service.rights == nil` → `false`, `service.bans == nil` → `false`) y empieza a resolver rights/bans reales — sin haber cambiado una sola línea de `entry.Service` ni de `enter.Handler`, exactamente el punto de diseño que `ENTRY.md` dejó preparado a propósito.
5. `entry.Service.CanAnswerDoorbell` (ya implementado, ya usado por el flujo de timbre) también empieza a resolver `HasRights` real en vez de degradar a "sin rights" — un rights-holder (no solo el dueño) puede aceptar/rechazar en la puerta apenas este plan aterriza, sin ningún cambio en el código de doorbell.

---

## Parte 6 — Hot paths, allocations, benchmarks

- **`IsBanned` corre en TODO intento de entrada**, incluso a rooms `Open` (`entry.Service.checkBan` se llama primero en `Authorize`, antes del branch por `DoorMode`) — es el chequeo más caliente de todo este plan. La PK `(room_id, player_id)` resuelve el lookup puntual y `ends_at > now` filtra la única fila de estado actual. PostgreSQL no permite `now()` en el predicado de un índice parcial porque no es inmutable; por eso se usan índices `(room_id, ends_at desc)` y `(player_id, ends_at desc)` para listas activas.
- **`HasRights` corre en cada entrada a `DoorModeInvisible`/`DoorModeDoorbell`, y en cada acción de moderación** (2.5) — mismo criterio, PK compuesta `(room_id, player_id)` ya cubre el acceso directo sin necesitar un índice adicional.
- **Cache solo donde existe un hot path real** — entrada y autorización durable consultan PostgreSQL; al activarse una room, sus rights se proyectan en un map embebido dentro de `live.Room`. Furniture place/move/pickup consulta ese map en `O(1)` y los eventos grant/revoke lo mantienen actualizado después del commit. No existe un registry paralelo ni cache Redis.
- **La auditoría es síncrona y atómica**: `pkg/bus.Publisher.Publish` ejecuta subscribers localmente. El subscriber de auditoría tiene prioridad alta y usa el executor transaccional compartido; si el insert de auditoría falla, la mutación hace rollback. Los broadcasters de packets/runtime registran callbacks `postgres.AfterCommit`, de modo que nunca proyectan una mutación abortada.

Benchmarks nuevos (mismo patrón ya establecido por `internal/realm/room/world/grid/benchmark_test.go`/`world/furniture/benchmark_test.go`):

```go
// internal/realm/room/rights/benchmark_test.go

// BenchmarkHasRights measures the Service resolution cost against a fake in-memory
// repository — isolates the Service's own logic from Postgres round-trip cost.
func BenchmarkHasRights(b *testing.B) { ... }
```
```go
// internal/realm/room/moderation/benchmark_test.go

// BenchmarkIsBanned measures the Service resolution cost against a fake repository.
func BenchmarkIsBanned(b *testing.B) { ... }

// BenchmarkAuthorizeModerationAction measures the combined formula's cost (2.5) —
// permission checks + policy read + rights lookup — against fakes, no I/O.
func BenchmarkAuthorizeModerationAction(b *testing.B) { ... }
```
```go
// internal/realm/room/audit/benchmark_test.go

// BenchmarkModerationHistoryQuery measures the real effect of the Parte 3.2 indexes
// against a representatively-sized table (populated to ~10k rows in the benchmark's
// setup), run against a real test Postgres instance — same convention already used
// by internal/realm/room/repository's own tests, not mocked.
func BenchmarkModerationHistoryQuery(b *testing.B) { ... }

// BenchmarkAuditInsert measures the cost of one append-only audit write.
func BenchmarkAuditInsert(b *testing.B) { ... }
```

---

## Parte 7 — Testing (resumen transversal, detalle por parte ya cubierto arriba)

- Fakes de repository para `rights.Service`/`moderation.Service` (sin Postgres real) — mismo patrón ya establecido en `enter/command_test.go`/`enter/helpers_test.go`.
- Tests de repository reales contra Postgres de test para las 5 tablas nuevas (`room_rights`, `room_bans`, `room_mutes`, `room_rights_audit`, `room_moderation_actions`) — mismo patrón ya usado por `internal/realm/room/repository/repository_test.go`.
- Reloj inyectable (`func() time.Time`) en `moderation.Service` para testear expiración de mute/ban sin `time.Sleep` real — mismo criterio ya fijado en `ENTRY.md` Parte 7.
- Test de integración explícito: `entry.Service` con `WithRights`/`WithBans` reales (no fakes) resolviendo un caso de ban activo y un caso de rights activo de punta a punta — confirma el acople de la Parte 5, no solo cada pieza aislada.
- Test de humo del subscriber de auditoría (3.5): publicar cada uno de los 9 eventos (4 de rights + 5 de moderación, kick incluido) contra un `bus` real de test y confirmar que cada uno termina como una fila en la tabla correspondiente, con `actor_kind`/`actor_id`/`action_type` correctos.

---

## Parte 8 — Milestones de implementación

1. **RM1 — Esquema completo**: migraciones de `room_rights`, `room_bans`, `room_mutes`, `room_rights_audit`, `room_moderation_actions`; agregar `ModerationMute/Kick/Ban ModerationPolicy` a `model.Room` + su scan (columnas ya existentes en Postgres, hoy sin leer).
2. **RM2 — `internal/realm/room/rights`**: repository, `Service`/`HasRights`, comandos `grant/revoke/revokeall/list/relinquish`, packets, eventos, broadcaster — Parte 1 completa.
3. **RM3 — `internal/realm/room/moderation`**: repository, `Service`/`IsBanned`/`IsMuted`/`ListMutes`, la fórmula de autorización combinada (2.5), comandos y packets Nitro confirmados, eventos y broadcaster — Parte 2 completa. Depende de RM2 (la fórmula de 2.5 necesita `rights.Service.HasRights` para el caso "rights-holder").
4. **RM4 — Nodos de permiso**: agregar los 11 nodos de la Parte 4 a `internal/realm/room/permissions.go` — puede correr en paralelo a RM2/RM3, ambos lo consumen.
5. **RM5 — `internal/realm/room/audit`**: `Manager`/`Query` de consulta (3.3), subscriber (3.5) — depende de RM2/RM3 (necesita que los eventos ya existan).
6. **RM6 — HTTP admin de historial**: los 6 endpoints de la Parte 3.4 — depende de RM5.
7. **RM7 — Cierre del acople con `ENTRY.md`**: wiring `entryService.WithRights(...).WithBans(...)` en `internal/realm/room/module.go` (Parte 5) — depende de RM2 y RM3. Este es el milestone que específicamente termina el trabajo que `ENTRY.md` dejó abierto.
8. **RM8 — Benchmarks**: los 4 benchmarks de la Parte 6 — puede correr en paralelo a RM6, depende de RM1-RM3.

### Milestones futuros confirmados (fuera de este documento, no descartados)

- **`IsMuted` enganchado al pipeline de chat** — el servicio y la tabla ya existen desde RM3, pero no hay ningún consumidor real hasta que exista un realm de chat (`// TODO(chat)`, mismo criterio ya usado en `plan/PERMISSIONS.md`).
- **Excepción explícita a `Unkickable`** ("puedo moderar incluso a alguien unkickable si además tengo tal nodo") — no se agrega en este plan por falta de un caso real que lo justifique; se evalúa si en la práctica hace falta.
- **Sistema de detección de abuso de moderación** (ej. alertar si un moderador puntual ejecuta un volumen anómalo de bans en poco tiempo) — mencionado en 3.5 como un consumidor futuro plausible del mismo bus de eventos; no se construye sin una necesidad real confirmada.
