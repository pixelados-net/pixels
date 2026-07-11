# Plan: Floor Plan Editor (`internal/realm/room/control/floorplan`)

Implementa completamente la **Parte 6** de `plan/REMAINING-ROOMS.md` — el dueño diseña su propio heightmap en vez de usar solo layouts fijos (`model_a`, etc.). Este es el plan más grande de la serie de `plan/rooms/*`, y el que más se benefició de cruzar el diseño original contra una implementación real: leí directamente el código fuente de **Polaris-Emulator** (`github.com/duckietm/Polaris-Emulator`, Java, fork activo y mejorado de Arcturus) — `FloorPlanEditorSaveEvent.java`, `FloorPlanEditorRequestDoorSettingsEvent.java`, `FloorPlanEditorRequestBlockedTilesEvent.java`, sus composers de salida, y `CustomRoomLayout.java` — y encontré features reales que `REMAINING-ROOMS.md` no había capturado, además de un punto donde el diseño de Pixels puede mejorar deliberadamente sobre la referencia (no solo igualarla).

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding, confirmado leyendo el código actual)

| Ya existe | Dónde | Cómo lo usa este plan |
| --- | --- | --- |
| `Room.LoadWorld(config) error` — "loads **or replaces**" | `internal/realm/room/runtime/live/world.go` | **La pieza clave de este plan**: ya soporta reemplazar el mundo cargado de una sala activa sin destruir el `Room` (ocupantes, doorbell, mutes siguen intactos) — reconstruye el `World` desde cero y re-agrega cada ocupante actual como unit nuevo en la puerta. Esto es lo que permite refrescar en vivo sin forward forzado (Parte 7). |
| `world/grid.Parse(heightmap, opts...)` | `internal/realm/room/world/grid/parser.go` | Ya valida: normalización de fin de línea, longitud de fila pareja (`ErrIrregularRows`), charset (`0-9`, `a-z`/`A-Z`, `x`/`X`), puerta dentro del mapa y no sobre tile inválido. **No** impone un máximo de 64×64 — ese es un límite de negocio del editor, no del parser compartido (Parte 5.1), así que se agrega en la capa de este plan, no tocando `grid`. |
| `world/items.LoadRoomFurniture`/`ResolveWorldItem` | `internal/realm/room/world/items/{loader.go,placement.go}` | Ya resuelve definiciones + footprints de furniture colocada — reusado para calcular qué tiles están ocupados (Parte 5.3), sin duplicar esa consulta. |
| `world/layout.{Layout,Service,Store}` | `internal/realm/room/world/layout/*` | Hoy solo resuelve layouts **fijos por nombre** (`FindByName`) — este plan le agrega un segundo camino de resolución por `room_id` cuando exista una fila en `room_custom_layouts` (Parte 4), mismo patrón que ya usa, no un sistema paralelo. |
| `control/settings.Authorizer` | `internal/realm/room/control/settings/authorization.go` | Patrón de autorización (`AnyManage` global, o dueño + `OwnManage`) — este plan crea su propio `Authorizer` equivalente con nodos de floorplan (Parte 8), mismo shape, no lo reusa directo porque son permisos distintos (`REMAINING-ROOMS.md` ya preveía un permiso propio para el editor, no el de settings general). |
| `control/commands/settings/save.go` | mismo paquete | Plantilla exacta de wiring a copiar: `control.Actor`+`control.MatchRoom`, autorizar, mutar, refrescar runtime, broadcastear, confirmar al que guardó, publicar evento. |
| `roomsession`/`control.Actor` | `internal/realm/room/control/commands/resolve/session.go` | Reusado para resolver actor + room actual en los 3 comandos nuevos. |
| `permissions.go` — 15 nodos ya existentes | `internal/realm/room/permissions.go` | Ninguno de floorplan todavía — este plan agrega `FloorplanOwnEdit`/`FloorplanAnyEdit` al mismo archivo (Parte 8). |
| Furniture pickup/inventario ya implementado | `FURNITURE-INVENTORY.md` (ya implementado) | Reusado para el auto-pickup (Parte 5.4, feature nueva adoptada de Polaris) — la furniture bloqueante se devuelve al inventario de su dueño real con el mismo mecanismo que ya usa un pickup normal, no uno especial. |

---

## Parte 1 — Lo que confirmé en Polaris-Emulator (código real, no Arcturus original)

Leí `FloorPlanEditorSaveEvent.java` completo (el handler de guardado), `FloorPlanEditorRequestDoorSettingsEvent.java`, `FloorPlanEditorRequestBlockedTilesEvent.java`, sus composers de salida (`FloorPlanEditorDoorSettingsComposer`, `FloorPlanEditorBlockedTilesComposer`), y `CustomRoomLayout.java`. Esto confirma y **extiende** lo que `REMAINING-ROOMS.md` ya tenía:

### 1.1 Constantes exactas confirmadas (no estimadas)
```java
public static int MAXIMUM_FLOORPLAN_WIDTH_LENGTH = 64;
public static int MAXIMUM_FLOORPLAN_SIZE = 64 * 64; // 4096 caracteres totales
private static final int SAVE_COOLDOWN_SECONDS = 3;
private static final Pattern ALLOWED_MAP_CHARS = Pattern.compile("[a-zA-Z0-9\r]+");
```
- Ancho/alto máximo: **64×64** (confirmado, no una estimación de `REMAINING-ROOMS.md`).
- Rotación de puerta: **0-7** (confirmado).
- **Thickness de pared/piso: rango `-2` a `1`**, no `0` a un máximo — valores negativos representan modos especiales (auto/oculto), no un simple grosor creciente. `REMAINING-ROOMS.md` no tenía este rango exacto.
- **Campo nuevo que `REMAINING-ROOMS.md` no capturó**: `wallHeight` (altura fija de pared, `-1` a `15`, donde `-1` = auto-calculada según el mapa) — una feature real que Polaris agrega sobre Arcturus base. Se incorpora al esquema de Pixels (Parte 3).
- **Cooldown de guardado: 3 segundos** entre saves del mismo jugador — no estaba en el diseño original, se adopta (Parte 5.5).

### 1.2 Feature real no capturada en `REMAINING-ROOMS.md`: auto-pickup de furniture bloqueante
En vez de solo "fallar si un tile bloqueado tiene furniture", el guardado acepta un flag `autoPickup boolean`: si viene activado, la furniture que bloquearía el cambio se **recoge automáticamente y se devuelve al inventario de su dueño real** (no necesariamente el dueño de la sala — cada item vuelve a quien lo posee), con un tope (`MAX_AUTO_PICKUP_ITEMS = 500`) para que un save no evacúe de golpe una cantidad absurda de items por error. Sin el flag, el comportamiento es el original: rechazar indicando la primera coordenada bloqueada.

### 1.3 Detalle de validación que `REMAINING-ROOMS.md` simplificaba de más
Polaris no solo chequea "¿el tile nuevo es `x` y tiene furniture encima?" — también chequea **cualquier tile cuya altura CAMBIE** contra la furniture existente (`height != tile.z && room.getTopItemAt(x, y) != null`), no solo los tiles que pasan a ser inválidos. Es decir: subir o bajar el piso debajo de un mueble ya colocado también bloquea (o dispara auto-pickup), no solo borrar el tile por completo. Este plan adopta esta validación más completa (Parte 5.3).

### 1.4 "Tiles bloqueados" se precalcula y se expone ANTES de guardar
`FloorPlanEditorRequestBlockedTilesEvent` → `FloorPlanEditorBlockedTilesComposer` manda la lista completa de tiles actualmente ocupados por furniture **al abrir el editor**, no solo como respuesta de error al guardar — el cliente puede marcarlos visualmente de entrada, antes de que el jugador intente encogerlos. Este plan reusa la misma consulta para ambos momentos (Parte 5.3/6.1).

### 1.5 Confirmación importante: incluso un fork moderno sigue usando "unload + forward forzado"
```java
Emulator.getGameEnvironment().getRoomManager().unloadRoom(room);
room = Emulator.getGameEnvironment().getRoomManager().loadRoom(room.getId());
ServerMessage message = new ForwardToRoomComposer(room.getId()).compose();
for (Habbo habbo : habbos) habbo.getClient().sendResponse(message);
```
Polaris — un fork activamente mantenido y mejorado, no el Arcturus original de 2015 — **sigue** descargando la sala completa de memoria y mandando un forward forzado a cada ocupante para que reingresen. Esto confirma que no es una limitación superada en la práctica por otros emuladores: es la forma estándar de resolverlo cuando el motor no separa bien world-state de connection-state. **Pixels sí puede hacerlo mejor** (Parte 2/7), precisamente porque `Room.LoadWorld` ya soporta reemplazo en caliente sin destruir el objeto `Room` — una mejora real confirmada contra la referencia, no una suposición.

### 1.6 Validación de "altura efectiva 0", configurable
```java
if (Emulator.getConfig().getBoolean("hotel.room.floorplan.check.enabled")) {
    if (map.replace("x", "").replace("\r", "").isEmpty()) { errors.add(...); }
}
```
Un mapa 100% inválido (todo `x`) se puede rechazar, pero es un chequeo **configurable por el operador del hotel**, no fijo. Pixels adopta el mismo criterio (Parte 5.2).

---

## Parte 2 — Mejoras deliberadas de este plan sobre la referencia (Arcturus Y Polaris, no solo paridad)

1. **Refresco en caliente, sin forward forzado** (Parte 7) — `Room.LoadWorld` ya permite reemplazar el mundo de una sala activa sin destruir el objeto `Room` ni desconectar a nadie. Ni el Arcturus original ni Polaris (2026, activamente mantenido) lo hacen así — ambos descargan y fuerzan reconexión. Pixels puede evitarlo porque su motor ya separa mejor world-state (`World`) de connection-state (`Room`/ocupantes) — confirmado en el código real de este repo, no una suposición.
2. **Validar todo ANTES de mutar cualquier estado** — Polaris muta `layout`/`room` progresivamente mientras valida (`layout.setDoorX(...)`, `layout.parse()`, y recién ahí chequea si `getDoorTile()` quedó `null`, con un fallback defensivo de "descargar la sala y mandar un alert de error genérico" si algo salió mal después de empezar a mutar). Pixels construye y valida el `grid.Grid` completo primero (función pura, `grid.Parse` + validaciones de negocio) y **solo si todo es válido** empieza a tocar persistencia/runtime — sin necesitar ningún camino de "deshacer a mitad de camino".
3. **Errores agregados, siempre en el mismo formato** — Polaris colecciona algunos errores en un `StringJoiner` (puerta/rotación/thickness/wallheight) pero otros simplemente `return` en silencio (tamaño excedido, charset inválido) sin mandar ningún feedback al cliente — inconsistente. Pixels colecciona **todos** los errores de validación de forma uniforme y los reporta juntos (Parte 5.6), nunca en silencio, mismo criterio de "nunca dejar al cliente sin feedback" ya establecido en el resto de esta serie de planes.
4. **`wallHeight` no es un campo suelto** — se agrega al mismo `UpdateParams`/esquema del resto de la configuración de la sala, no un caso especial.

---

## Parte 3 — Esquema

```sql
create table room_custom_layouts (
    room_id bigint primary key references rooms(id) on delete cascade,
    heightmap text not null,
    door_x smallint not null,
    door_y smallint not null,
    door_direction smallint not null,
    wall_thickness smallint not null default 0,
    floor_thickness smallint not null default 0,
    wall_height smallint not null default -1,
    updated_at timestamptz not null default now(),
    constraint room_custom_layouts_door_direction_chk check (door_direction between 0 and 7),
    constraint room_custom_layouts_wall_thickness_chk check (wall_thickness between -2 and 1),
    constraint room_custom_layouts_floor_thickness_chk check (floor_thickness between -2 and 1),
    constraint room_custom_layouts_wall_height_chk check (wall_height between -1 and 15)
);
```
`wall_height` es la única columna que `REMAINING-ROOMS.md` no tenía — confirmado real en Polaris (Parte 1.1), se agrega con su propio rango validado a nivel de base, igual que el resto.

---

## Parte 4 — Resolución de layout: fijo por nombre, o custom por sala

```go
// internal/realm/room/world/layout/custom.go

// CustomReader reads a room's custom floor plan when one exists.
type CustomReader interface {
    FindByRoomID(ctx context.Context, roomID int64) (Layout, bool, error)
}

// ResolveForRoom returns the room's custom layout when saved, otherwise its named
// fixed layout — the two-path resolution REMAINING-ROOMS.md already specified,
// implemented as an explicit extension point rather than a branch inline in callers.
func (service *Service) ResolveForRoom(ctx context.Context, roomID int64, modelName string) (Layout, error) {
    if service.custom != nil {
        if custom, found, err := service.custom.FindByRoomID(ctx, roomID); err != nil {
            return Layout{}, err
        } else if found {
            return custom, nil
        }
    }

    layout, found, err := service.FindByName(ctx, modelName)
    if err != nil {
        return Layout{}, err
    }
    if !found {
        return Layout{}, ErrLayoutNotFound
    }

    return layout, nil
}
```
`access/commands/enter` (la entrada normal a una sala) cambia su llamada de `Layouts.FindByName(ctx, room.ModelName)` a `Layouts.ResolveForRoom(ctx, room.ID, room.ModelName)` — un solo punto de cambio, el resto del flujo de entrada (bootstrap, heightmap, floor items) no se toca.

---

## Parte 5 — Validación del guardado (`FloorPlanEditorSaveHandler`), función pura antes de tocar nada

### 5.1 Tamaño y charset (confirmado, Parte 1.1)

```go
const (
    MaxFloorplanDimension = 64
    MaxFloorplanChars     = MaxFloorplanDimension * MaxFloorplanDimension
)
```
`grid.Parse` ya valida charset/filas parejas — este plan agrega el chequeo de **tamaño máximo de negocio** (64×64) antes de siquiera llamar a `grid.Parse`, ya que el parser compartido no lo impone (se reusa igual, sin límite ahí, porque un layout fijo del hotel podría legítimamente ser más grande).

### 5.2 Puerta, rotación, thickness, altura de pared

Cada uno con su propio código de error de campo — puerta dentro del mapa y no sobre un tile inválido (`grid.Parse(heightmap, grid.WithDoor(x, y))` ya lo valida, se reusa tal cual), rotación 0-7, thickness -2..1 (ambos), wallHeight -1..15. **Chequeo de "altura efectiva 0" configurable** (Parte 1.6):
```go
// internal/realm/room/control/floorplan/config.go
type Config struct {
    RejectZeroEffectiveHeight bool `env:"PIXELS_ROOM_FLOORPLAN_REJECT_ZERO_HEIGHT" envDefault:"true"`
}
```

### 5.3 Tiles bloqueados por furniture existente — chequeo completo, no solo tiles inválidos

```go
// blockedTiles compares the new grid against currently placed furniture, returning
// every tile whose resolved height changed AND currently holds a top item — not just
// tiles becoming fully invalid (Parte 1.3, confirmado contra Polaris real).
func blockedTiles(newGrid grid.Grid, oldGrid grid.Grid, furniture []worldfurniture.Item) []grid.Point {
    occupied := topItemsByTile(furniture) // map[grid.Point]worldfurniture.Item, reusa el índice ya construido por world/items
    var blocked []grid.Point
    for point, item := range occupied {
        newHeight, validNew := newGrid.HeightAt(point)
        oldHeight, _ := oldGrid.HeightAt(point)
        if !validNew || newHeight != oldHeight {
            blocked = append(blocked, point)
            _ = item
        }
    }

    return blocked
}
```
Reusa `worldfurniture.Item`/`world/items.LoadRoomFurniture` ya existentes — no se duplica ninguna consulta de qué hay colocado en la sala.

### 5.4 Auto-pickup (feature adoptada de Polaris, Parte 1.2)

```go
type SaveParams struct {
    // ... Heightmap, DoorX, DoorY, DoorDirection, WallThickness, FloorThickness, WallHeight
    // AutoPickup reports whether blocking furniture should be returned to its
    // owners' inventories instead of aborting the save entirely.
    AutoPickup bool
}

const MaxAutoPickupItems = 500 // mismo tope confirmado en Polaris
```
Si `AutoPickup` y la cantidad de tiles bloqueados con furniture es `<= MaxAutoPickupItems`: cada item bloqueante se recoge (reusa el mismo mecanismo de pickup ya implementado en `FURNITURE-INVENTORY.md`, agrupado por dueño real del item, no por dueño de la sala) y se notifica a cada dueño afectado que esté online (mismo patrón de refresco de inventario ya existente). Si excede el tope, falla explícitamente pidiendo remover manualmente furniture antes de reintentar — mismo criterio que Polaris.

### 5.5 Cooldown de guardado (Redis, mismo mecanismo ya construido en `ENTRY.md`/`CHAT.md`)

```go
// internal/realm/room/control/floorplan/config.go (extendido)
type Config struct {
    RejectZeroEffectiveHeight bool          `env:"PIXELS_ROOM_FLOORPLAN_REJECT_ZERO_HEIGHT" envDefault:"true"`
    SaveCooldown              time.Duration `env:"PIXELS_ROOM_FLOORPLAN_SAVE_COOLDOWN" envDefault:"3s"`
}
```
Clave `floorplan:cooldown:{playerID}`, `pkg/redis.Client.Set` con TTL = `SaveCooldown` tras un guardado exitoso; un intento durante el cooldown rechaza de inmediato sin validar nada más (mismo principio de "chequear lo barato primero" ya aplicado en `ENTRY.md`). 3 segundos por defecto, confirmado igual a Polaris, pero configurable (mismo criterio ya pedido para flood control en `CHAT.md`: nada hardcodeado que un operador pueda necesitar ajustar).

### 5.6 Errores agregados, nunca en silencio (mejora sobre la referencia, Parte 2.3)

```go
type ValidationErrors struct {
    Codes []ErrorCode
}
```
Cada chequeo de 5.1-5.4 que falla agrega su código a la lista en vez de retornar inmediatamente — el handler junta **todos** los fallos de un intento y los manda juntos (Parte 9), nunca deja al cliente sin ninguna respuesta.

---

## Parte 6 — Comandos: poblar el editor + guardar

### 6.1 `floorplan/doorsettings` (poblar posición/rotación de puerta + thickness actual)

Mismo par de packets que Polaris manda juntos en un solo request (`FloorPlanEditorDoorSettingsComposer` + `RoomFloorThicknessUpdatedComposer`, Parte 1): al pedir el editor, el jugador recibe la puerta actual Y el thickness actual en la misma respuesta lógica — dos packets, un comando.

### 6.2 `floorplan/blockedtiles` (poblar tiles bloqueados, Parte 1.4)

Reusa `blockedTiles`-style query (5.3) contra el heightmap **actual** (antes de cualquier cambio) para mostrarle al jugador, de entrada, qué tiles no puede encoger sin auto-pickup.

### 6.3 `floorplan/save`

Mismo wiring que `control/commands/settings/save.go` (Parte 0): `control.Actor` + `control.MatchRoom`, `Authorizer.Authorize` (Parte 8), validar completo (Parte 5) **sin tocar nada todavía**, y solo si pasa todo:
1. Persistir `room_custom_layouts` (upsert).
2. Si `AutoPickup`, ejecutar los pickups (5.4).
3. Recargar el mundo en caliente (Parte 7) — **sin** forward forzado.
4. Confirmar al que guardó + broadcastear a los demás ocupantes.
5. Publicar evento `floorplansaved`.

---

## Parte 7 — Recarga en caliente: la mejora central de este plan

```go
// internal/realm/room/control/commands/floorplan/reload.go

// reloadWorld rebuilds the active room's World from the newly-saved layout and its
// current placed furniture, replacing it in place via Room.LoadWorld — no unload,
// no forced forward, no reconnect. Occupants keep their session; their unit simply
// respawns at the (possibly relocated) door of the new layout, exactly like joining
// fresh — because LoadWorld already re-adds every currently tracked occupant.
func reloadWorld(active *roomlive.Room, newLayout layout.Layout, furniture []worldfurniture.Item) error {
    roomGrid, err := newLayout.Grid()
    if err != nil {
        return err
    }
    door, ok := grid.NewPoint(newLayout.DoorX, newLayout.DoorY)
    if !ok {
        return ErrInvalidWorld // ya validado en Parte 5, esto no debería poder pasar — defensivo, no un camino esperado
    }

    return active.LoadWorld(roomlive.WorldConfig{
        Grid: roomGrid, Furniture: furniture,
        Door: worldpath.Position{Point: door, Z: grid.Height(newLayout.DoorZ)},
        Rules: worldpath.DefaultRules(),
    })
}
```
Tras el reload, cada ocupante recibe los mismos packets de bootstrap que ya se mandan al entrar por primera vez (`SendModel`, `sendFloorItems`, `sendHeightMap` de `access/commands/enter` — reusados tal cual, llamados ahora también desde el flujo de guardado, no solo desde `enter`) — el cliente re-renderiza la sala con el heightmap nuevo, sin necesitar reconectar ni perder su sesión de room.

**Nota honesta**: esto reposiciona a todos en la puerta (mismo comportamiento que un `LoadWorld` de reemplazo ya implica, confirmado leyendo el código real) — no es "conservar la posición exacta de cada uno sobre el mapa nuevo" (que además no tendría sentido si el mapa cambió de forma), es "todos entran de nuevo a la sala ya actualizada, sin la fricción de un forward". Se documenta como el comportamiento esperado, no un bug.

---

## Parte 8 — Nodos de permiso nuevos

```go
// internal/realm/room/permissions.go (agregado)

var (
    // FloorplanOwnEdit allows owners and rights holders to edit a room's floor plan.
    FloorplanOwnEdit = permission.RegisterNode("room.floorplan.own.edit", "")
    // FloorplanAnyEdit allows staff to edit any room's floor plan.
    FloorplanAnyEdit = permission.RegisterNode("room.floorplan.any.edit", "")
)
```
Mapea 1:1 a `ACC_FLOORPLAN_EDITOR` + `ACC_ANYROOMOWNER` confirmados en el código real de Polaris (Parte 1) — unificado al mismo patrón `.own`/`.any` ya usado en el resto de `permissions.go`, en vez de dos flags globales sueltos como en la referencia.

```go
// internal/realm/room/control/floorplan/authorization.go
type Authorizer struct {
    permissions permissionservice.Checker
    rights      RightsChecker
}

func (authorizer *Authorizer) Authorize(ctx context.Context, room roommodel.Room, actorID int64) error {
    if allowed, err := authorizer.has(ctx, actorID, FloorplanAnyEdit); err != nil || allowed {
        return authorizer.result(allowed, err)
    }
    isOwner := room.OwnerPlayerID == actorID
    hasRights := isOwner
    var err error
    if !isOwner {
        hasRights, err = authorizer.rights.HasRights(ctx, room.ID, actorID)
        if err != nil {
            return err
        }
    }
    if !hasRights {
        return ErrAccessDenied
    }

    return authorizer.result(authorizer.has(ctx, actorID, FloorplanOwnEdit))
}
```

---

## Parte 9 — Packets

| Dirección | Paquete | Contenido | Header |
| --- | --- | --- | --- |
| Inbound | `room/floorplan/doorsettings/request` | sin campos | TBD |
| Inbound | `room/floorplan/blockedtiles/request` | sin campos | TBD |
| Inbound | `room/floorplan/save` | `heightmap string`, `doorX int32`, `doorY int32`, `doorDirection int32`, `wallThickness int32`, `floorThickness int32`, `wallHeight int32`, `autoPickup bool` | TBD |
| Outbound | `room/floorplan/doorsettings` | `doorX int32`, `doorY int32`, `doorDirection int32` | TBD |
| Outbound | `room/floorplan/blockedtiles` | `count int32`, N × `(x int32, y int32)` | TBD |
| Outbound | `room/floorplan/error` | lista de códigos de campo (Parte 5.6) | TBD |
| Outbound | `room/floorplan/saved` | confirmación al que guardó | TBD |

Headers `TBD — a confirmar contra Nitro real`. Los shapes en sí (no los headers) están anclados directamente al código real de Polaris leído en la Parte 1, con más confianza que el resto de los packets "a confirmar" de esta serie de planes.

---

## Parte 10 — Hot paths, allocations, benchmarks

- **Guardar un floor plan no es un hot path** (acción rara, deliberada, con su propio cooldown de 3s) — no necesita el mismo nivel de cuidado de allocation que `chat`/movimiento.
- **El reload en caliente (Parte 7) sí debe ser barato en el momento en que ocurre**, porque pasa mientras la sala está activa con ocupantes reales adentro: `worldruntime.New(config)` ya reconstruye fixtures/resolver desde cero (mismo costo que cargar una sala por primera vez, ya optimizado) — no se agrega ningún costo nuevo más allá de eso.
- **`blockedTiles` (5.3) es O(tiles con furniture)**, no O(ancho×alto) — itera solo sobre los tiles que realmente tienen un item encima (`topItemsByTile`, ya indexado por `world/items`), no sobre el mapa completo.

```go
// internal/realm/room/control/floorplan/benchmark_test.go

// BenchmarkValidateSave measures the full validation pipeline (5.1-5.4) against a
// 64x64 map with realistic furniture density, no I/O.
func BenchmarkValidateSave(b *testing.B) { ... }

// BenchmarkReloadWorld measures the cost of LoadWorld replacement at max room size.
func BenchmarkReloadWorld(b *testing.B) { ... }
```

---

## Parte 11 — Tests

- Guardar un heightmap válido persiste `room_custom_layouts`, recarga el mundo **sin** forward (confirmado inspeccionando que ningún packet `forward` se envía, a diferencia de Polaris), y refresca a los ocupantes con los mismos packets de bootstrap que un `enter` normal.
- Caracteres inválidos, filas de longitud distinta, o tamaño excedido (>64×64) fallan sin persistir nada — y el jugador SÍ recibe un error (a diferencia de los `return` silenciosos de Polaris, Parte 2.3).
- Puerta fuera del mapa, sobre un tile inválido, rotación fuera de 0-7, thickness fuera de -2..1, `wallHeight` fuera de -1..15 — cada uno con su propio código, todos reportados juntos si fallan varios a la vez.
- Un tile cuya altura cambia (no solo uno que se vuelve `x`) y tiene furniture encima bloquea sin `AutoPickup` — regresión directa de la Parte 1.3.
- Con `AutoPickup: true`, la furniture bloqueante se recoge y vuelve al inventario de su dueño real (no necesariamente el dueño de la sala), con notificación si está online; excede `MaxAutoPickupItems` falla explícitamente.
- Intentar guardar dentro de la ventana de cooldown (3s por defecto) rechaza sin validar nada más.
- Solo el dueño (+ `FloorplanOwnEdit`), un rights-holder (+ `FloorplanOwnEdit`), o staff (`FloorplanAnyEdit`) pueden guardar — cualquier otra combinación falla.
- `ResolveForRoom` retorna el layout custom cuando existe, y cae al fijo por nombre cuando no — regresión de la Parte 4.

---

## Parte 12 — Milestones de implementación

1. **FP1 — Esquema + `wall_height`** (Parte 3): migración de `room_custom_layouts`.
2. **FP2 — `world/layout.ResolveForRoom`** (Parte 4): extiende la resolución de layout; cambia `access/commands/enter` para usarla — depende de FP1.
3. **FP3 — Validación completa** (Parte 5): tamaño/charset/puerta/rotación/thickness/wallHeight/tiles-bloqueados/auto-pickup/cooldown/errores agregados — depende de FP1.
4. **FP4 — Nodos de permiso + `Authorizer`** (Parte 8): puede correr en paralelo a FP3.
5. **FP5 — Comandos `doorsettings`/`blockedtiles`/`save`** (Parte 6, 9): depende de FP2, FP3, FP4.
6. **FP6 — Recarga en caliente** (Parte 7): depende de FP5 — el punto donde este plan mejora deliberadamente sobre Arcturus y Polaris.
7. **FP7 — Benchmarks** (Parte 10): depende de FP3/FP6.

### Milestones futuros confirmados (fuera de este documento, no descartados)

- **Auto-pickup con notificación push si el dueño real está en otra sala** — hoy se asume reuso directo del mecanismo de refresco de inventario ya existente (`FURNITURE-INVENTORY.md`); si ese mecanismo no cubre "el dueño está online pero en otra sala", se ajusta ahí, no en este plan.
- **Límite de `MaxUsers` atado a la capacidad real del nuevo layout** — ya anotado como pendiente en `plan/rooms/CONFIG.md` Parte 11 ("Milestones futuros confirmados"); un floor plan más chico podría necesitar bajar el `MaxUsers` configurado, pero no hay evidencia todavía de que el editor real fuerce ese ajuste — se revisita si aparece un caso real.
