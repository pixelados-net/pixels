# Plan: Teleportadores (`internal/realm/furniture/interactions/teleport`)

Este plan profundiza completamente el **Milestone I6** que `plan/INTERACTIONS.md` Parte 8 dejó a nivel de boceto — cruzando ese diseño contra la máquina de estados real de un fork activo (**Polaris-Emulator**, `github.com/duckietm/Polaris-Emulator`, Java), leyendo directamente `InteractionTeleport.java` y las 5 clases `TeleportAction{One,Two,Three,Four,Five}.java` que implementan cada paso con sus delays exactos. El resultado es mucho más detallado que el boceto original: expone exactamente cómo se ve la animación de entrada y de salida, dos variantes reales (pad clickeable vs. tile caminable encadenable), y varios casos límite que `INTERACTIONS.md` no había capturado.

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding)

| Ya existe | Dónde | Nota |
| --- | --- | --- |
| `furniture.used`/`furniture.walkedon`/`furniture.walkedoff` (eventos scaffold) | `internal/realm/furniture/events/{used,walkedon,walkedoff}` | "Planned for future interactions" — nadie los publica todavía. Este plan es el primer consumidor real de `walkedon` (variante tile, Parte 3.1) y usa el mismo patrón de evento para `used` (variante pad). |
| `Definition.InteractionType string` | `internal/realm/furniture/model/definition.go` | El punto de enganche de datos ya existe — falta el código que lo lea para algo más que sit/lay. |
| `Item.ExtraData string` (persistente) | `internal/realm/furniture/model/item.go` | Ya existe en el modelo durable — la máquina de estados del teleportador (Parte 3) escribe ahí. |
| `world/furniture.Item` (runtime) — **sin `ExtraData`** | `internal/realm/room/world/furniture/item.go` | Gap real: el snapshot en memoria de un item colocado no carga su estado visual. Se extiende en este plan (Parte 2.3) — beneficia también al futuro Milestone I1 (toggle genérico), no es exclusivo de teleportadores. |
| `worldunit.Unit` — `goal`/`hasGoal`/`steps`/`statuses`/**`exiting bool`** | `internal/realm/room/world/unit/unit.go` | `exiting` ya modela "este movimiento lo controla el servidor y debe terminar en un tile específico sin que el jugador lo interrumpa" — exactamente el concepto que este plan generaliza para el tránsito de teleporte (Parte 2.4), en vez de agregar banderas sueltas nuevas como hace la referencia (`isTeleporting`/`isLeavingTeleporter`). |
| `worldunit.Status` (`mv`/`sit`/`lay`/`flatctrl`) | `internal/realm/room/world/unit/status.go` | Sistema de estados visuales ya genérico — este plan no agrega un mecanismo nuevo de "estado de unidad", reusa este. |
| `Room.LoadWorld`/Mecanismo B de rejoin directo | `internal/realm/room/runtime/live/world.go`, `plan/rooms/ENTRY.md` Parte 4.5/4.6 | Ya diseñado específicamente pensando en este caso — el teleportador es el consumidor que `ENTRY.md` dejó anotado como dependencia. |
| `entry.Service` — `Trusted`/`GrantTrusted`/nodo `room.EnterAny` | `internal/realm/room/access/entry/*`, `internal/realm/room/permissions.go` | Ya implementados — este plan los reusa para el bypass de gating del room destino (Parte 4), sin crear ningún mecanismo nuevo. |
| `Room.Tick()` a `DefaultTickInterval = 500ms` | `internal/realm/room/runtime/live/movement.go`, `model.go` | **Coincide exactamente** con el timing de 500ms por paso que la referencia real usa para cada fase del teleporte — este plan no necesita ningún scheduler/timer nuevo, avanza la máquina de estados del teleporte como un paso más del tick ya existente (Parte 3.6), mismo patrón que ya usa `SweepDoorbell`. |
| `furniture_item_teleport_pairs` (boceto) | `plan/INTERACTIONS.md` Parte 8 | Ya bosquejado — este plan lo precisa (Parte 5) y resuelve la pregunta que INTERACTIONS.md dejó abierta ("mecanismo de emparejamiento exacto a decidir en implementación", Parte 6). |

---

## Parte 1 — Lo que confirmé en Polaris-Emulator (Java real, no una descripción de segunda mano)

Leí `InteractionTeleport.java` completo y las 5 clases `TeleportActionOne` a `Five` (`com.eu.habbo.threading.runnables.teleport`). Esto es sustancialmente más detallado que el research que ya tenía `INTERACTIONS.md`:

### 1.1 Dos variantes reales, no una

- **`InteractionTeleport`** — el pad clásico, requiere click y estar parado exactamente encima; tiene "puerta" (extradata pasa por `"1"` con delays de 500ms entre fases).
- **`InteractionTeleportTile`** — una variante **sin click**, se dispara por `onWalkOn` (caminar encima alcanza), y sus delays son **0ms** en cada fase (no hay animación de puerta que esperar) — y lo más importante: al llegar, si el tile de destino tiene ENCIMA otro `InteractionTeleportTile` distinto del propio, se **re-dispara automáticamente** (`onWalkOn` de ese otro) — permite encadenar tiles de teleporte uno tras otro sin clicks intermedios.

Pixels adopta ambas como dos comportamientos del mismo `interaction_type` base (`teleport`), diferenciados por un flag de definición (`RequiresClick bool` o similar), no dos sistemas separados — comparten toda la máquina de estados de la Parte 3, solo cambia el disparador y los delays.

### 1.2 Guard contra doble-disparo y contra "montado en algo"

```java
public boolean canUseTeleport(GameClient client, Room room) {
    ...
    return habbo.getHabboInfo().getRiding() == null; // no podés teleportar montado en un pet/carruaje
}

public void startTeleport(Room room, Habbo habbo, int delay) {
    if (habbo.getRoomUnit().isTeleporting) { // ya está en tránsito, ignora el segundo click
        walkable = this.getBaseItem().allowWalk();
        return;
    }
    ...
}
```
Dos guards reales que `INTERACTIONS.md` no tenía: (1) no se puede usar un teleportador mientras se está montado sobre algo (pet/carruaje, si Pixels llega a tener ese concepto — se anota como guard reservado, no bloqueante si el concepto no existe todavía), (2) un segundo click mientras ya hay un tránsito en curso se ignora sin error.

### 1.3 Auto-reparación del emparejamiento, con detalle real de simetría rota

```java
if (targetTeleport.getTargetRoomId() != currentTeleport.getTargetRoomId()) {
    // el par ya no apunta de vuelta correctamente — limpiar AMBOS lados, no solo el actual
    currentTeleport.setTargetId(0); currentTeleport.setTargetRoomId(0);
    targetTeleport.setTargetId(0); targetTeleport.setTargetRoomId(0);
}
```
No es solo "si el par desapareció, animar localmente" (lo único que `INTERACTIONS.md` tenía) — Polaris además detecta **asimetría** (el item A apunta a B, pero B ya no apunta de vuelta a A porque lo emparejaron con C) y limpia la referencia de **ambos** lados, no solo el que se está usando. Pixels adopta esto: la limpieza de un par roto es una operación simétrica, nunca deja un lado con una referencia colgante mientras el otro ya se reemparejó.

### 1.4 Config real para bypassear el gating del room destino

```java
Emulator.getGameEnvironment().getRoomManager().enterRoom(
    this.client.getHabbo(), targetRoom.getId(), "", 
    Emulator.getConfig().getBoolean("hotel.teleport.locked.allowed"), 
    teleportLocation
);
```
**Hallazgo que refina `plan/rooms/ENTRY.md` Parte 4.6**: ese documento dejó "sin bypass por defecto... si en el futuro se decide que sí, es cambiar un booleano". Acá está confirmado que la referencia real **sí tiene ese booleano, y es configurable por el operador del hotel** (`hotel.teleport.locked.allowed`), no hardcodeado a ningún valor fijo. Este plan lo adopta como `PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED` (Parte 4) — no cambia el default de `ENTRY.md` (sigue siendo "no bypasea" salvo que el operador lo prenda explícitamente), pero corrige que la opción de prenderlo **ya está confirmada como un patrón real**, no una especulación.

### 1.5 Otros dos flags reales que `INTERACTIONS.md` no tenía

```java
public boolean allowWiredResetState() { return false; }
public boolean invalidatesToRoomKick() { return true; }
```
- Un teleportador **nunca** se resetea por un wired genérico de "reset state" (si/cuando Pixels tenga wired) — su estado lo controla exclusivamente su propia máquina, nada externo lo pisa.
- `invalidatesToRoomKick` — el teleportador invalida el mecanismo genérico de "kickear a un jugador atascado a un tile válido" mientras está en tránsito (evita que un sistema de recuperación general interrumpa una animación de teleporte en curso pensando que el jugador quedó atascado).

### 1.6 Cualquier visitante puede usar un teleportador ajeno — no requiere rights sobre el room

Confirmado leyendo `onClick`/`tryTeleport`: el único chequeo de autorización es `canUseTeleport` (guard de riding, Parte 1.2) — **no** se chequea `HasRights` sobre el room ni ninguna otra condición de permisos. Un teleportador es furniture de uso público dentro de la sala donde está parado, igual que cualquier furniture clickeable normal — cualquiera presente en la sala puede usarlo, sin importar si tiene derechos de construcción. Pixels adopta esto tal cual (Parte 7).

---

## Parte 2 — Arquitectura de Pixels

### 2.1 Interaction type + variante

```go
// internal/realm/furniture/interactions/teleport/definition.go

// Variant distinguishes the two real teleporter behaviors (Parte 1.1).
type Variant int

const (
    // VariantPad requires a click and standing exactly on the tile; animates a
    // "door" through extradata "0" → "1" → "2" → "0" with 500ms steps.
    VariantPad Variant = iota
    // VariantTile triggers by walking onto it, no click, 0ms steps, and chains
    // automatically into an adjacent teleport tile on arrival.
    VariantTile
)
```
Ambas comparten la misma tabla de emparejamiento (Parte 5) y la misma máquina de estados (Parte 3) — la diferencia es puramente el disparador (`furniture.used` vs `furniture.walkedon`, Parte 0) y la constante de delay (500ms vs 0ms).

### 2.2 Extender `world/furniture.Item` con `ExtraData` (mejora general, no exclusiva de teleport)

```go
// internal/realm/room/world/furniture/item.go (extendido)
type Item struct {
    ID         int64
    Definition Definition
    Point      grid.Point
    Z          grid.Height
    Rotation   worldunit.Rotation
    // ExtraData stores the item's current protocol-facing visual state — mirrors
    // the persistent model's ExtraData, but lives in the live snapshot so it can
    // change every 500ms during a teleport sequence without a Postgres round-trip
    // per frame (Parte 3.6). Useful beyond teleport: the future toggle/gate
    // milestones (I1/I2, INTERACTIONS.md) need exactly the same primitive.
    ExtraData string
}
```

```go
// internal/realm/room/runtime/live/world.go (extendido)

// SetFurnitureExtraData updates one item's visual state without touching its
// footprint/fixtures — no ReplaceFixtures call, since extradata alone never
// changes height/stacking for a teleporter (mejora de costo sobre el genérico
// ReloadFurniture, que sí reconstruye fixtures completos).
func (room *Room) SetFurnitureExtraData(itemID int64, value string) bool {
    room.mutex.Lock()
    defer room.mutex.Unlock()
    if room.world == nil {
        return false
    }

    return room.world.SetFurnitureExtraData(itemID, value)
}
```
Persistencia: el valor final "en reposo" (`"0"`) se escribe a Postgres de forma asíncrona best-effort (mismo criterio ya usado para logging en `CHAT.md` — no bloquea la animación esperando el write), los estados intermedios (`"1"`, `"2"`) **no** se persisten individualmente — si el proceso cae a mitad de una animación, el peor caso es que el item queda con el último valor persistido (`"0"` de la vez anterior), nunca un estado intermedio inconsistente, y el próximo `Room.LoadWorld` (recarga de room) simplemente no muestra ninguna animación colgada.

### 2.3 Estado de tránsito de la unidad: generalizar `exiting`, no clonar banderas sueltas

En vez de las 3 banderas ad-hoc de la referencia (`isTeleporting`, `isLeavingTeleporter`, override de `canLeaveRoomByDoor`), Pixels generaliza el `exiting bool` que `worldunit.Unit` **ya tiene** ("movimiento controlado por el servidor que debe terminar en un tile específico"):

```go
// internal/realm/room/world/unit/unit.go (extendido)

// ControlKind names why a unit's movement is currently server-controlled,
// generalizing the single "exiting" boolean into a small enum — teleport transit
// is conceptually the same "don't let anything else interrupt this" lock that
// exiting-to-the-door already models, just with a different terminal action.
type ControlKind uint8

const (
    ControlNone ControlKind = iota
    ControlExitingRoom
    ControlTeleporting
)
```
Mientras `Control() != ControlNone`, el pathfinder/comandos de movimiento normal (`walk`) rechazan cualquier goal nuevo del propio jugador (mismo criterio que ya aplica hoy a `exiting`) — sin necesitar un segundo mecanismo de "está ocupado" paralelo.

### 2.4 Dónde vive el estado de la secuencia en curso

```go
// internal/realm/furniture/interactions/teleport/transit.go

// Transit tracks one unit's in-progress teleport sequence — lives on the active
// Room (roomlive.Room gains a `teleports map[int64]*Transit` keyed by player id,
// lazily allocated exactly like `doorbell`, Parte 0/ENTRY.md), not on the
// persistent furniture item, since it is pure session state.
type Transit struct {
    PlayerID       int64
    SourceItemID   int64
    TargetItemID   int64
    TargetRoomID   int64
    Phase          Phase // A..F, Parte 3
    PhaseDeadline  time.Time
    Variant        Variant
}
```

---

## Parte 3 — La máquina de estados completa, fase por fase (lo que pidió el usuario: cómo se ve entrar y salir)

Seis fases, confirmadas 1:1 contra `TeleportAction{One..Five}.java` — cada una con su animación exacta.

### 3.1 Fase A — Acercamiento (disparo + caminata previa)

- **Variante pad**: click sobre el teleportador. Si la unidad no está exactamente sobre el tile ni en el tile-en-frente, primero camina al tile-en-frente (reusa el pathfinder ya existente, `worldpath`) y, al llegar, reintenta automáticamente (recursión de un solo nivel, no un loop).
- Al llegar al tile-en-frente (o si ya estaba ahí): el tile del teleportador se marca temporalmente caminable (override, para permitir pisarlo aunque normalmente no lo sea) y la unidad camina el último paso hasta pararse exactamente encima — con `Control = ControlTeleporting` ya activo desde este punto (equivalente a `setCanLeaveRoomByDoor(false)`), así ninguna otra interrupción (salir por la puerta, otro comando de movimiento) puede cortar la secuencia a mitad de camino.
- **Variante tile**: no hay fase de acercamiento — `onWalkOn` dispara directo la Fase B en cuanto el pathfinder normal (sin ninguna lógica especial) hace que la unidad pise el tile como parte de un movimiento cualquiera.
- Guard: si `Transit` ya existe para ese jugador (ya en tránsito), un segundo click/paso se ignora sin error (Parte 1.2).
- Guard: si la unidad está "montada" (concepto reservado, Parte 1.2), rechaza sin iniciar nada.

### 3.2 Fase B — "La puerta se abre" (extradata `"1"`, 500ms / 0ms)

- El item de origen pasa a extradata `"1"` — visualmente, el teleportador "se abre" o se ilumina (el asset del cliente Nitro decide el aspecto exacto; el servidor solo cambia el número).
- La unidad se reposiciona exactamente sobre el tile del teleportador, **mirando hacia adentro** (rotación = rotación del mueble + 4, es decir, encarada en sentido opuesto a como el mueble mira hacia la sala) — y se le aplica un status `mv` (movimiento) puramente ilusorio: no se mueve a ningún lado nuevo, pero el cliente reproduce la animación de "dar un paso" hacia el interior del teleportador.
- Duración: 500ms (pad) / 0ms (tile, sin puerta que animar).

### 3.3 Fase C — Resolución del destino (limpia el status de movimiento, decide si hay para dónde ir)

- Se quita el status `mv` (la "animación de paso" termina).
- Se resuelve el par: lee `TargetItemID`/`TargetRoomID` cacheados en el item; si el item emparejado ya no existe, o la referencia resultó asimétrica (Parte 1.3), limpia AMBOS lados y vuelve a resolver desde la tabla de pares (Parte 5).
- El item de origen vuelve a extradata `"0"` de inmediato (la puerta ya "hizo su parte" del lado de origen).
- **Sin destino válido** → salta directo a la Fase F (Parte 3.5) con un `Transit` marcado "sin reubicación" — la unidad simplemente термина la animación de vuelta, sin cambiar de lugar. Nunca un error duro, mismo criterio que `INTERACTIONS.md` ya tenía.
- **Con destino válido** → programa, en 500ms/0ms más: (a) un flash del item de origen a `"2"` seguido de vuelta a `"0"` 1000ms después (puramente cosmético, el "clic" final de la puerta cerrándose del lado de origen, que el jugador ya no ve porque para entonces está del otro lado), y (b) la Fase D.

### 3.4 Fase D — El cruce real (reubicación instantánea, sin pathfinding)

- La unidad se coloca **directamente** (sin calcular ningún camino, es un teletransporte, no una caminata) sobre el tile exacto del teleportador emparejado, con su `Z` ajustado a la altura de ese tile.
- **Si el room destino es distinto del actual**: se usa `ROOM_FORWARD`, porque Nitro necesita ejecutar su ciclo normal de desmontaje, navegación y carga del nuevo renderer. Antes de enviarlo se registra un destino efímero de un solo uso `{playerID, roomID, targetItemID}`. El evento síncrono `room.entered` consume ese destino y coloca la unidad sobre el teleportador antes de proyectar las entidades del nuevo room. Así el cliente y los observadores reciben la misma posición autoritativa sin reemplazar estado server-side debajo de un renderer viejo.
- El item de destino pasa a extradata `"2"` — la "puerta" del otro lado ya está mostrándose abierta/activa justo cuando la unidad aparece ahí (el jugador nunca ve el destino "cerrado" en el instante exacto de llegar).
- Rotación de la unidad al llegar: igual a la rotación del teleportador destino (queda mirando hacia afuera de él, en la dirección natural de salida).
- Duración hasta la Fase E: 500ms (pad) / 0ms (tile).

### 3.5 Fase E-F — Asentamiento y salida caminando (la animación de "cómo sale del otro lado" que pidió el usuario)

- **Fase E** (breve, transición): marca el `Transit` como "saliendo" (equivalente a `isLeavingTeleporter`), simplemente para que la Fase F sepa que debe reproducir la caminata de salida.
- **Fase F**: el item de destino se mantiene en `"1"` (abierto) mientras se calcula el tile-en-frente del teleportador destino y la unidad **camina automáticamente** ahí (mismo pathfinder normal, no un teletransporte) — esta es la animación visible de "salir caminando del teleportador hacia la sala". Al terminar esa caminata (falle o no — ambos casos restauran el estado normal): se libera `Control` (vuelve a `ControlNone`, la unidad recupera control de movimiento normal), y el item de destino programa su propio regreso a `"0"` 1000ms después (la puerta del destino "se cierra sola").
- **Encadenamiento (solo variante tile)**: si el tile exacto donde la unidad terminó de aparecer (antes de caminar hacia afuera) tiene ENCIMA otro teleportador variante-tile distinto del que se acaba de usar, se dispara automáticamente su propia Fase A/B — permite una fila de tiles de teleporte consecutivos sin que el jugador tenga que caminar entre ellos.
- Si no hay un tile-en-frente válido (ej. el teleportador destino está pegado contra una pared sin salida), se salta la caminata y se restaura el control de movimiento igual, sin mover a nadie más.

### 3.6 Todo esto corre sobre el tick existente, no un scheduler nuevo

Cada `Transit` activo guarda su `PhaseDeadline` (un `time.Time` calculado al entrar a cada fase); `Room.Tick()` (ya existente, 500ms, `DefaultTickInterval`) gana un paso más, análogo al ya existente `SweepDoorbell`: recorre los `Transit`s activos y avanza cualquiera cuyo `PhaseDeadline` ya pasó. Como el intervalo del tick (500ms) coincide exactamente con el delay real confirmado en Polaris, cada fase avanza en, como mucho, un tick de diferencia — sin necesitar ningún `time.AfterFunc`/timer independiente por transición, mismo principio ya aplicado al timeout de hangout de `ENTRY.md`.

---

## Parte 4 — Bypass de gating del room destino: configurable, no fijo

```go
// internal/realm/furniture/interactions/teleport/config.go

// Config controls cross-room teleport behavior.
type Config struct {
    // BypassLockedRoom mirrors the confirmed real hotel.teleport.locked.allowed —
    // when true, a cross-room teleport ignores the target room's door mode
    // (password/doorbell/invisible) via entry.Command.Trusted. Default false: a
    // teleporter's target room gating applies exactly like a normal entry, unless
    // the hotel operator explicitly opts in.
    BypassLockedRoom bool `env:"PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED" envDefault:"false"`
}
```
El **ban nunca se bypasea** por este flag — mismo criterio ya fijado en `ENTRY.md` Parte 4.2/4.5 (un ban es "no quiero verte acá bajo ninguna circunstancia", conceptualmente distinto del control de acceso social que sí tiene sentido que un operador decida relajar para sus propios teleportadores). Si en algún momento hace falta bypasear también el ban, la única vía es que el jugador tenga `room.EnterAny` de verdad — nunca este flag por sí solo.

**Nota de actualización cruzada**: `plan/rooms/ENTRY.md` Parte 4.5 dice hoy "sin bypass por defecto... si en el futuro se decide que sí, es cambiar un booleano" — este plan confirma que esa decisión ya tiene precedente real confirmado (Parte 1.4) y le da nombre concreto (`BypassLockedRoom`/`PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED`). Vale la pena reflejar este nombre en `ENTRY.md` cuando se implemente, para que ambos documentos usen el mismo término.

---

## Parte 5 — Esquema: emparejamiento

```sql
create table furniture_item_teleport_pairs (
    item_one_id bigint not null references furniture_items(id) on delete cascade,
    item_two_id bigint not null references furniture_items(id) on delete cascade,
    created_at timestamptz not null default now(),
    unique (item_one_id, item_two_id),
    constraint furniture_item_teleport_pairs_distinct_chk check (item_one_id <> item_two_id)
);
create index furniture_item_teleport_pairs_item_two_id_idx on furniture_item_teleport_pairs (item_two_id);
```
Resolución **perezosa** (no cacheada en la fila del item, mismo criterio que la referencia confirmó): al necesitar el destino de un item, se consulta esta tabla en cualquier dirección (`item_one_id = X or item_two_id = X`) y se toma el otro lado — si no hay fila, no hay par, se resuelve como "sin destino" (Fase C). La limpieza de un par roto/asimétrico (Parte 1.3) borra la fila de la tabla directamente, no dos columnas sueltas en cada item — al ser una tabla de relación pura, "romper el par" es simplemente `delete` de esa fila, sin ningún estado duplicado que pueda desincronizarse entre los dos lados (una mejora estructural real sobre la referencia, que sí guarda `targetId`/`targetRoomId` por separado en cada item y por eso necesita el chequeo de asimetría de la Parte 1.3 en primer lugar — con una tabla de relación, la asimetría **no puede existir** por diseño).

---

## Parte 6 — Emparejamiento: cómo se crean los pares (resuelve la pregunta abierta de `INTERACTIONS.md` Parte 8)

`INTERACTIONS.md` dejaba esto "a decidir en implementación". Este plan lo resuelve:

- **No** se ata al flujo de catálogo (`CATALOG.md` no necesita ninguna noción de "oferta que crea 2 instancias emparejadas") — evita acoplar el sistema de compras a un caso especial de furniture.
- Comando dedicado, dueño-del-room-only (o rights-holder, mismo criterio que el resto de acciones de gestión de furniture en la sala): `internal/realm/furniture/interactions/teleport/commands/pair` — recibe dos `ItemID` ya colocados en (potencialmente distintas) salas propias del mismo jugador, valida que ambos sean definiciones `interaction_type = "teleport"`, que ninguno ya esté emparejado (un teleportador solo puede tener un par activo a la vez — emparejar de nuevo reemplaza el par anterior, dejando al viejo compañero sin destino, mismo criterio que la limpieza de la Parte 1.3), y crea la fila.
- **Cross-room desde el diseño**: los dos items de un par no necesitan estar en la misma sala — de hecho ese es el caso de uso real (viajar de una sala a otra del mismo jugador, o eventualmente entre salas de distintos jugadores si el dueño de ambas lo permite — fuera de alcance inicial, ver "Milestones futuros").

---

## Parte 7 — Permisos y validaciones

- **Usar un teleportador**: cualquiera presente en la sala, sin requerir rights de construcción (Parte 1.6) — solo el guard de "no montado" (Parte 1.2, reservado hasta que exista el concepto de monturas).
- **Emparejar/desemparejar**: dueño de AMBOS items (o rights-holder + `room.SettingsOwnManage`/staff con `.any`, mismo criterio ya establecido para gestión de furniture en el resto del proyecto) — no se requiere ser dueño del ROOM donde está parado un visitante que simplemente lo usa, solo de la furniture al emparejarla.
- **Bypass de room destino** (Parte 4): decisión de configuración del hotel, no de permiso por jugador — si en la práctica hace falta un nodo de permiso además del flag global (ej. "mis teleportadores sí bypasean, los de otros dueños no"), se anota como milestone futuro, sin evidencia real todavía de que haga falta esa granularidad.

---

## Parte 8 — Eventos (`pkg/bus`)

```
furniture.teleport.started   {PlayerID, SourceItemID, RoomID}
furniture.teleport.completed {PlayerID, SourceItemID, TargetItemID, SourceRoomID, TargetRoomID}
furniture.teleport.failed    {PlayerID, SourceItemID, Reason} // par roto/sin destino — Fase C
```
Mismo patrón de desacople ya usado en el resto del proyecto (`RIGHTS.md`'s subscriber de auditoría, `CHAT.md`'s historial) — un futuro consumidor (ej. estadísticas de uso de furniture, o un log de movimiento entre salas) se suscribe sin que la máquina de estados de la Parte 3 sepa que existe.

---

## Parte 9 — Packets

| Dirección | Paquete | Contenido | Header |
| --- | --- | --- | --- |
| Inbound | `furniture/interact` (compartido con el futuro Milestone I1, no dedicado a teleport) | `itemId int32` | TBD |
| Outbound | `room/furniture/state` (ya diseñado en `INTERACTIONS.md` Parte 3, reusado tal cual) | `itemId int32`, `state string` (los valores `"0"`/`"1"`/`"2"` de la Parte 3) | TBD |
| Outbound | `room/entities/status` (ya existente, reusado) | status `mv` ilusorio de la Fase B/caminata de salida de la Fase F | ya existente |
| Outbound | `ROOM_FORWARD` | Inicia el ciclo nativo de navegación de Nitro para cruces entre salas; el destino efímero se consume antes del bootstrap del room | `160` |

El teleportador **no necesita ningún packet nuevo dedicado** más allá de lo que el futuro Milestone I1 (click genérico) y el sistema de entrada ya existente proveen — toda la "magia" es orquestación server-side sobre primitivas de protocolo que ya están diseñadas en otros documentos.

---

## Parte 10 — Hot paths y allocations

- **El paso de tick de teleportes (Parte 3.6) es O(tránsitos activos)**, no O(ocupantes) — se espera un puñado como mucho en cualquier sala en un instante dado (usar un teleportador es una acción deliberada y poco frecuente comparada con chat/movimiento), mismo argumento de "no es un hot path" ya usado en `UPVOTES.md`.
- **`SetFurnitureExtraData` (2.2) evita reconstruir fixtures** para algo que no cambia footprint/altura — más barato que el `ReloadFurniture` genérico ya existente, a propósito.
- **El cruce de sala (Fase D) reusa exactamente el Mecanismo B ya optimizado en `ENTRY.md`** — no se introduce ningún costo nuevo de allocation ahí, es la misma función interna que ya usa `room.enter`.

---

## Parte 11 — Tests

- Click estando ya sobre el tile → arranca directo la Fase B (sin caminata previa).
- Click estando lejos → camina al tile-en-frente, luego al tile exacto, luego arranca la Fase B (sin que el jugador tenga que volver a clickear).
- Segundo click mientras ya hay un `Transit` activo → no-op, sin duplicar la secuencia.
- Par válido en la misma sala → Fase D reubica sin cambiar de room.
- Par válido en otra sala → Fase D envía `ROOM_FORWARD`; `room.entered` consume el destino de un solo uso y fija el spawn antes del bootstrap.
- Par roto (item destino eliminado) → salta a Fase F sin reubicar, sin error duro; y si la relación era asimétrica, ambos lados quedan limpios (no solo el usado).
- `BypassLockedRoom = false` (default) → un teleport hacia una sala con `DoorModePassword` falla el cruce igual que una entrada normal; `= true` → entra igual que un `Trusted: true` normal, salvo ban.
- Variante tile: caminar encima dispara la secuencia sin click, con delays en 0; aparecer sobre OTRO tile de teleporte variante-tile lo encadena automáticamente.
- El status `mv` ilusorio de la Fase B se limpia correctamente al entrar a la Fase C, sin quedar "pegado".
- `Room.Tick()` avanza cada `Transit` activo según su `PhaseDeadline`, con un reloj inyectable (mismo patrón que el resto del proyecto), sin depender de `time.Sleep` real en los tests.

---

## Parte 12 — Milestones de implementación

1. **TP1 — Esquema de emparejamiento** (Parte 5): migración de `furniture_item_teleport_pairs`.
2. **TP2 — `world/furniture.Item.ExtraData` + `SetFurnitureExtraData`** (Parte 2.2): extensión general, útil también para I1/I2 futuros — sin dependencias.
3. **TP3 — `ControlKind` en `worldunit.Unit`** (Parte 2.3): generaliza `exiting` — depende de TP2 solo en el sentido de que ambos tocan el mismo paquete `world`, no hay dependencia lógica real.
4. **TP4 — Comando de emparejamiento** (Parte 6-7): depende de TP1.
5. **TP5 — Máquina de estados completa (Fases A-F)** (Parte 3): depende de TP2, TP3, TP4, y de que exista al menos el disparo básico de click/walkedon (comparte el punto de enganche con el futuro Milestone I1 de `INTERACTIONS.md`, pero no bloquea a esperar que I1 esté terminado — el teleportador puede tener su propio manejador de click dedicado desde el principio, y I1 lo reusa/generaliza después si conviene).
6. **TP6 — Integración cross-room** (Parte 3.4, 4): depende de TP5 y de `entry.Service`/`Room.LoadWorld` ya existentes (sin trabajo adicional ahí, son consumidores, no cambios).
7. **TP7 — Config del bypass** (Parte 4): puede ir en paralelo a TP6.
8. **TP8 — Variante tile + encadenamiento** (Parte 1.1, 3.5): depende de TP5, se agrega después del pad clásico.

### Milestones futuros confirmados (fuera de este documento, no descartados)

- **Guard de "montado en algo"** (Parte 1.2) — reservado hasta que Pixels tenga un concepto de monturas/pets-que-se-cabalgan; el guard se agrega ahí, sin rediseñar nada de este plan.
- **Emparejar teleportadores entre dueños distintos** (Parte 6) — hoy se asume que ambos lados de un par pertenecen al mismo jugador; si aparece un caso de uso real (ej. una red pública de teleportadores del hotel), se decide ahí si hace falta un permiso especial.
- **Nodo de permiso granular para el bypass de sala destino** (Parte 7) — hoy es una decisión de configuración global del hotel, no por jugador; se agrega si aparece evidencia real de que hace falta esa granularidad.
