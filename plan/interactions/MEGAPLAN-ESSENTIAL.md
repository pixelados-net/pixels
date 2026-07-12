# Megaplan: Interacciones esenciales pendientes (`internal/realm/furniture/interactions/*`)

Cubre, totalmente detallados, los 13 `interaction_type` de `plan/INTERACTIONS.md` Parte 10.1 que faltan y **no** están diferidos: `dice`, `colorwheel`, `random_state`, `pressureplate`, `colorplate`, `onewaygate`, `switch`, `switch_remote_control`, `multiheight`, `vendingmachine` (+ `vendingmachine_no_sides`), `handitem`, `handitem_tile`, y `cannon`. Cada uno se cruzó contra su clase real en **Polaris-Emulator** (`github.com/duckietm/Polaris-Emulator`, Java) — no contra el resumen ya existente en la tabla de `INTERACTIONS.md`, sino leyendo el código fuente de nuevo para los siete que anclan cada familia de comportamiento (`InteractionDice`, `InteractionPressurePlate`, `InteractionOneWayGate`, `InteractionMultiHeight`, `InteractionVendingMachine`, `InteractionCannon`, y las runnables `CannonKickAction`/`RoomUnitGiveHanditem`).

**Explícitamente fuera de este megaplan** (a pedido, aunque algunos no figuran como "Diferido" en la tabla de `INTERACTIONS.md`): `roller` (tiene su propio Milestone I5 ya planeado en `INTERACTIONS.md` Parte 7), `pyramid`, `stack_helper`, `pressureplate_group` (diferido a grupos), `obstacle`/`snowboard_slope`/`water`/`water_item` (diferidos a efectos), y todo lo demás de las tablas 10.2-10.4.

Estado: implementado en Pixels. Las dependencias explícitamente diferidas al
final de este documento permanecen fuera de alcance.

---

## Parte 0 — Punto de partida real (grounding)

`plan/interactions/TELEPORT.md` y `plan/interactions/TOGGLE-GATE.md` ya se implementaron, y este megaplan reusa directo lo que construyeron:

| Ya existe | Dónde | Cómo lo usa este plan |
| --- | --- | --- |
| `world/furniture.Item.ExtraData` + `Room.SetFurnitureExtraData` | `internal/realm/room/world/furniture/item.go`, `internal/realm/room/runtime/live/world.go` | Canal de estado visual sin reconstruir fixtures — la base de `dice`/`colorwheel`/`random_state`/`vendingmachine`/`cannon` (ninguno cambia walkability real). |
| `worldunit.ControlKind` | `internal/realm/room/world/unit/unit.go` | Generaliza "movimiento controlado por servidor" — reusado por `onewaygate`/`vendingmachine`/`switch` para el workflow de "caminar al activador" (Parte 2.4), en vez de banderas sueltas nuevas. |
| `furnitureaccess.CanManage` | `internal/realm/furniture/access/manage.go` | Autorización de rights ya resuelta — reusada por `multiheight`/`switch_remote_control` (los dos que exigen rights) sin nodo nuevo. |
| `worldfurniture.Footprint` | `internal/realm/room/world/furniture/footprint.go` | Reusado por `pressureplate`/`colorplate`/`multiheight` para el chequeo de ocupación de todo el footprint (mismo patrón ya usado por Gate). |
| `room.Unkickable` (nodo de permiso) | `internal/realm/room/permissions.go`, `plan/rooms/RIGHTS.md` | El cañón (Parte 8) lo consulta directo para excluir jugadores inmunes de la línea de disparo — mapeo 1:1 confirmado contra `ACC_UNKICKABLE` real. |
| `leavecmd.Handler` (expulsión de room) | `internal/realm/room/access/commands/leave` | Reusado por el cañón para expulsar sin pasar por `moderation.Service.Kick` (no hay un "actor moderando", es un efecto ambiental de furniture — Parte 8.1). |
| `Behavior`/comando compartido `furniture.interact` | `plan/interactions/TOGGLE-GATE.md` Parte 2 | `switch`/`switch_remote_control` terminan delegando al mismo toggle genérico tras resolver su propio workflow de adyacencia. |
| `room/furniture/state` (packet de estado) | `plan/interactions/TOGGLE-GATE.md` Parte 5 | Reusado tal cual por todos los tipos de este plan — ninguno necesita un packet de estado nuevo. |

---

## Parte 1 — Primitivas compartidas (se construyen antes que cualquier tipo)

`INTERACTIONS.md` Parte 10.5 ya las anotó como prerequisito; este plan las implementa con el detalle real confirmado en la Parte 2-8.

### 1.1 Scheduler de tareas retrasadas por room — generaliza el `PhaseDeadline` de teleport

Teleport avanzaba su máquina de estados enganchado al tick de 500ms porque **todos** sus delays eran múltiplos exactos de 500ms. Este megaplan tiene delays que no lo son: `100ms` (pressureplate, debounce de reevaluación), `750ms`/`2000ms` (cannon), `1500ms` (dice/vending), `3000ms` (colorwheel), `500ms` (el resto). Generalizar a una cola de tareas por room, no una bandera de fase:

```go
// internal/realm/room/runtime/live/schedule.go

// ScheduledTask runs once its deadline passes, checked on every tick — same
// principle as SweepDoorbell/teleport's PhaseDeadline, generalized to an
// arbitrary deadline and an arbitrary callback instead of a fixed enum of phases.
type ScheduledTask struct {
    Deadline time.Time
    Run      func(now time.Time)
}

// scheduled stores pending tasks, lazily allocated exactly like doorbell/teleports.
func (room *Room) Schedule(after time.Duration, run func(time.Time)) {
    room.mutex.Lock()
    defer room.mutex.Unlock()
    room.scheduled = append(room.scheduled, ScheduledTask{Deadline: room.now().Add(after), Run: run})
}
```
`Room.Tick()` gana un paso más: recorre `scheduled`, ejecuta y remueve las que ya vencieron. **Trade-off documentado, no un descuido**: al estar atado al tick de 500ms, una tarea de `100ms`/`750ms` puede disparar hasta ~500ms tarde en el peor caso — imperceptible para efectos visuales de furniture (un cañón que dispara a los 750-1200ms en vez de exactamente 750ms no se nota), y evita el costo real de un timer/goroutine por tarea. Si el profiling real muestra que hace falta más precisión, se ajusta el `tickInterval` global o se agrega un segundo tick más fino — no se construye esa complejidad sin evidencia.

### 1.2 Estado efímero vs. durable, formalizado

Regla única para los 13 tipos: **todo estado intermedio** (`"-1"` de dice/colorwheel mientras se resuelve el azar, `"1"` de vendingmachine/cannon mientras dura la animación) usa `SetFurnitureExtraData` (memoria + broadcast, sin escritura síncrona a Postgres). **Solo el estado final en reposo** se persiste, de forma asíncrona best-effort (mismo criterio ya fijado en `TOGGLE-GATE.md` Parte 3.3) — ninguno de estos tipos debería generar más de una escritura real a la base por uso, sin importar cuántos frames intermedios tenga la animación.

### 1.3 RNG inyectable

```go
// internal/realm/furniture/interactions/random.go

// Source provides deterministic-in-tests randomness for dice/colorwheel/random_state/vending.
type Source interface {
    IntN(n int) int
}
```
Una sola interfaz, una sola implementación real (`crypto/rand`-backed o `math/rand/v2`, a decidir en implementación), inyectada donde haga falta — nunca un generador nuevo por tipo de interacción.

### 1.4 Workflow compartido: "caminar al activador válido más cercano, luego reintentar"

Confirmado como el mismo patrón en **tres** clases reales leídas (`InteractionOneWayGate`, `InteractionVendingMachine`, y descrito igual para `switch` en la tabla de `INTERACTIONS.md`): si la unidad no está ya parada en un tile de activación válido, se calcula el más cercano de los válidos, se camina ahí, y **al llegar** se reintenta la interacción original (nunca un segundo click del jugador).

```go
// internal/realm/furniture/interactions/activator.go

// WalkToNearestActivator walks a unit to the closest tile in activatorTiles when
// it isn't already standing on one, then re-invokes retry — shared by onewaygate,
// vendingmachine, and switch instead of three copies of the same walk+callback.
func WalkToNearestActivator(active *roomlive.Room, playerID int64, activatorTiles []grid.Point, retry func()) error { ... }
```

### 1.5 `HandItemHolder` — más simple de lo que el nombre sugiere

Confirmado leyendo `RoomUnitGiveHanditem.java`: es literalmente un entero en la unidad + un packet dedicado, nada más.

```go
// internal/realm/room/world/unit/unit.go (extendido)
// handItem stores the currently held hand item id, 0 when empty.
handItem int32

func (unit *Unit) SetHandItem(itemID int32) { unit.handItem = itemID }
func (unit *Unit) HandItem() int32          { return unit.handItem }
```
```
outbound room/entities/handitem — (unitId int32, itemId int32), equivalente a RoomUserHandItemComposer
```
Sin expiración por sí solo (el hand item se reemplaza al recibir otro, o al soltarlo — comando `handitem/drop`/click en el propio avatar, fuera de alcance de este plan si Nitro lo maneja client-side; se anota como dependencia a confirmar).

---

## Parte 2 — Grupo A: revelado aleatorio con delay (`dice`, `colorwheel`, `random_state`)

### 2.1 Dice — confirmado, Java real

```java
// InteractionDice.onClick (Polaris real)
if (RoomLayout.tilesAdjecent(diceTile, unit.getCurrentLocation())) {
    if (!this.getExtradata().equalsIgnoreCase("-1")) {
        this.setExtradata("-1");                 // "rodando"
        room.updateItemState(this);
        Emulator.getThreading().run(this);        // el propio item se re-ejecuta como Runnable (ver 2.1.1)
        Emulator.getThreading().run(new RandomDiceNumber(...), 1500); // resultado a los 1500ms
    }
}
```
- **Requiere adyacencia** (no encima) — a diferencia de teleport/vendingmachine, dice **no** camina automáticamente al jugador hasta el tile adyacente; si no está adyacente, el click simplemente no hace nada (confirmado: no hay ningún camino de "caminar hacia" en esta clase).
- Mientras `extradata == "-1"` ("rodando"), un segundo click se ignora — mismo guard de doble-disparo que teleport.
- A los 1500ms: estado final aleatorio `1..StateCount` (vía `Source.IntN`, Parte 1.3).
- `isWalkable() = false` siempre — un dado nunca es caminable, sin importar su definición base.
- `onPickUp` resetea a `"0"` — nunca se persiste "rodando" ni un resultado viejo al recogerlo.
- `allowWiredResetState() = false` — igual que teleport, su estado lo controla únicamente su propia secuencia.

### 2.1.1 El "re-run" del propio item — patrón de auto-programación

`Emulator.getThreading().run(this)` — el propio `HabboItem` implementa `Runnable` y se reprograma a sí mismo (visto también en `CustomRoomLayout`, `plan/rooms/FLOORPLAN.md` Parte 1.1, y en `InteractionVendingMachine.run()`, Parte 5). Pixels no necesita este patrón de auto-referencia: el `ScheduledTask.Run` de la Parte 1.1 ya recibe un closure — el "qué hacer al vencer" se captura ahí directamente, sin que el item necesite conocerse a sí mismo como tarea programable.

### 2.2 Colorwheel — mismo primitivo, tres diferencias reales

Confirmado por la tabla ya investigada de `INTERACTIONS.md` (no releído de nuevo, dado que es una variación directa de `dice` sin lógica adicional que justifique un segundo fetch):
1. **Requiere rights de room** (`furnitureaccess.CanManage`), a diferencia de dice que no chequea nada más que adyacencia.
2. **Delay de 3000ms**, no 1500ms.
3. **Es un wall item** — el estado se transmite vía el packet de wall-update (`plan/INTERACTIONS.md` Parte 5, Milestone I3, todavía sin implementar) en vez del de piso — **dependencia real**: colorwheel no puede completarse hasta que I3 (wall items) exista. Se anota en milestones (Parte 13).

### 2.3 `random_state` — la generalización pura del primitivo

Sin objeto físico especial: "limpia el estado y, tras el delay definido en `customparams`, publica un estado aleatorio entre los permitidos" (confirmado en `INTERACTIONS.md`). Se modela como el mismo primitivo de Dice sin el guard de adyacencia (no hay evidencia de que lo requiera) y con el delay leído de un campo de configuración de la definición en vez de una constante:

```go
// internal/realm/furniture/interactions/randomstate/behavior.go
type Behavior struct {
    Delay time.Duration // parseado una vez desde Definition.CustomParams al cargar la definición, no en cada click
}
```
**Único tipo de este grupo que toca `customparams`** — Pixels no construye un sistema tipado genérico de custom values para esto (`INTERACTIONS.md` Parte 10.5 ya advertía no copiar el `key=value;` mutable de Arcturus sin necesidad): se parsea el único campo que `random_state` necesita (un entero de milisegundos) al cargar la definición, sin generalizar de más.

### 2.4 Packets/eventos compartidos del grupo

```
outbound room/furniture/state — reusado, valor "-1" mientras "rueda", valor final al asentarse
furniture.random_resolved {PlayerID, ItemID, RoomID, Result} — evento nuevo, común a los tres
```

### 2.5 Tests

- Dice: click adyacente → extradata `"-1"`, segundo click mientras rueda es no-op, a los 1500ms (reloj inyectado) el estado final está en `1..StateCount`; click NO adyacente no hace nada (ni camina, ni rueda).
- Colorwheel: mismo flujo con 3000ms y autorización de rights (rechaza sin `CanManage`).
- `random_state`: el delay leído de `CustomParams` se respeta exactamente (parametrizado, no hardcodeado); el resultado está siempre dentro del rango de estados permitido.
- Los tres: `Source` inyectado con una secuencia fija hace el resultado determinístico en el test, sin flakiness.

---

## Parte 3 — Grupo B: ocupación de tile (`pressureplate`, `colorplate`)

### 3.1 Pressure plate — confirmado, Java real

```java
// InteractionPressurePlate (Polaris real) — extiende InteractionDefault
public void onWalkOn(...)  { super.onWalkOn(...); Emulator.getThreading().run(() -> updateState(room), 100); }
public void onWalkOff(...) { super.onWalkOff(...); Emulator.getThreading().run(() -> updateState(room), 100); }
public void onMove(...)    { super.onMove(...); updateState(room); } // sin debounce acá, inmediato

public void updateState(Room room) {
    boolean occupied = false;
    for (RoomTile tile : /* todo el footprint */) {
        var unitsHere = room.getHabbosAndBotsAt(tile.x, tile.y);
        if (unitsHere.isEmpty() && this.requiresAllTilesOccupied()) { occupied = false; break; }
        if (!unitsHere.isEmpty()) occupied = true;
    }
    this.setExtradata(occupied ? "1" : "0");
    room.updateItemState(this);
}
```
- **Sin click** (`onClick` vacío) — el estado se deriva **exclusivamente** de ocupación, nunca de una interacción directa.
- **Debounce de 100ms** en walk-on/walk-off (no en move) — evita recalcular una vez por cada unidad si varias entran/salen casi simultáneamente; se resuelve con `Room.Schedule` (Parte 1.1), sobrescribiendo cualquier recálculo ya pendiente para el mismo item en vez de apilar N tareas redundantes (`Room.Schedule` gana una variante `ScheduleReplacing(key, after, run)` para esto, identificado por `itemID`).
- **`requiresAllTilesOccupied()`** — hook para una variante (`pressureplate_group`, diferida) que exige TODOS los tiles del footprint ocupados, no solo alguno; la variante base (este plan) siempre retorna `false` (alguno alcanza).
- **`isWalkable() = true` siempre** — a diferencia de Gate, una pressure plate nunca bloquea el paso, solo reporta si está ocupada.
- **`allowWiredResetState() = true`** — a diferencia de dice, se puede resetear externamente (sin efecto práctico hasta que exista wired).

### 3.2 Colorplate — mismo hook, semántica de contador acotado

`INTERACTIONS.md`: "incrementa el estado por cada unidad que entra y lo decrementa al salir, acotado al state count". Mismos hooks (`onWalkOn`/`onWalkOff`/`onMove`), pero en vez de recomputar un booleano por ocupación total, mantiene un contador:

```go
func (colorplateBehavior) OnWalkOn(active *roomlive.Room, item worldfurniture.Item, definition furnituremodel.Definition) {
    next := clamp(currentState(item)+1, 0, definition.InteractionModesCount-1)
    active.SetFurnitureExtraData(item.ID, strconv.Itoa(next))
}
func (colorplateBehavior) OnWalkOff(active *roomlive.Room, item worldfurniture.Item, definition furnituremodel.Definition) {
    next := clamp(currentState(item)-1, 0, definition.InteractionModesCount-1)
    active.SetFurnitureExtraData(item.ID, strconv.Itoa(next))
}
```
**Sin debounce** — a diferencia de pressureplate, un contador incremental no tiene el mismo problema de "recalcular de más" (cada entrada/salida es un delta independiente, no una recomputación completa de ocupación), así que corre inmediato, sin pasar por `Room.Schedule`.

### 3.3 Tests

- Pressureplate: una unidad entra → tras 100ms (reloj inyectado) pasa a `"1"`; se va → tras 100ms vuelve a `"0"`; dos unidades entran casi simultáneo → una sola recomputación, no dos (regresión del debounce).
- Colorplate: tres unidades entran sucesivamente → el estado sube de a uno, acotado a `InteractionModesCount-1`; salir baja de a uno, nunca por debajo de `0`.
- Ambos: `move`-ando la furniture con unidades ya sobre su footprint viejo recomputa contra el nuevo footprint (reusa el "walked off sintético" ya construido en `TOGGLE-GATE.md` Parte 3.4).

---

## Parte 4 — Grupo C: adyacencia + toggle (`onewaygate`, `switch`, `switch_remote_control`)

### 4.1 One-way gate — el más elaborado de los tres, confirmado Java real

```java
// InteractionOneWayGate.onClick (Polaris real), resumen del flujo:
// 1. Solo dispara si la unidad está parada EXACTAMENTE en el tile-en-frente (no en cualquier lado).
// 2. Si el tile del gate no tiene ya alguien parado:
//    - marca el tile caminable (override), lo agrega como override-tile de la unidad
//    - la hace caminar hacia el tile del gate
//    - onSuccess: deshabilita salir por la puerta, y hace caminar a la unidad HACIA EL OTRO LADO
//      (tileInFront con rotación+4, es decir el frente opuesto) — el "paso a través"
//    - 500ms después del onSuccess: dispara un trigger opcional de wired (WiredManager.triggerUserWalksOn)
//    - onFail: restaura todo, vuelve extradata a "0"
// 3. Nunca queda "abierto" de forma persistente — onWalkOff/onPickUp/onMove/onPlace siempre llaman refresh() → "0"
```
Confirma que un one-way gate es, conceptualmente, una versión de un solo uso de la Fase A-D de teleport (caminar hacia el activador → tránsito controlado → caminar hacia afuera del otro lado) pero **sin cambiar de room ni de tile real** — la unidad simplemente atraviesa el mueble en línea recta. Pixels reusa el mismo `WalkToNearestActivator`/`ControlKind` de las Partes 1.4/0, no un mecanismo paralelo:

```go
// internal/realm/furniture/interactions/onewaygate/behavior.go

func (behavior) Trigger(active *roomlive.Room, playerID int64, item worldfurniture.Item) error {
    // 1. Debe estar en el tile-en-frente exacto — a diferencia de vendingmachine, sin
    //    fallback de "caminar desde cualquier lado", solo desde el frente.
    // 2. Debe estar libre el tile propio del gate (nadie más cruzando ahora mismo).
    // 3. Marca Control = ControlOneWayGateCrossing (nuevo valor de ControlKind, Parte 0).
    // 4. Camina al tile del gate (extradata "1" mientras tanto) → al llegar, camina al
    //    tile-en-frente OPUESTO (rotación+4) → libera Control.
    // 5. Falla en cualquier paso → refresh(): extradata "0", libera Control, sin dejar
    //    a la unidad a mitad de camino.
}
```
**Siempre vuelve a `"0"`** — a diferencia de Gate (Milestone I2, estado persistente hasta el próximo toggle), un one-way gate es momentáneo por diseño, nunca queda "abierto" esperando.

### 4.2 Switch — el más simple: adyacencia + toggle normal

`INTERACTIONS.md`: "si el keko no está adyacente, camina al tile válido más cercano y luego ejecuta el toggle". Reusa `WalkToNearestActivator` (Parte 1.4) y, al llegar, delega directo al `Behavior` de Toggle **ya implementado** (`TOGGLE-GATE.md` Parte 3) — `switch` no tiene ninguna lógica de estado propia, es 100% el workflow de acercamiento aplicado sobre un toggle común.

### 4.3 Switch remote control — sin adyacencia, y una advertencia real

`INTERACTIONS.md` ya señalaba: "la clase de referencia termina ciclando dos veces por su herencia, comportamiento sospechoso que no debe copiarse literalmente". Pixels implementa la versión **correcta**: toggle inmediato sin ningún requisito de posición, delegando directo al mismo `Behavior` de Toggle (una sola vez, no dos) — la única diferencia real con `switch` es que **no** exige adyacencia ni dispara ningún workflow de caminata.

### 4.4 Tests

- One-way gate: cruzarlo desde el frente correcto atraviesa y vuelve a `"0"` solo; intentarlo desde el lado equivocado no hace nada; dos jugadores casi simultáneos — el segundo no puede iniciar mientras el primero está cruzando (tile propio ocupado).
- Switch: click lejos del activador camina antes de togglear; click ya adyacente togglea directo.
- Switch remote control: un solo click produce **un solo** cambio de estado (regresión explícita contra el bug de doble-ciclo confirmado en la referencia — este es el test que garantiza que Pixels no lo heredó).

---

## Parte 5 — Multiheight

### 5.1 Diseño real confirmado

```java
// InteractionMultiHeight.onClick (Polaris real)
if (!room.hasRights(habbo) && !ejecutadoPorWired) return;
HabboItem topItem = room.getTopItemAt(x, y);
if (topItem != null && !topItem.equals(this)) return; // NO se puede cambiar de altura con algo apilado encima
extradata = (Integer.parseInt(extradata) + 1) % multiHeights.length;
room.updateItem(this);
updateUnitsOnItem(room); // recalcula Z de cada unidad parada/sentada encima
```
```java
// updateUnitsOnItem: por cada unidad sobre el footprint —
// si está en movimiento hacia OTRO tile, se ignora (no se le pisa el goal)
// si el mueble permite sentarse (o la unidad ya está sentada): sitUpdate = true (recalcula la animación de sentado a la nueva altura)
// si no: fuerza Z = altura actual del tile, sin animación de movimiento — un "snap" instantáneo
```
- **Requiere rights** (o wired) — mismo `furnitureaccess.CanManage` que Toggle/Gate.
- **Bloqueado si hay OTRA furniture apilada encima** (no unidades — unidades sí se ajustan, otra furniture NO permite el cambio en absoluto) — chequeo de `TopItemAt` distinto del chequeo de ocupación de Gate/pressureplate.
- **Recalcula la posición Z de cada unidad sobre el footprint** al cambiar de altura — esto es lo que justifica la nota de `INTERACTIONS.md` 10.5 ("Surface resolver versionado + actualización de unidades"): el cambio de altura debe reflejarse en `surface.Resolver` (reconstrucción de fixture, igual que Gate) **y además** recorrer las unidades ya paradas/sentadas ahí para reposicionarlas, sin esperar a que caminen para "caer" a la nueva altura.

### 5.2 Diseño Pixels

```go
// internal/realm/furniture/interactions/multiheight/behavior.go

func (behavior) Apply(active *roomlive.Room, item worldfurniture.Item, definition furnituremodel.Definition) (next string, rebuildFixtures bool, err error) {
    if active.HasOtherFurnitureOn(item.Point, item.ID) { // nuevo helper: TopItemAt distinto del propio
        return item.ExtraData, false, ErrBlockedByStackedFurniture
    }
    // cicla sobre definition.MultiHeights (nuevo campo, paralelo a InteractionModesCount
    // pero con una altura real por índice, no solo un contador cosmético)
    return next, true, nil // true: SIEMPRE reconstruye fixtures, igual que Gate
}

// tras Apply exitoso, el comando llama:
func (active *Room) ResettleUnitsOn(point grid.Point, footprint []grid.Point) []UnitSnapshot {
    // mismo criterio confirmado: unidades en movimiento hacia otro destino se ignoran;
    // sentadas conservan el status "sit" reproyectado a la nueva altura; el resto snap-ea Z
}
```
Reusa el mismo campo `AllowSit`/status `sit` ya existente (`worldunit.Status`, `TOGGLE-GATE.md` Parte 0) — no se inventa un concepto nuevo de "sentado", se reproyecta el que ya existe a la altura nueva.

### 5.3 Tests

- Ciclar altura sin nada encima → nuevo `ExtraData`, fixture reconstruido, `surface.Resolver` refleja la nueva altura.
- Otra furniture apilada encima → rechaza sin cambiar nada.
- Una unidad parada (no sentada) sobre el footprint → su Z salta a la nueva altura sin animación de movimiento.
- Una unidad sentada sobre el footprint → conserva el status `sit`, reproyectado a la nueva altura (no se pone de pie).
- Una unidad caminando HACIA otro tile (no el multiheight) que pasa a estar temporalmente sobre el footprint → no se le pisa el `goal` en curso.

---

## Parte 6 — Grupo E: hand items (`vendingmachine`, `vendingmachine_no_sides`, `handitem`, `handitem_tile`)

### 6.1 Vending machine — confirmado, Java real

```java
// InteractionVendingMachine (Polaris real)
Set<RoomTile> activatorTiles = { tile propio, tile-en-frente }; // exactamente estos dos, no más

onClick:
  si la unidad NO está en ninguno de los dos → camina al más cercano de los válidos, al llegar reintenta
  si ya está en uno → useVendingMachine():
    extradata = "1"                                  // "en uso"
    si no está caminando/sentada/acostada → rota a mirar hacia la máquina
    tras 1500ms:
      entrega un hand item ALEATORIO de la lista vending_ids de la definición (Source.IntN, Parte 1.3)
      si la definición tiene efecto por género, lo otorga (diferido — sistema de efectos no existe, Parte 6.1.1)
      tras 500ms más: extradata vuelve a "0"
```
- **Activator tiles = exactamente dos** (propio + frente) — a diferencia de `switch`/`onewaygate` que solo aceptan un tile válido, y de `vendingmachine_no_sides` (6.2) que acepta más.
- **Rotación automática hacia la máquina** si la unidad no está ya en medio de otra animación (caminando/sentada/acostada) — detalle real que ninguna otra interacción de este plan necesita.
- **El efecto de género al entregar el item** — mismo hallazgo que `TOGGLE-GATE.md` Parte 1.4 (efectos de avatar): confirmado real, **diferido** por el mismo motivo (sin sistema de efectos en Pixels).

### 6.2 Vending machine (no sides) — variante de política de activadores

`INTERACTIONS.md`: "acepta cualquiera de los tiles alrededor del footprint". Mismo comando, misma máquina de estados (6.1) — la única diferencia es la función que calcula `activatorTiles` (todos los tiles perimetrales del footprint en vez de solo el frente). Se modela como un parámetro de `Behavior`, no una implementación paralela.

### 6.3 Hand item — el caso simple, sin azar

`INTERACTIONS.md`: "adyacente o encima: entrega un hand item; si tiene varios estados anima `0→1→0` durante 500ms". Mismo primitivo que vending (extradata "1" transitorio, `SetHandItem` al final del delay), pero:
- **El item entregado es fijo** (definido en la propia definición, no elegido al azar de una lista) — no necesita `Source` (Parte 1.3).
- **Delay fijo de 500ms**, no 1500/1000 como vending.
- **Acepta encima O adyacente** — más laxo que vending (que exige EXACTAMENTE los dos tiles de su propia definición geométrica).

### 6.4 Hand item tile — disparo por caminar, no por click

Mismo `handitem` (6.3), pero el disparador es `furniture.walkedon` (ya real desde `TOGGLE-GATE.md` Parte 7) en vez de un click — sin ningún estado ni lógica adicional distinta.

### 6.5 Tests

- Vending: click lejos de los dos tiles válidos camina antes de activar; click ya en uno de los dos activa directo; tras 1500ms (reloj inyectado) el jugador recibe un hand item de la lista configurada (determinístico con `Source` fijo en el test); 500ms después el item vuelve a `"0"`.
- Vending no-sides: cualquier tile perimetral del footprint activa, no solo el frente — test que confirma la diferencia real de política.
- Handitem: funciona parado encima Y desde cualquier tile adyacente (a diferencia de vending); el item entregado es siempre el mismo configurado, nunca aleatorio.
- Handitem tile: caminar encima (sin click) entrega el item.
- Los cuatro: `SetHandItem`/`HandItem()` reflejan el id correcto tras cada entrega; el packet `room/entities/handitem` se broadcastea a todo el room, no solo a quien lo recibió (para que los demás vean el prop en su mano).

---

## Parte 7 — Cannon

### 7.1 Diseño real confirmado

```java
// InteractionCannon (Polaris real)
onClick:
  calcula la "mecha" (fuseTile) según rotación, y los tiles válidos alrededor de ella
  si el jugador NO está en un tile válido, o ya hay cooldown activo → no hace nada (SIN caminar automático, a diferencia de vending/onewaygate)
  si es válido:
    congela al jugador (canWalk = false), lo orienta mirando hacia la mecha
    cooldown = true
    extradata alterna 0/1
    tras 750ms: CannonKickAction
    tras 2000ms: se libera el cooldown
```
```java
// CannonKickAction (Polaris real)
descongela al que disparó
calcula una línea de 3 tiles delante del cañón (dirección derivada de la rotación)
para cada Habbo en esos 3 tiles:
    si NO tiene ACC_UNKICKABLE Y NO es el dueño de la sala → lo expulsa del room (leaveRoom) + le manda un bubble alert "cannon.png"
```
- **Sin workflow de caminata** — a diferencia de vending/onewaygate/switch, el cañón NO camina al jugador hacia un tile válido; si no estás ya parado en uno de los tiles de activación, el click simplemente no hace nada.
- **Cooldown real de 2000ms**, independiente del delay de 750ms del disparo — dos temporizadores (Parte 1.1) por uso, no uno.
- **La exclusión de inmunes mapea 1:1 a lo que Pixels ya tiene**: `!ACC_UNKICKABLE && !isOwner` es exactamente `!HasPermission(target, room.Unkickable) && room.OwnerPlayerID != target` — cero conceptos nuevos de permisos.
- **La expulsión es un kick ambiental, no una acción de moderación discretionary** — se implementa llamando directo al mismo mecanismo que usa `leave`/`moderation.Kick` internamente (`leavecmd.Handler`), **sin** pasar por la fórmula de autorización de `moderation.Service.Kick` (que espera un actor humano decidiendo caso por caso) — quien "decide" acá es el dueño de la sala, en el momento de colocar/activar el cañón, no en cada disparo puntual.

### 7.2 Diseño Pixels

```go
// internal/realm/furniture/interactions/cannon/behavior.go

func (behavior) Trigger(active *roomlive.Room, playerID int64, item worldfurniture.Item) error {
    if active.CannonOnCooldown(item.ID) {
        return nil // no-op silencioso, mismo criterio que la referencia
    }
    fuseTiles := activationTiles(item) // calculado desde rotación, igual que fuseTile+tilesAround
    if !containsUnit(fuseTiles, playerID) {
        return nil // sin caminata automática — a propósito, distinto del resto de este plan
    }
    active.FreezeUnit(playerID)          // Control = ControlFrozen, nuevo valor de ControlKind
    active.FaceToward(playerID, item.Point)
    active.SetCannonCooldown(item.ID, true)
    active.SetFurnitureExtraData(item.ID, toggle(item.ExtraData))
    active.Schedule(750*time.Millisecond, func(time.Time) { kickLine(active, item) })
    active.Schedule(2000*time.Millisecond, func(time.Time) { active.SetCannonCooldown(item.ID, false) })

    return nil
}

func kickLine(active *roomlive.Room, item worldfurniture.Item) {
    active.UnfreezeUnit(/* quien disparó */)
    for _, point := range lineOfThree(item) {
        for _, occupant := range active.OccupantsAt(point) {
            if occupant.PlayerID == active.Snapshot().OwnerPlayerID {
                continue
            }
            if immune, _ := permissions.HasPermission(ctx, occupant.PlayerID, room.Unkickable); immune {
                continue
            }
            _ = leavecmd.Handler{...}.Handle(ctx, command.Envelope[leavecmd.Command]{Command: leavecmd.Command{PlayerID: occupant.PlayerID}})
            sendAlert(occupant, "cannon.kicked") // reusa GENERIC_ALERT, ya existente
        }
    }
}
```

### 7.3 Tests

- Disparar desde un tile válido con el cooldown libre → congela, orienta, alterna extradata, y a los 750ms (reloj inyectado) expulsa a quien esté en la línea de 3 tiles.
- El dueño de la sala nunca es expulsado, incluso parado en la línea de tiro.
- Un jugador con `room.Unkickable` nunca es expulsado.
- Disparar durante el cooldown (antes de los 2000ms) es un no-op — sin congelar, sin volver a alternar el estado.
- Disparar desde un tile fuera de la zona de activación no hace nada — y, a diferencia de vending, no dispara ninguna caminata automática hacia la mecha.

---

## Parte 8 — Packets nuevos (resumen completo del megaplan)

| Dirección | Paquete | Contenido | Reusado por |
| --- | --- | --- | --- |
| Outbound | `room/furniture/state` (ya existente, `TOGGLE-GATE.md`) | `itemId int32`, `state string` | Todos los 13 tipos |
| Outbound | `room/entities/handitem` (nuevo, Parte 1.5) | `unitId int32`, `itemId int32` | vending, vending-no-sides, handitem, handitem_tile |
| — | `GENERIC_ALERT` (ya existente) | mensaje localizado | cannon (Parte 7.2) |
| Inbound | `furniture/interact` (ya existente, `TOGGLE-GATE.md`) | `itemId int32` | dice, colorwheel, onewaygate, switch, switch_remote_control, multiheight, vending, handitem, cannon |

**Ningún packet inbound nuevo** — los 13 tipos de este megaplan reusan el mismo `furniture/interact` ya diseñado para Toggle/Gate; lo único nuevo del lado del protocolo es el packet de hand item (Parte 1.5).

---

## Parte 9 — Nodos de permiso

**Ninguno nuevo**, igual que `TOGGLE-GATE.md`. `multiheight`, `switch_remote_control`, y `colorwheel` reusan `furnitureaccess.CanManage` (rights de room o `room.furniture.any.manage`); `cannon` reusa `room.Unkickable` (ya existente, `RIGHTS.md`); el resto (`dice`, `pressureplate`, `colorplate`, `onewaygate`, `switch`, `vending`, `handitem*`) no requiere ninguna autorización especial más allá de estar presente en la sala — confirmado leyendo cada clase real, ninguna chequea rights salvo las tres ya nombradas.

---

## Parte 10 — Hot paths y allocations

- **`pressureplate`/`colorplate` corren en el camino de movimiento** (cada walk-on/walk-off) — el debounce de 100ms de pressureplate (Parte 3.1) existe precisamente para no recomputar de más cuando varias unidades entran/salen casi simultáneo; colorplate no lo necesita porque cada entrada/salida es un delta O(1), no una recomputación completa.
- **El resto (dice, colorwheel, onewaygate, switch, vending, handitem, cannon, multiheight) son acciones deliberadas del jugador**, no hot paths — sin necesidad de benchmarks dedicados, mismo criterio ya aplicado en `UPVOTES.md`/`TELEPORT.md`.
- **`Room.Schedule` (Parte 1.1) es una lista, no una cola de prioridad** — el volumen esperado de tareas programadas simultáneas por room (unos pocos cañones/dados/vendings en uso a la vez, como mucho) no justifica una estructura más compleja; se revisita con evidencia real, no preventivamente.
- **`ResettleUnitsOn` (multiheight, 5.2) es O(unidades sobre el footprint)**, no O(ocupantes de la sala) — mismo principio que el resto de esta serie de planes.

```go
// internal/realm/room/runtime/live/schedule_benchmark_test.go
func BenchmarkScheduleTick(b *testing.B) { ... } // costo de recorrer N tareas pendientes por tick, N realista (<20)
```

---

## Parte 11 — Testing (resumen transversal)

- Reloj inyectable en `Room` (mismo patrón ya usado en el resto del proyecto) para testear cada delay (100ms/500ms/750ms/1500ms/2000ms/3000ms) sin `time.Sleep` real.
- `Source` (Parte 1.3) inyectado con secuencias fijas para volver determinísticos dice/colorwheel/random_state/vending.
- Fakes de `furnitureaccess.CanManage`/`permission.Checker` para los tres tipos que sí autorizan (multiheight, switch_remote_control, colorwheel).
- Regresión explícita del bug de doble-ciclo de `switch_remote_control` (Parte 4.3) — el único test de este megaplan que existe específicamente para **no** replicar un comportamiento de la referencia.

---

## Parte 12 — Milestones de implementación

Mismo orden ya recomendado en `INTERACTIONS.md` Parte 10.5, con las primitivas compartidas primero:

1. **MP1 — Primitivas compartidas** (Parte 1): scheduler de tareas por room, `Source` inyectable, `WalkToNearestActivator`, `HandItemHolder` + packet. Sin esto, ningún tipo de este plan puede empezar.
2. **MP2 — Dice + Colorwheel + Random state** (Parte 2): depende de MP1. Colorwheel específicamente depende además de `INTERACTIONS.md` Milestone I3 (wall items) para su packet de estado — puede implementarse la lógica de negocio antes, pero no puede salir a producción sin I3.
3. **MP3 — Pressureplate + Colorplate** (Parte 3): depende de MP1 (scheduler para el debounce de pressureplate).
4. **MP4 — Onewaygate + Switch + Switch remote control** (Parte 4): depende de MP1 (`WalkToNearestActivator`, `ControlKind`) y de `TOGGLE-GATE.md` (Toggle genérico, que `switch`/`switch_remote_control` reusan tal cual).
5. **MP5 — Multiheight** (Parte 5): depende de MP1 y de la reconstrucción de fixtures ya usada por Gate (`TOGGLE-GATE.md` Parte 4.3).
6. **MP6 — Vending + Vending no-sides + Handitem + Handitem tile** (Parte 6): depende de MP1 (`HandItemHolder`, `WalkToNearestActivator`) — el grupo más grande, agrupado al final porque comparte más superficie nueva (el packet de hand item) que cualquier otro.
7. **MP7 — Cannon** (Parte 7): depende de MP1 y de `room.Unkickable`/`leavecmd.Handler` ya existentes — sin dependencias de los otros grupos, podría adelantarse si conviene por carga de trabajo, aunque el orden recomendado lo deja al final.

### Milestones futuros confirmados (fuera de este documento, no descartados)

- **Efectos de avatar por género** (vending, multiheight — Parte 6.1.1/5.1) — confirmado real en ambos, mismo motivo de exclusión ya documentado en `TOGGLE-GATE.md` Parte 1.4: requiere su propio sistema de efectos de principio a fin.
- **Wall items para colorwheel** (Parte 2.2) — bloqueado por `INTERACTIONS.md` Milestone I3, no por este plan.
- **`pressureplate_group`** — diferido a un futuro sistema de grupos, ya anotado en `INTERACTIONS.md`.
