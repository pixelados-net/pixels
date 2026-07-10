# Plan: Configuración del Room (`internal/realm/room/settings`, correcciones a `navigator/commands/info`)

Este plan implementa completamente la **Parte 3** de `plan/REMAINING-ROOMS.md` (pedir/guardar la configuración de una sala: nombre, descripción, tags, categoría, modo de puerta, moderación, chat, thickness), y cierra el círculo de sincronización — qué le llega a quién, cuándo, y por qué canal — para que el nombre/info de una sala se vea igual en el navegador, en el panel de información ("room info"), y para los ocupantes actuales, sin importar cuál de los tres se consulte. También corrige dos bugs reales encontrados durante el research de este plan en el `ROOM_INFO` ya implementado.

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding, confirmado leyendo el código actual)

`plan/rooms/ENTRY.md` y `plan/rooms/RIGHTS.md` ya están implementados de punta a punta. Este plan parte de ahí — reusa `rights.Service`/`moderation.Service`/los nodos de permiso ya existentes en `internal/realm/room/permissions.go`, y el helper `internal/realm/room/commands/control` para resolver actor+room.

| Ya existe | Dónde | Nota |
| --- | --- | --- |
| `GET_GUEST_ROOM` (2230, inbound) → `ROOM_INFO` (687, outbound) | `internal/realm/navigator/commands/info`, `networking/{in,out}bound/navigator/roominfo` | **El panel de "información de la sala" ya funciona de punta a punta**: resuelve el room, sus tags, la ocupancia en vivo, y arma el packet completo (nombre, dueño, descripción, door mode, trade mode, score, tags, `AllowPets`, moderación, chat). Este plan no lo reconstruye — lo corrige y lo completa (Parte 3). |
| `navprojection.RoomCard(room, userCount, ranking, tags) roomcard.Card` | `internal/realm/navigator/projection/card.go` | La misma función arma tanto las tarjetas de resultados del navegador como el cuerpo de `ROOM_INFO` — **lee siempre en vivo desde `roommodel.Room`, sin ningún cache de por medio**. Esto es clave para la Parte 7 (sincronización): en cuanto `settings/save` persiste, cualquier búsqueda o `ROOM_INFO` *nuevo* ya ve los datos frescos, sin invalidar nada. El problema de sincronización real es otro (Parte 7). |
| **Bug real #1**: `moderation(room)` mapea mal | `internal/realm/navigator/commands/info/packet.go:39` | `AllowMute: int32(room.ChatProtection)` — usa la columna de chat, no `room.ModerationMute`. `AllowKick`/`AllowBan` están **hardcodeados a `0`**, sin mapear `room.ModerationKick`/`room.ModerationBan` (que ya existen en `model.Room` desde `plan/rooms/RIGHTS.md`). Se corrige en la Parte 3. |
| **Bug real #2**: `CanMute: false` hardcodeado | `internal/realm/navigator/commands/info/command.go:84` | Nunca refleja si el jugador que pide la info realmente puede mutear en esa sala — `moderation.Service` (ya implementado) tiene exactamente la lógica para resolver esto. Se corrige en la Parte 3. |
| **Gap real**: `AllInRoomMuted` nunca se setea | mismo archivo | El campo del packet existe, pero el toggle de "silenciar todo el chat de la sala" (`REMAINING-ROOMS.md` Parte 3) no existe en absoluto todavía — greenfield, Parte 5 de este plan. |
| `model.Room` ya tiene **todos** los campos de settings | `internal/realm/room/model/room.go` | `Name`, `Description`, `DoorMode`, `PasswordHash`, `TradeMode`, `AllowWalkthrough`, `AllowPets`, `AllowPetsEat`, `HideWalls`, `WallThickness`, `FloorThickness`, `ChatMode/Weight/Speed/Distance/Protection`, `ModerationMute/Kick/Ban`, `CategoryID` — el modelo está completo. Lo que falta es el método de escritura y los comandos de protocolo alrededor. |
| `roomservice.Manager` — sin `Update` | `internal/realm/room/service/contract.go` | Confirmado: solo `Creator`+`Finder`+`SoftDelete`+`ListCategories`. Greenfield para la Parte 2. |
| `rights.Service.HasRights` / `moderation.Service` / nodos `.own`/`.any` | `internal/realm/room/rights`, `internal/realm/room/moderation`, `internal/realm/room/permissions.go` | Ya implementados (`plan/rooms/RIGHTS.md`). Este plan reusa el mismo patrón de fórmula de autorización, y agrega dos nodos nuevos al mismo archivo `permissions.go` (Parte 6) — no crea uno nuevo. |
| `internal/realm/room/commands/control.Actor`/`MatchRoom` | `internal/realm/room/commands/control/session.go` | Helper ya existente para resolver actor+room-actual y validar que un packet apunta al room correcto — reusado por `settings`/`wordfilter`/`mute` en vez de reimplementar la resolución de sesión. |
| Ningún wordfilter (ni global, ni por-room) | grep sin resultados en todo el repo | Greenfield total — ningún otro consumidor de "validar texto de usuario contra palabras prohibidas" existe todavía en el proyecto. |
| Ningún toggle de "silenciar todo" en `roomlive.Room` | grep sin resultados | Greenfield, Parte 5. |
| `rooms.version` (optimistic locking) ya establecido | `internal/realm/room/repository/room.go` (`softDeleteRoomSQL` usa `version = version + 1`) | Mismo patrón a reusar para el `update` de settings — evita pisar una escritura concurrente sin darse cuenta. |

---

## Parte 1 — Pedir el estado actual (`settings/request` → `settings/current`)

### 1.1 Autorización

Misma fórmula ya establecida por `plan/rooms/RIGHTS.md` para moderación, aplicada acá con los nodos nuevos de la Parte 6:

```go
func (service *Service) authorizeSettings(ctx context.Context, room roommodel.Room, actorID int64) (bool, error) {
    if allowed, err := service.hasPermission(ctx, actorID, room.SettingsAnyManage); err != nil || allowed {
        return allowed, err
    }
    if room.OwnerPlayerID == actorID {
        return service.hasPermission(ctx, actorID, room.SettingsOwnManage)
    }
    hasRights, err := service.rights.HasRights(ctx, room.ID, actorID)
    if err != nil || !hasRights {
        return false, err
    }

    return service.hasPermission(ctx, actorID, room.SettingsOwnManage)
}
```

Esto honra explícitamente lo que `REMAINING-ROOMS.md` Parte 3 ya dejó decidido en su propia sección de tests ("un jugador sin derechos ni ser el dueño no puede guardar settings... acá sí tiene sentido: ni dueño ni derechos → no puede") — a diferencia del bug invertido de Arcturus en 1.3 (`RequestRoomSettingsEvent` exigía **ambas** cosas por un error de escritura), acá la condición correcta es dueño **O** rights-holder (ambos además sujetos al nodo `.own`, para poder revocarle a una cuenta puntual la capacidad de tocar la configuración de su propia sala sin tocar el resto del sistema de staff — mismo criterio ya aplicado en moderación).

### 1.2 Comando y packet

`internal/realm/room/commands/settings/request` — sin payload más que "mi room actual" (resuelto vía `control.Actor`).

```go
// networking/outbound/room/settings/current/packet.go

// Params contains the full editable room settings snapshot.
type Params struct {
    Name             string
    Description      string
    CategoryID       int32
    Tags             []string
    MaxUsers         int32
    DoorMode         int32
    TradeMode        int32
    AllowWalkthrough bool
    AllowPets        bool
    AllowPetsEat     bool
    HideWalls        bool
    WallThickness    int32
    FloorThickness   int32
    ChatMode         int32
    ChatWeight       int32
    ChatSpeed        int32
    ChatDistance     int32
    ChatProtection   int32
    ModerationMute   int32
    ModerationKick   int32
    ModerationBan    int32
    HasPassword      bool // nunca se manda el hash ni la contraseña en claro de vuelta
}
```

`HasPassword` (no `PasswordHash`, y mucho menos la contraseña en texto plano) — el diálogo de settings solo necesita saber "¿ya hay una seteada?" para decidir si exigir una nueva al cambiar a `DoorModePassword` (Parte 2.3); jamás se devuelve el hash ni nada derivado de la contraseña real al cliente.

### 1.3 Tests

- Dueño pide settings → recibe el snapshot completo.
- Rights-holder con `SettingsOwnManage` pide settings → recibe el snapshot.
- Rights-holder sin `SettingsOwnManage` (revocado puntualmente) → rechazado, aunque tenga derechos.
- Staff con `SettingsAnyManage` → recibe el snapshot de cualquier sala, sin ser dueño ni tener rights.
- Ningún jugador sin ninguna de las tres condiciones → rechazado.
- `HasPassword` refleja `true`/`false` según `PasswordHash != nil`, nunca expone el hash.

---

## Parte 2 — Guardar settings (`settings/save`)

### 2.1 `roomservice.Manager` gana `Update`

```go
// internal/realm/room/service/contract.go (extendido)

// Updater persists room configuration changes.
type Updater interface {
    // Update applies a partial room configuration change, requiring the current version
    // for optimistic concurrency (mismo patrón que SoftDelete ya usa internamente).
    Update(ctx context.Context, roomID int64, expectedVersion int64, params UpdateParams) (roommodel.Room, error)
}
```

```go
// internal/realm/room/service/update.go

// UpdateParams contains optional room configuration fields — nil/zero-value pointers
// leave the corresponding column untouched, mismo patrón ya usado por
// permission.UpdateGroupParams (plan/PERMISSIONS.md) para mutaciones parciales.
type UpdateParams struct {
    Name             *string
    Description      *string
    CategoryID       **int64
    Tags             *[]string
    MaxUsers         *int
    DoorMode         *roommodel.DoorMode
    Password         *string // texto plano recibido del cliente, nunca persistido tal cual — ver 2.3
    TradeMode        *roommodel.TradeMode
    AllowWalkthrough *bool
    AllowPets        *bool
    AllowPetsEat     *bool
    HideWalls        *bool
    WallThickness    *int
    FloorThickness   *int
    ChatMode         *int16
    ChatWeight       *int16
    ChatSpeed        *int16
    ChatDistance     *int16
    ChatProtection   *int16
    ModerationMute   *roommodel.ModerationPolicy
    ModerationKick   *roommodel.ModerationPolicy
    ModerationBan    *roommodel.ModerationPolicy
}
```

### 2.2 Validación

La mayoría de las restricciones de longitud/rango **ya las hace Postgres** (`rooms_name_length_chk`, `rooms_description_length_chk`, `rooms_door_mode_chk`, `rooms_max_users_chk`, `rooms_trade_mode_chk`) — el `service.Update` no las duplica, deja que la base rechace y traduce el error de constraint a un código de dominio. Lo que Postgres **no puede validar**, el service sí:

- **Palabras prohibidas en nombre/descripción/tags**: contra un filtro de contenido **global** (no el wordfilter por-room de la Parte 4, que es un concepto distinto — ver esa parte). Este filtro global no existe todavía en ningún lado del proyecto — es la primera vez que algo necesita "validar texto libre de usuario contra una lista de palabras prohibidas". Se modela como una interfaz angosta, con degradación nula explícita (mismo patrón ya usado por `entry.Service.rights`/`entry.Service.bans`, que son `nil`-safe hasta que alguien los implemente):
  ```go
  // ProfanityChecker validates free-text user content against prohibited words.
  type ProfanityChecker interface {
      Contains(ctx context.Context, text string) (bool, error)
  }
  ```
  Con `service.profanity == nil`, la validación se saltea (siempre `false`) — no bloquea este plan a que exista un filtro global compartido; cuando exista (en cualquier realm, no necesariamente en `room`), se inyecta acá sin cambiar la forma de `Update`.
- **Tags restringidos a staff**: un prefijo/lista reservada (a definir en implementación, ej. tags que empiecen con `official:`) solo puede aparecer si el actor tiene `room.SettingsAnyManage` — de lo contrario, error de campo específico.
- **Contraseña requerida al cambiar a `DoorModePassword`**: si `params.DoorMode` apunta a `DoorModePassword` y ni la sala ya tiene `PasswordHash` ni `params.Password` trae una nueva, falla con un código de error de campo específico (`ErrPasswordRequired`) — igual que Arcturus.
- **Consistencia de `MaxUsers`**: acotado por el check de Postgres (1-100); si además se quiere acotar por la capacidad real del layout (algunos layouts son más chicos que 100 tiles caminables), **a confirmar en implementación** contra `layout.Layout`/`room/world/grid` si exponen ya un techo de capacidad utilizable — no bloquea este plan, se valida solo contra el rango genérico mientras tanto.

### 2.3 Contraseña: nunca se persiste en claro

`params.Password` (si viene) se hashea con bcrypt (mismo mecanismo ya usado por `internal/realm/room/entry/password.go`, **reusado tal cual**, no reinventado) antes de escribir `PasswordHash` — el service de settings nunca guarda el texto plano en ningún lado, ni siquiera transitoriamente en un log (mismo cuidado de redacción ya aplicado en `enter.Command.MarshalLogObject`; `settings.Command` gana el mismo tratamiento).

### 2.4 Persistencia

```sql
update rooms set
    name = coalesce($2, name),
    description = coalesce($3, description),
    -- ... resto de columnas, todas ya existentes
    version = version + 1,
    updated_at = now()
where id = $1 and version = $expected_version and deleted_at is null
returning ...;
```

Sin migración nueva para los campos en sí (todos ya existen en `rooms`, confirmado en Parte 0) — solo se agrega la capacidad de escribirlos, que hoy no existe. Si `expectedVersion` no matchea (alguien más guardó settings mientras tanto), el `Update` retorna `ErrVersionConflict` — el cliente vuelve a pedir el estado actual (`settings/request`) y reintenta, en vez de pisar un cambio concurrente en silencio.

### 2.5 Tags: reemplazo completo, no diff incremental

Guardar settings reemplaza el conjunto completo de tags de la sala (`delete from room_tags where room_id = $1` + insert de los nuevos, en la misma transacción que el `update rooms`) — mismo criterio que Arcturus (el diálogo de settings manda la lista completa de tags cada vez, no operaciones de agregar/quitar una por una).

### 2.6 Broadcasts tras guardar (los 3 + la confirmación, tal como especifica `REMAINING-ROOMS.md` Parte 3)

| # | Packet | A quién | Contenido |
| --- | --- | --- | --- |
| 1 | `outbound/room/settings/updated` | Todos los ocupantes activos del room (broadcast, `broadcast.RoomPacket`) | Todo lo que no es thickness ni chat: `Name`, `Description`, `DoorMode`, `TradeMode`, `AllowWalkthrough`, `AllowPets`, `AllowPetsEat`, `CategoryID`, `Tags`, `ModerationMute/Kick/Ban`. **Este es el packet que sincroniza el nombre en tiempo real** — a diferencia de Arcturus (`// TODO Find packet for update room name.`, nunca resuelto, confirmado en el research), Pixels reusa este mismo packet para el nombre en vez de dejarlo sin sincronizar: no hay ninguna razón de protocolo real para separarlo, el cliente ya recibe el payload completo y puede releer el nombre de ahí. |
| 2 | `outbound/room/thickness/updated` | Todos los ocupantes activos | `WallThickness`, `FloorThickness`, `HideWalls` — separado porque son campos puramente de render 3D, mismo criterio que Arcturus. |
| 3 | `outbound/room/chatsettings/updated` | Todos los ocupantes activos | `ChatMode/Weight/Speed/Distance/Protection`. |
| 4 | `outbound/room/settings/saved` | Solo quien guardó (respuesta directa, no broadcast) | Confirmación + el `Room` actualizado completo (mismo shape que `settings/current`, Parte 1.2) — evita que el que guardó tenga que volver a pedir `settings/request` para ver su propio cambio reflejado. |

Los 3 primeros se emiten **incluso si el que guardó ya no está físicamente parado en el room activo en ese instante** (ej. guardó desde un futuro editor externo) — se resuelven contra `roomlive.Registry.Find(roomID)`; si el room no está activo (nadie adentro), no hay a quién broadcastear y no pasa nada, lo cual es correcto (nadie hay para sincronizar).

### 2.7 Tests

- Guardar cada campo individualmente actualiza exactamente esa columna y dispara los broadcasts correctos, sin tocar el resto.
- Guardar con nombre/descripción/tag que matchea el `ProfanityChecker` (cuando esté configurado) falla sin persistir nada; con `ProfanityChecker == nil`, la validación se saltea y persiste igual (documenta la degradación explícita).
- Guardar con un tag reservado a staff, sin `SettingsAnyManage` → falla; con `SettingsAnyManage` → persiste.
- Cambiar a `DoorModePassword` sin `PasswordHash` existente y sin `Password` nuevo → `ErrPasswordRequired`, nada persiste.
- Cambiar a `DoorModePassword` mandando `Password` nuevo → el hash se persiste, nunca el texto plano (assertion directa sobre la fila en Postgres en el test de repository).
- Conflicto de versión (`expectedVersion` desactualizado) → `ErrVersionConflict`, nada persiste.
- Guardar sin ser dueño/rights-holder/staff → rechazado antes de tocar Postgres.
- Reemplazar tags con una lista distinta borra los viejos e inserta los nuevos atómicamente (test de repository, confirmando que un fallo a mitad de camino no deja tags huérfanos — misma transacción).
- El broadcast de `settings/updated` llega a **todos** los ocupantes activos, incluido alguien que entró al room milisegundos antes del guardado (sin condición de carrera con `Join`).

---

## Parte 3 — Corrección de los bugs reales encontrados en `ROOM_INFO`

Ninguno de estos tres puntos depende de que exista `settings/save` — son correcciones al código YA implementado, que se pueden aplicar de forma aislada e inmediata:

### 3.1 `moderation(room)` — mapeo de columna equivocada

```go
// internal/realm/navigator/commands/info/packet.go (corregido)

func moderation(room roommodel.Room) outinfo.ModerationSettings {
    return outinfo.ModerationSettings{
        AllowMute: int32(room.ModerationMute), // antes: int32(room.ChatProtection) — bug
        AllowKick: int32(room.ModerationKick), // antes: hardcodeado a 0
        AllowBan:  int32(room.ModerationBan),  // antes: hardcodeado a 0
    }
}
```
Sin este fix, todo cliente que abre el panel de información de CUALQUIER sala ve datos de moderación incorrectos (una mezcla de la config de chat en el campo de mute, y kick/ban siempre en "solo dueño" sin importar la config real) — el bug ya existe en producción potencial hoy, este plan lo corrige como parte de terminar la Parte 3 de `REMAINING-ROOMS.md`, no como un efecto secundario incidental.

### 3.2 `CanMute` — reflejar al viewer real, no `false` fijo

```go
// internal/realm/navigator/commands/info/command.go (corregido)

func (handler Handler) sendRoomInfo(ctx context.Context, input Command, room roommodel.Room, viewerID int64, tags []string) error {
    canMute, err := handler.Moderation.CanExecute(ctx, room, viewerID, moderation.ActionMute)
    if err != nil {
        return err
    }
    packet, err := outinfo.Encode(outinfo.Params{
        // ...
        CanMute: canMute,
        // ...
    })
    // ...
}
```
`handler.Moderation` es el mismo `moderation.Service` de `plan/rooms/RIGHTS.md` — `Handler` gana ese campo, y `command.go` gana `viewerID` (ya resuelto por `navsession.Player`, que hoy se descarta después de validarlo — se empieza a propagar).

### 3.3 `AllInRoomMuted` — depende de la Parte 5 de este plan

Se deja preparado el campo, sin valor real hasta que la Parte 5 (toggle de "silenciar todo") exista — mismo criterio de "punto de enganche reservado, sin cambio de forma" ya usado en `plan/rooms/ENTRY.md`/`RIGHTS.md` para dependencias entre partes de un mismo plan.

### 3.4 Tests

- `moderation(room)` con cada combinación de `ModerationMute/Kick/Ban` produce el `ModerationSettings` esperado — regresión directa del bug encontrado.
- `CanMute` refleja `true` para un dueño con `ModerationOwnMute`, `false` para un visitante sin ningún permiso, `true` para staff con `ModerationAnyMute` en una sala ajena.

---

## Parte 4 — Wordfilter propio del room (distinto del filtro global de la Parte 2.2)

**No es lo mismo que el `ProfanityChecker` de la Parte 2.2**: ese valida el NOMBRE/DESCRIPCIÓN/TAGS de la sala al guardarla (contenido que el DUEÑO escribe sobre su propia sala). Este es una lista de palabras **adicionales** que el dueño quiere bloquear del **chat de los visitantes** dentro de su sala — un moderation tool, no una validación de settings. Confirmado como dos conceptos distintos en `REMAINING-ROOMS.md` Parte 3 ("Adicionalmente: filtro de palabras propio del room").

### 4.1 Esquema

```sql
create table room_word_filters (
    room_id bigint not null references rooms(id) on delete cascade,
    word text not null,
    created_at timestamptz not null default now(),
    primary key (room_id, word),
    constraint room_word_filters_word_length_chk check (char_length(word) between 1 and 32)
);
```

### 4.2 Servicio y comandos

```go
// internal/realm/room/wordfilter/service.go

type Manager interface {
    List(ctx context.Context, roomID int64) ([]string, error)
    Add(ctx context.Context, roomID int64, actorID int64, word string) error
    Remove(ctx context.Context, roomID int64, actorID int64, word string) error
    // Contains reports whether text contains any room-specific filtered word — el
    // futuro consumidor real de este servicio, el pipeline de chat.
    Contains(ctx context.Context, roomID int64, text string) (bool, error)
}
```

Autorización: misma fórmula de la Parte 1.1 (dueño/rights-holder con `SettingsOwnManage`, o staff con `SettingsAnyManage`) — pedir/modificar el wordfilter usa el mismo nodo que pedir/guardar settings, ya que ambos viven en el mismo diálogo de configuración según el research de Arcturus.

`internal/realm/room/commands/wordfilter/{request,modify}` — comandos nuevos, mismo patrón `command.Command`/`Handler[T]`.

### 4.3 Sin consumidor real todavía

`Contains` queda como `// TODO(chat): consultar esto en el pipeline de envío de mensajes una vez exista el realm de chat` — mismo criterio ya usado para `moderation.Service.IsMuted` en `plan/rooms/RIGHTS.md`. El servicio y la tabla existen desde este plan; el enganche real llega con el chat.

### 4.4 Tests

- Agregar/quitar una palabra requiere la misma autorización que settings.
- `List` refleja el estado actual tras agregar/quitar.
- `Contains` (aislado, sin esperar al chat) detecta coincidencias case-insensitive de palabras completas — no substring suelto (para no bloquear, por ejemplo, "clasico" por contener una palabra prohibida de 4 letras adentro).

---

## Parte 5 — Toggle "silenciar todo el chat de la sala" (mute-all, en memoria)

Confirmado por `REMAINING-ROOMS.md` Parte 3: en memoria, **no persistido**, dueño-only — mismo criterio que Arcturus, es un estado de sesión de la sala activa, no un dato durable de la fila `rooms`.

### 5.1 Estado en `roomlive.Room`

```go
// internal/realm/room/live/muteall.go

// SetMuteAll toggles whether all non-privileged chat is currently silenced.
func (room *Room) SetMuteAll(muted bool) {
    room.mutex.Lock()
    defer room.mutex.Unlock()
    room.muteAll = muted
}

// MuteAll reports whether the room is currently silencing all chat.
func (room *Room) MuteAll() bool {
    room.mutex.RLock()
    defer room.mutex.RUnlock()

    return room.muteAll
}
```

Un simple `bool` en el mismo `struct Room` ya protegido por `room.mutex` — no hace falta ninguna estructura lazy como `doorbell` (esto no tiene un costo de memoria por-entrada que valga la pena diferir, es un solo booleano).

### 5.2 Comando y packet

`internal/realm/room/commands/mute/toggle` — dueño-only (o `SettingsOwnManage`, mismo nodo que el resto de la configuración — activar/desactivar el silencio total vive en el mismo panel que el resto de settings según el research). Si el room no está activo (`roomlive.Registry.Find` no lo encuentra), no hay nada que mutear — responde éxito trivial sin crear un room activo solo para esto.

```go
// networking/outbound/room/mute/state/packet.go
// Params: Muted bool
```

Broadcast a todos los ocupantes activos tras cada toggle.

### 5.3 Wiring con `ROOM_INFO` (cierra la Parte 3.3)

```go
func (handler Handler) allInRoomMuted(roomID int64) bool {
    active, found := handler.Runtime.Find(roomID)
    if !found {
        return false
    }

    return active.MuteAll()
}
```

### 5.4 Tests

- Toggle solo lo puede accionar quien pasa la autorización de settings (2.1) — mismo test de tabla que el resto de esta serie de documentos.
- Activar el toggle en un room sin runtime activo no crea un room activo ni falla — responde éxito sin efecto real (no hay nadie a quien mutear).
- `ROOM_INFO` refleja `AllInRoomMuted` correctamente antes/después del toggle, para un room activo.
- El toggle **no persiste** — cerrar y reactivar el room (pasar por `Registry.Close`/`Registry.Activate`, ej. por quedar vacío y recargarse) resetea `MuteAll()` a `false`, confirmando que es estado de sesión, no dato durable (regresión explícita contra el bug de Arcturus que el mute POR USUARIO tenía y que `RIGHTS.md` ya decidió no replicar para ESE caso — acá, en cambio, la pérdida de estado al descargar el room es el comportamiento CORRECTO y buscado, porque el research confirma que el toggle en sí nunca fue pensado como durable).

---

## Parte 6 — Nodos de permiso nuevos (extiende `internal/realm/room/permissions.go`, ya existente)

```go
// internal/realm/room/permissions.go (agregado a los nodos ya existentes de ENTRY.md/RIGHTS.md)

var (
    // SettingsOwnManage allows editing the settings of a room you own or hold rights in.
    SettingsOwnManage = permission.RegisterNode("room.settings.own.manage", "")
    // SettingsAnyManage allows editing the settings of any room (staff).
    SettingsAnyManage = permission.RegisterNode("room.settings.any.manage", "")
)
```

Estos dos nodos ya estaban previstos en `plan/PERMISSIONS.md` Parte 3.1 — este plan es el que finalmente los declara y los consume (settings, wordfilter, y el toggle de mute-all, los tres bajo el mismo par de nodos, ya que los tres viven en el mismo panel de configuración según el research).

---

## Parte 7 — Sincronización completa: qué dato vive dónde, y quién lo refresca

Tabla resumen explícita, porque esto fue lo que se pidió puntualmente ("que se sincronice todo"):

| Dato | Fuente de verdad | Quién lo lee | Cómo se entera de un cambio |
| --- | --- | --- | --- |
| Nombre/descripción/tags/door mode/trade mode/moderación | columna de `rooms`/`room_tags` | Navegador (búsqueda), `ROOM_INFO`, ocupantes activos | Navegador y `ROOM_INFO` **leen en vivo** de Postgres cada vez que se piden (Parte 0) — no hace falta empujarles nada. Los ocupantes activos SÍ necesitan un empuje en tiempo real, porque no vuelven a pedir nada por su cuenta mientras están adentro: lo reciben vía el broadcast `settings/updated` (Parte 2.6). |
| Thickness/paredes ocultas | columna de `rooms` | Solo renderizado dentro del room (no aparece en navegador/`ROOM_INFO`) | Broadcast `thickness/updated` a ocupantes activos (2.6) — nadie más necesita este dato. |
| Chat mode/weight/speed/distance/protection | columna de `rooms` | `ROOM_INFO` (lectura), pipeline de chat futuro (aplicación real), ocupantes activos | `ROOM_INFO` ya lo lee en vivo (Parte 0, `chat(room)` ya está bien mapeado, a diferencia de `moderation(room)`). Ocupantes activos vía `chatsettings/updated` (2.6). |
| Moderación (quién puede mutear/kickear/banear) | columna de `rooms` (`ModerationMute/Kick/Ban`) | `ROOM_INFO` (corregido, Parte 3.1), `moderation.Service.Authorize` (ya implementado en `RIGHTS.md`, lee la fila de room directo en cada chequeo, no cachea) | Nada que sincronizar de más — cada chequeo de autorización ya relee la fila actual; `ROOM_INFO` se corrige en la Parte 3. |
| Contraseña (hash) | columna `password_hash` | `entry.Service` (gating de entrada), nunca se expone al cliente | `settings/current` solo expone `HasPassword bool` (1.2) — el hash en sí nunca sale del servidor. |
| Wordfilter propio | `room_word_filters` | Futuro pipeline de chat | Sin consumidor real todavía (4.3) — nada que sincronizar hasta que exista. |
| Mute-all | memoria (`roomlive.Room.muteAll`) | `ROOM_INFO` (`AllInRoomMuted`), futuro pipeline de chat | Broadcast `mute/state` a ocupantes activos (5.2); `ROOM_INFO` lo lee en vivo del runtime activo (5.3) — nunca se persiste, así que no hay "desincronización" posible entre reinicios: simplemente vuelve a `false`. |

**Regla general** (la que unifica toda la tabla): todo lo que un cliente puede *pedir bajo demanda* (navegador, `ROOM_INFO`) se resuelve leyendo el estado actual en el momento de la consulta, sin ningún cache — la sincronización ahí es gratis por construcción. Todo lo que un cliente *no vuelve a pedir por su cuenta* mientras está dentro de la sala (porque ya la tiene "abierta") necesita un broadcast explícito cuando cambia — eso es exactamente lo que la Parte 2.6/5.2 cubren, y nada más necesita mecanismo de sync porque no hay ningún otro estado con esa característica en el alcance de este plan.

---

## Parte 8 — Protocolo completo (resumen)

| Dirección | Paquete | Contenido | Header |
| --- | --- | --- | --- |
| Inbound | `room/settings/request` | sin campos | TBD |
| Inbound | `room/settings/save` | todos los campos de `UpdateParams` (2.1) | TBD |
| Inbound | `room/wordfilter/request` | sin campos | TBD |
| Inbound | `room/wordfilter/modify` | `word string`, `add bool` | TBD |
| Inbound | `room/mute/toggle` | sin campos (toggle, no un valor explícito) | TBD |
| Outbound | `room/settings/current` | `Params` completo (1.2) | TBD |
| Outbound | `room/settings/error` | código de error de campo (`ErrPasswordRequired`, conflicto de versión, tag reservado, etc.) | TBD |
| Outbound | `room/settings/saved` | confirmación + `Room` actualizado (2.6) | TBD |
| Outbound | `room/settings/updated` | broadcast, ver 2.6 | TBD |
| Outbound | `room/thickness/updated` | broadcast, ver 2.6 | TBD |
| Outbound | `room/chatsettings/updated` | broadcast, ver 2.6 | TBD |
| Outbound | `room/wordfilter/list` | lista actual | TBD |
| Outbound | `room/mute/state` | `Muted bool`, broadcast (5.2) | TBD |
| Outbound (ya existe, corregido) | `navigator/roominfo` (`ROOM_INFO`, **687**) | mismo shape, `moderation`/`CanMute`/`AllInRoomMuted` corregidos (Parte 3) | 687 |

Headers nuevos `TBD — a confirmar contra Nitro real`, mismo criterio que el resto de esta serie de planes; `ROOM_INFO` ya tiene su header real confirmado (687) porque ya está implementado.

---

## Parte 9 — Hot paths, allocations, benchmarks

- **`ROOM_INFO`/navegador nunca tocan el hot path de movimiento/entrada** — son requests bajo demanda del jugador (abrir el panel de info, buscar en el navegador), no algo que corra en cada tick ni en cada entrada a un room. El fix de la Parte 3 no cambia el perfil de costo, solo corrige qué columnas se leen (ya estaban siendo leídas, solo mal mapeadas).
- **`settings/save` no es un hot path** — es una acción explícita y poco frecuente del dueño; el `update` con `version = version + 1` + el replace completo de tags corren en una sola transacción, sin necesitar ninguna optimización especial más allá de los índices ya existentes.
- **El broadcast de los 3 packets (2.6) reusa `broadcast.RoomPacket`**, que ya tolera fallos de envío individuales sin abortar el resto (mismo mecanismo ya usado por moderación/rights) — un ocupante con la conexión cayendo en simultáneo no rompe el guardado de settings para nadie más.
- **`MuteAll()`/`SetMuteAll()` son lecturas/escrituras triviales de un `bool`** bajo un mutex que ya existe — costo insignificante, no justifica ningún benchmark dedicado por sí solo.
- **`CanExecute`/`Authorize` de `moderation.Service`, reusado en `ROOM_INFO` (Parte 3.2)**: cada apertura del panel de información ahora dispara un chequeo de permisos adicional (antes no lo hacía, porque `CanMute` estaba hardcodeado) — mismo costo ya cubierto por `BenchmarkAuthorizeModerationAction` en `plan/rooms/RIGHTS.md`, no se duplica acá; se nota como el único cambio de costo real introducido por este plan en una ruta que antes no tocaba permisos.

Benchmark nuevo (mismo patrón ya establecido):
```go
// internal/realm/room/settings/benchmark_test.go

// BenchmarkUpdateValidation measures the service-side validation cost (ProfanityChecker
// nil, tag policy check, password-required check) against fakes, no Postgres I/O.
func BenchmarkUpdateValidation(b *testing.B) { ... }
```

---

## Parte 10 — Testing (resumen transversal, detalle por parte ya cubierto arriba)

- Fakes de repository para `settings.Service`/`wordfilter.Service` — mismo patrón ya establecido en el resto del realm `room`.
- Tests de repository reales contra Postgres de test para `Update` (conflicto de versión, replace de tags atómico) y `room_word_filters` — mismo patrón que `internal/realm/room/repository/repository_test.go`.
- Test de regresión explícito sobre los dos bugs de la Parte 3 (`moderation(room)`, `CanMute`) — falla si alguien los reintroduce.
- Test de integración: guardar settings con un `ProfanityChecker` real de prueba (fake que marca ciertas palabras) confirma que la validación efectivamente bloquea, no solo que la interfaz existe.

---

## Parte 11 — Milestones de implementación

1. **C1 — Corrección de los bugs reales en `ROOM_INFO`** (Parte 3): arreglar `moderation(room)`, propagar `viewerID` y wirear `CanMute` contra `moderation.Service`. Sin dependencias — se puede hacer hoy mismo, es una corrección aislada sobre código ya en producción potencial.
2. **C2 — Nodos de permiso** (Parte 6): `SettingsOwnManage`/`SettingsAnyManage` en `internal/realm/room/permissions.go`. Puede correr en paralelo a C1.
3. **C3 — `roomservice.Manager.Update` + validación + persistencia** (Parte 2.1-2.5): depende de C2 (la autorización de `Update` necesita los nodos).
4. **C4 — Comandos y packets de `settings/{request,save}`** (Parte 1, 2.6, 8): depende de C3.
5. **C5 — Wordfilter propio del room** (Parte 4): depende de C2 (misma autorización que settings). Puede correr en paralelo a C3/C4.
6. **C6 — Toggle de mute-all** (Parte 5): depende de C2. Puede correr en paralelo a C3/C4/C5 — cierra el `AllInRoomMuted` de la Parte 3.3 al converger con C1.
7. **C7 — `ProfanityChecker` real** (Parte 2.2): fuera del alcance inmediato de este plan si ningún otro realm necesita un filtro de contenido antes — hasta entonces, `service.profanity == nil` degrada correctamente (documentado como milestone futuro, no bloqueante).

### Milestones futuros confirmados (fuera de este documento, no descartados)

- **`ProfanityChecker` compartido** (C7) — se construye cuando exista una necesidad real de validar texto de usuario en más de un lugar del proyecto (settings de room es hoy el único candidato confirmado); hasta entonces, la degradación a `nil` es el comportamiento correcto, no una deuda técnica.
- **Límite de `MaxUsers` atado a la capacidad real del layout** (2.2) — depende de que `layout.Layout`/`room/world/grid` expongan un techo de capacidad utilizable; se confirma contra el estado real de esos paquetes al implementar.
- **`wordfilter.Contains`/`IsMuted` (RIGHTS.md) enganchados al pipeline de chat** — ambos ya anotados con `// TODO(chat)` en sus respectivos planes; se resuelven juntos el día que el realm de chat exista, no antes.
