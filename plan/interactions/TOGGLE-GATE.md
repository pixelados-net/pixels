# Plan: Toggle Genérico y Gate (`internal/realm/furniture/interactions/{toggle,gate}`)

Profundiza completamente los **Milestones I1 y I2** que `plan/INTERACTIONS.md` Partes 3-4 dejaron a nivel de boceto — cruzando ese diseño contra las clases reales de un fork activo (**Polaris-Emulator**, `github.com/duckietm/Polaris-Emulator`, Java): `InteractionDefault.java` (toggle N-estados, la base de casi toda la furniture clickeable) e `InteractionGate.java` (2-estados, bloquea mientras hay alguien encima). Encontré comportamiento real que el boceto original no capturaba — algunos adoptables ahora, otros deliberadamente diferidos con razón concreta.

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding, confirmado leyendo el código actual)

`plan/interactions/TELEPORT.md` ya se implementó, y construyó varias piezas que este plan reusa directo en vez de volver a diseñar:

| Ya existe | Dónde | Cómo lo usa este plan |
| --- | --- | --- |
| `world/furniture.Item.ExtraData string` (runtime) | `internal/realm/room/world/furniture/item.go` | El campo que la máquina de toggle/gate lee y escribe — ya existe, construido para teleport, genérico desde el día uno. |
| `Room.SetFurnitureExtraData(itemID, value) (Item, bool)` | `internal/realm/room/runtime/live/world.go` | Cambia el estado visual **sin** reconstruir fixtures — el camino correcto para Toggle (Parte 3), pero **no** alcanza para Gate (Parte 4.3, necesita reconstruir fixtures porque cambia walkability real). |
| `worldunit.ControlKind` | `internal/realm/room/world/unit/unit.go` | No lo necesita ninguno de los dos milestones de este plan (a diferencia de teleport) — se menciona solo para descartarlo explícitamente, no por omisión. |
| `furniture.used`/`walkedon`/`walkedoff` (eventos) | `internal/realm/furniture/events/*` | Sigue "planned for future interactions" para `walkedon`/`walkedoff` — este plan es el primer consumidor real de los tres. |
| `Definition.InteractionType`/`InteractionModesCount` | `internal/realm/furniture/model/definition.go` | Ya existen — Toggle es el primer código que los lee para algo más que sit/lay. |
| `furnitureaccess.CanManage(ctx, checker, room, playerID)` | `internal/realm/furniture/access/manage.go` | **Ya implementada** — `room.CanManageFurniture` (rights locales) O el nodo `room.furniture.any.manage` (staff). Mismo mecanismo que ya usan `move`/`place`/`pickup` — este plan lo reusa tal cual para autorizar el click de Toggle **y** de Gate (Parte 6: sin nodo de permiso nuevo). |
| `worldfurniture.Footprint(origin, width, length, rotation) []grid.Point` | `internal/realm/room/world/furniture/footprint.go` | Reusado para el chequeo de ocupación de **todo el footprint** del Gate (Parte 4.2), no solo el tile origen. |
| `surface.State` (`StateOpen`/`StateBlocked`/`StateSit`/`StateLay`) | `internal/realm/room/world/surface/section.go` | Sin ningún estado "bloqueado condicionalmente" — Gate lo resuelve reconstruyendo el fixture con `StateOpen`/`StateBlocked` según su extradata (Parte 4.3), no agregando un quinto valor al enum. |
| `outupdate.FloorItem` (packet `FLOOR_ITEM_UPDATE`, ya existente) | `networking/outbound/room/furniture/update` | Ya serializa `ExtraData` — confirmado leyendo `internal/realm/furniture/commands/move/command.go`. Ese packet es para posición/rotación; este plan sigue el criterio ya fijado en `INTERACTIONS.md` de **no** reusarlo para "solo cambió el estado" (Parte 5). |
| Ninguna tabla/columna nueva de esquema | — | A diferencia de teleport (que necesitó una tabla de pares) y de floorplan, Toggle/Gate no agregan ninguna tabla — todo el estado ya vive en `furniture_items.extra_data`, ya existente. |

---

## Parte 1 — Lo que confirmé en Polaris-Emulator (Java real)

Leí `InteractionDefault.java` e `InteractionGate.java` completos. Ambos son sustancialmente más ricos que el boceto de `INTERACTIONS.md`:

### 1.1 `canToggle`: rights de room, o un "área alquilada" que Pixels no tiene

```java
public boolean canToggle(Habbo habbo, Room room) {
    if (room.hasRights(habbo)) return true;
    if (!habbo.getHabboStats().isRentingSpace()) return false;
    HabboItem rentSpace = room.getHabboItem(habbo.getHabboStats().rentedItemId);
    return rentSpace != null && RoomLayout.squareInSquare(/* el toggle está dentro del footprint del área alquilada */);
}
```
Confirma exactamente lo que `INTERACTIONS.md` Parte 3 ya anticipaba entre paréntesis ("o que el item está dentro de un área alquilada suya, si/cuando exista ese concepto — no en este milestone") — el fallback de "área alquilada" es real, pero Pixels no tiene ningún concepto de "rentable space" todavía. Se confirma como dependencia futura concreta, no una suposición (Parte 9, milestones futuros).

### 1.2 Parseo defensivo de extradata corrupta

```java
try {
    currentState = Integer.parseInt(this.getExtradata());
} catch (NumberFormatException e) {
    LOGGER.error("Incorrect extradata ({}) for item ID ({})...", ...);
}
// currentState queda en 0 (su valor de declaración), sigue adelante sin abortar
```
Si `extra_data` está corrupto (no debería pasar nunca, pero un dato viejo/migrado mal podría tenerlo), Polaris **no** aborta el click — loguea el problema y sigue como si el estado actual fuera `0`. Pixels adopta el mismo criterio defensivo (Parte 3.1).

### 1.3 Mover furniture con alguien encima dispara un "walked off" sintético

```java
@Override
public void onMove(Room room, RoomTile oldLocation, RoomTile newLocation) {
    if (/* la vieja posición no tiene un roller */) {
        for (RoomUnit unit : room.getRoomUnits()) {
            if (unit estaba en la furniture en oldLocation && ya no lo está en newLocation)
                this.onWalkOff(unit, room, new Object[]{oldLocation, newLocation});
        }
    }
}
```
Hallazgo real que `INTERACTIONS.md` no tenía: cuando el comando `move` (ya implementado en Pixels) reubica una furniture toggleable que tenía unidades paradas encima, la referencia dispara el mismo evento de "se bajó" que dispararía caminar fuera — evita que una unidad quede "pisando" fantasma un efecto/estado que ya no le corresponde. Se excluye deliberadamente si la furniture es un roller (el roller ya reubica a la unidad junto con el mueble, disparar esto también sería doble). Este plan lo adopta (Parte 3.5) como parte de la integración con el comando `move` ya existente.

### 1.4 Efecto de avatar por género al pisar/despisar — feature real, deliberadamente fuera de este plan

```java
if (this.getBaseItem().getEffectF() > 0 || this.getBaseItem().getEffectM() > 0) {
    // onWalkOn: si hay un efecto configurado para el género del jugador, se lo otorga
    // onWalkOff: si no hay otra furniture con el mismo efecto en el tile de destino, se lo quita
}
```
Hallazgo real no capturado en `INTERACTIONS.md`: `InteractionDefault` no es *solo* un ciclo de estados — furniture como sillas/pisos especiales puede otorgar un efecto visual de avatar (glow, animación) específico por género al pararse encima, y quitarlo al bajarse (con lógica para no parpadear si el tile siguiente también otorga el mismo efecto). Pixels **no tiene ningún sistema de efectos de avatar** todavía (ni la networking, ni el concepto) — se documenta como confirmado y real, pero **fuera de alcance de este plan** (Parte 9, milestones futuros): implementarlo a medias sin el resto del sistema de efectos sería peor que no tocarlo. El toggle de estado en sí (lo que este plan sí construye) funciona perfectamente sin esto.

### 1.5 `allowWiredResetState`: `true` para ambos, a diferencia de teleport

```java
// InteractionDefault e InteractionGate:
public boolean allowWiredResetState() { return true; }
// InteractionTeleport (ya implementado):
public boolean allowWiredResetState() { return false; }
```
Confirma que Toggle/Gate SÍ deberían poder resetearse por un futuro wired genérico de "reset all states" — a diferencia de teleport, que se protegió explícitamente de esto. Sin wired implementado todavía, esto no tiene ningún efecto práctico hoy — se documenta como un dato a tener en cuenta el día que exista wired, no una feature a construir ahora.

### 1.6 Gate: más estricto y con más responsabilidades que un Toggle común

```java
public boolean isWalkable() { return this.getExtradata().equals("1"); }

public void onClick(...) {
    if (client != null && !room.hasRights(client.getHabbo()) && !executedByWired) return; // SIN fallback de área alquilada

    for (RoomTile tile : room.getLayout().getTilesAt(/* todo el footprint del gate */))
        if (room.hasHabbosAt(tile.x, tile.y)) return; // cualquier tile del footprint bloquea, no solo el origen

    this.setExtradata((Integer.parseInt(this.getExtradata()) + 1) % 2 + "");
    room.updateTile(room.getLayout().getTile(this.getX(), this.getY())); // recalcula walkability real, no solo visual
    room.updateItemState(this);
}
```
Tres diferencias reales que `INTERACTIONS.md` simplificaba de más:
1. **Sin el fallback de "área alquilada"** de `InteractionDefault` — la autorización de Gate es más estricta (solo rights de room o staff).
2. **El chequeo de "alguien encima" cubre todo el footprint del gate** (`getTilesAt` con ancho/alto/rotación), no solo el tile donde está el origen — un gate de 2×1 con alguien parado en CUALQUIERA de sus dos tiles bloquea el toggle.
3. **`isWalkable()` deriva directamente del extradata** (`"1"` = caminable, `"0"` = bloqueado) — no es un flag separado, y togglear un gate **recalcula la walkability real del tile** (`room.updateTile(...)`), no solo transmite un número cosmético — a diferencia de un Toggle normal (lámpara), donde cambiar de estado nunca afecta si se puede caminar encima.

---

## Parte 2 — Arquitectura: un comando compartido, dos comportamientos

Mismo criterio ya fijado en `INTERACTIONS.md` Parte 2 ("pocas formas compartidas, no una clase por interacción"): un solo comando de protocolo (`furniture.interact`) para ambos milestones — lo que cambia es una función de política por `interaction_type`, no dos comandos paralelos.

```go
// internal/realm/furniture/commands/interact/command.go

// Behavior resolves interaction_type-specific toggle rules — Toggle (I1) and Gate
// (I2) each provide one, resolved by definition.InteractionType at click time.
type Behavior interface {
    // CanToggle reports whether the current room/world state allows a state
    // change right now (Gate: nadie parado encima; Toggle: siempre true).
    CanToggle(active *roomlive.Room, item worldfurniture.Item) bool
    // Apply computes the next state and whatever fixture rebuild it implies.
    Apply(active *roomlive.Room, item worldfurniture.Item, definition furnituremodel.Definition) (next string, rebuildFixtures bool, err error)
}
```
El comando en sí (resolución de actor, autorización vía `furnitureaccess.CanManage`, broadcast, persistencia, evento) es **idéntico** para ambos — solo `Behavior` cambia. Esto evita exactamente la duplicación que tendría "un comando de toggle y otro de gate por separado".

---

## Parte 3 — Milestone I1: Toggle genérico

### 3.1 Ciclo de estados, con el parseo defensivo de la Parte 1.2

```go
// internal/realm/furniture/interactions/toggle/behavior.go

func (toggleBehavior) CanToggle(*roomlive.Room, worldfurniture.Item) bool { return true }

func (toggleBehavior) Apply(active *roomlive.Room, item worldfurniture.Item, definition furnituremodel.Definition) (string, bool, error) {
    if definition.InteractionModesCount <= 1 {
        return item.ExtraData, false, ErrNoStatesToCycle // no hay nada que ciclar — el comando no persiste ni broadcastea nada
    }
    current, err := strconv.Atoi(item.ExtraData)
    if err != nil {
        current = 0 // extradata corrupto o vacío — mismo criterio defensivo confirmado en Polaris (Parte 1.2), no aborta
    }
    next := (current + 1) % definition.InteractionModesCount

    return strconv.Itoa(next), false, nil // false: Toggle nunca reconstruye fixtures, a diferencia de Gate
}
```
`InteractionModesCount <= 1` corta temprano (ni siquiera vale la pena escribir/broadcastear un estado que no cambia visualmente) — resultado práctico idéntico al `getStateCount() > 0` de la referencia, sin gastar un write+broadcast de más.

### 3.2 Autorización — `furnitureaccess.CanManage`, sin nodo nuevo

Reusa exactamente lo que `move`/`place`/`pickup` ya usan (Parte 0) — rights de room o `room.furniture.any.manage`. El fallback de "área alquilada" de la Parte 1.1 queda anotado como dependencia futura, no bloqueante.

### 3.3 Broadcast — `SetFurnitureExtraData`, sin reconstruir fixtures

```go
updated, changed := active.SetFurnitureExtraData(itemID, next)
```
Toggle nunca cambia footprint/altura/walkability — el camino barato (ya construido para teleport, Parte 0) alcanza sin ningún trabajo adicional. Se persiste el nuevo `ExtraData` en la fila durable (síncrono, a diferencia del parpadeo intermedio de teleport — un toggle de lámpara no tiene "estados intermedios" que descartar, cada click es un estado final real que vale la pena guardar) y se broadcastea `furniture/state` (Parte 5) a todo el room.

### 3.4 Integración con `move` — el "walked off" sintético de la Parte 1.3

`internal/realm/furniture/commands/move/command.go` (ya existente) gana, tras `ReloadFurniture`, un chequeo: si el item movido es togglable y tenía unidades sobre su footprint viejo que ya no están sobre el nuevo, publica `furniture.walkedoff` para cada una — **excepto** si el item es un roller (reservado para `INTERACTIONS.md` Parte 7, Milestone I5, que ya reubica unidades como parte de su propio mecanismo).

### 3.5 Efectos de avatar — reconocidos, diferidos (Parte 1.4)

No se implementa en este plan. Se anota `interaction_type = "toggle"` como el punto de enganche exacto donde un futuro sistema de efectos de avatar se conectaría (`onWalkOn`/`onWalkOff` ya van a existir como eventos reales gracias a este milestone) — sin construir el sistema de efectos en sí sin necesidad confirmada.

### 3.6 Eventos

```
furniture.used       {PlayerID, ItemID, RoomID}              // ya scaffoldeado, este plan lo publica de verdad
furniture.walkedon    {PlayerID, ItemID, RoomID}              // disparado por el pathfinder al completar un paso sobre el footprint (integración nueva, Parte 7)
furniture.walkedoff   {PlayerID, ItemID, RoomID}              // disparado por walk normal, o sintético desde move (3.4)
```

### 3.7 Tests

- `InteractionModesCount = 1` (o `0`) → click no persiste ni broadcastea nada, sin error.
- `InteractionModesCount = 3` → tres clicks ciclan `"0"` → `"1"` → `"2"` → `"0"`.
- `ExtraData` corrupto (`"abc"`) → se trata como `"0"`, el click igual avanza a `"1"`, sin abortar (regresión directa de la Parte 1.2).
- Click sin `CanManage` → rechazo, mismo error que `move`/`place`/`pickup`.
- Mover un item togglable con una unidad parada encima, hacia un tile donde esa unidad ya no queda sobre el footprint → `furniture.walkedoff` se publica (Parte 3.4); moviendo un roller con la misma situación → NO se publica (excluido a propósito).
- `furniture.used` se publica con el payload correcto en cada click exitoso.

---

## Parte 4 — Milestone I2: Gate

### 4.1 Estado binario, walkability derivada del propio estado

```go
// internal/realm/furniture/interactions/gate/behavior.go

func (gateBehavior) Apply(active *roomlive.Room, item worldfurniture.Item, definition furnituremodel.Definition) (string, bool, error) {
    current := item.ExtraData
    if current == "" {
        current = "0"
    }
    next := "0"
    if current == "0" {
        next = "1"
    }

    return next, true, nil // true: Gate SIEMPRE reconstruye fixtures — cambia walkability real
}
```
`"1"` = abierto/caminable, `"0"` = cerrado/bloqueado — exactamente como la referencia (Parte 1.6), sin ningún campo separado de "está abierto".

### 4.2 Ocupación: todo el footprint, no solo el tile origen

```go
func (gateBehavior) CanToggle(active *roomlive.Room, item worldfurniture.Item) bool {
    for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
        if active.HasOccupantAt(point) { // nuevo helper liviano sobre Room.Units(), ver 4.4
            return false
        }
    }

    return true
}
```
Reusa `worldfurniture.Footprint` (ya existente, Parte 0) — para un gate 1×1 el comportamiento es idéntico a "el tile origen"; para un gate más grande, cualquier tile de su huella bloquea el toggle, confirmado real (Parte 1.6, punto 2).

### 4.3 Reconstrucción de fixture: la diferencia real con Toggle

```go
// internal/realm/furniture/interactions/gate/fixture.go

// fixtureState maps the gate's own extradata to a real walkable surface.State,
// closing the gap plan/INTERACTIONS.md Parte 1 already flagged ("surface.State no
// tiene hoy ningún estado de bloqueado condicionalmente") — no se agrega un quinto
// valor al enum, se resuelve seteando StateOpen o StateBlocked al reconstruir el
// fixture del propio gate, exactamente como cualquier otra furniture ya hace.
func fixtureState(extraData string) surface.State {
    if extraData == "1" {
        return surface.StateOpen
    }

    return surface.StateBlocked
}
```
El comando de `interact` para Gate llama `active.ReloadFixtures(itemID, newFixtures)` (ya existente, usado hoy por `move`) en vez de `SetFurnitureExtraData` — esta es la razón concreta por la que `Behavior.Apply` retorna `rebuildFixtures = true` para Gate y `false` para Toggle (Parte 2): togglear un gate cambia si el tile es caminable de verdad, así que el pathfinder (`world/path`) necesita ver el cambio reflejado en el resolver de superficie, no solo en un número cosmético que viaja al cliente.

### 4.4 Autorización: sin fallback de área alquilada (a propósito, decisión consciente)

`furnitureaccess.CanManage` de nuevo (Parte 0) — **la misma función que Toggle**, no una más laxa ni más estricta. Esto es una decisión deliberada de Pixels, distinta de la asimetría real confirmada en Polaris (Gate ahí es más estricto que Default porque Default tiene el fallback de área alquilada que Gate no tiene): como Pixels **no implementa** el fallback de área alquilada en ninguno de los dos (Parte 3.2/1.1), no hay ninguna asimetría real que preservar — ambos terminan usando exactamente la misma autorización, y se documenta explícitamente para que nadie intente "restaurar" una diferencia que en Pixels no tiene motivo de existir todavía.

`active.HasOccupantAt(point grid.Point) bool` — helper nuevo y liviano (`Room.Units()` ya existente, ver si algún `UnitSnapshot.Position.Point == point`) para el chequeo de la Parte 4.2, sin necesitar ninguna estructura de índice nueva (el número de unidades por room es chico, un scan lineal sobre `Units()` alcanza).

### 4.5 Eventos

Mismos tres eventos que Toggle (Parte 3.6) — Gate es, a nivel de protocolo/eventos, indistinguible de un Toggle de 2 estados; lo único que cambia es la política interna (`Behavior`).

### 4.6 Tests

- Toggle de un gate libre (nadie parado en su footprint) alterna `"0"`↔`"1"` y `ReloadFixtures` corre con el `surface.State` correcto en cada dirección.
- Alguien parado en CUALQUIER tile del footprint (no solo el de origen, para un gate >1×1) rechaza el toggle sin cambiar estado ni broadcastear.
- Tras abrir un gate (`"1"`), el pathfinder (`world/path`) permite atravesarlo; tras cerrarlo (`"0"`), lo bloquea — test de integración contra el resolver real, no solo contra el extradata.
- Autorización idéntica a Toggle: mismo test de tabla que 3.7, reusado (no un test paralelo que pueda desincronizarse).

---

## Parte 5 — Packets (compartidos entre I1 y I2)

| Dirección | Paquete | Contenido | Header |
| --- | --- | --- | --- |
| Inbound | `furniture/interact` | `itemId int32` | TBD — a confirmar contra el composer real de Nitro (`RoomUnitUseFurnitureComposer`/equivalente) |
| Outbound | `room/furniture/state` | `itemId int32`, `state string` | TBD — deliberadamente **distinto** de `FLOORPLAN_ITEM_UPDATE`/`FLOOR_ITEM_UPDATE` (posición/rotación, ya existente), mismo criterio que Arcturus separa `ItemStateComposer` de `FloorItemUpdateComposer` |

Ningún packet nuevo más allá de estos dos — teleport (Parte 0) ya demostró que el mismo canal (`ExtraData` + un packet de estado) alcanza para una máquina de estados mucho más compleja que esta.

---

## Parte 6 — Nodos de permiso

**Ninguno nuevo.** `room.furniture.any.manage` (ya existente, `internal/realm/furniture/access/manage.go`) cubre ambos milestones — Toggle y Gate no introducen ningún concepto de autorización que no exista ya para `move`/`place`/`pickup`.

---

## Parte 7 — `walkedon`/`walkedoff`: quién los dispara de verdad

Hoy son scaffolds sin publicador (Parte 0). Este plan es el primero en conectarlos:
- **`walkedon`**: se publica desde el motor de movimiento (`world/path`/`worldunit`, donde un paso de camino se asienta sobre un tile) cuando el tile de llegada tiene una furniture cuyo `Definition.InteractionType` es `"toggle"` o `"gate"` — un solo punto de enganche genérico, no un chequeo especial por tipo de interacción (el motor de movimiento no necesita saber qué hace cada tipo con el evento, solo que alguien puede estar escuchando).
- **`walkedoff`**: análogo, al abandonar ese tile — más el caso sintético de la Parte 3.4 (mover la furniture debajo de alguien).

---

## Parte 8 — Hot paths y allocations

- **El click de toggle/gate no es un hot path** (acción deliberada del jugador, no algo por tick) — sin necesidad de ningún benchmark dedicado, mismo criterio que `UPVOTES.md`.
- **`walkedon`/`walkedoff` SÍ corren en el camino de movimiento** (cada paso de cada unidad que camina, potencialmente) — el chequeo debe ser barato: `Definition.InteractionType` ya está resuelto en el `worldfurniture.Item` cacheado del tile (no requiere ninguna consulta nueva), así que el costo real es solo publicar el evento cuando aplica, no evaluarlo — se mide si en la práctica hace falta, no se optimiza preventivamente sin evidencia (mismo principio ya aplicado en el resto de esta serie de planes).
- **`HasOccupantAt` (Gate, 4.4)** es O(ocupantes de la sala) — aceptable dado que corre solo al togglear un gate (poco frecuente), no en cada tick ni en cada paso de movimiento.

---

## Parte 9 — Milestones de implementación

1. **TG1 — Comando compartido `furniture.interact` + `Behavior`** (Parte 2): resolución de actor, autorización (`furnitureaccess.CanManage`, sin nodo nuevo), wiring del packet inbound.
2. **TG2 — Toggle genérico** (Parte 3): `toggleBehavior`, broadcast vía `SetFurnitureExtraData`, packet `furniture/state` — depende de TG1.
3. **TG3 — Integración con `move`** (Parte 3.4): el "walked off" sintético al mover furniture togglable con unidades encima — depende de TG2.
4. **TG4 — Gate** (Parte 4): `gateBehavior`, reconstrucción de fixture (`ReloadFixtures` + `fixtureState`), chequeo de footprint completo (`HasOccupantAt`) — depende de TG1, puede ir en paralelo a TG2/TG3.
5. **TG5 — `walkedon`/`walkedoff` reales** (Parte 7): conecta los eventos scaffold al motor de movimiento — depende de TG2/TG4 para tener algún `interaction_type` real que consumirlos.
6. **TG6 — Tests de integración cruzada** (Parte 3.7, 4.6): pathfinder real respetando el estado de un Gate; regresión de move+walkedoff.

### Milestones futuros confirmados (fuera de este documento, no descartados)

- **Fallback de "área alquilada" para `CanToggle`** (Parte 1.1/3.2) — depende de un futuro concepto de "rentable space" que Pixels no tiene; se agrega el día que exista, sin cambiar la forma de `Behavior.CanToggle`.
- **Sistema de efectos de avatar por género** (Parte 1.4/3.5) — confirmado real en la referencia, pero requiere su propio sistema de principio a fin (networking, persistencia de qué efecto tiene activo cada unidad); se define en su propio plan si aparece necesidad real, no se construye a medias acá.
- **Reset de estado vía wired** (Parte 1.5) — sin efecto práctico hasta que exista un sistema de wired en Pixels (excluido explícitamente en todo `INTERACTIONS.md`).
