# Plan: Bots + remanentes de Room/Furniture/Room Entities

Dos cosas en un documento, deliberadamente: (1) el sistema de **bots** completo, diseñado al máximo nivel de detalle posible — dominio, IA de ciclo, comportamientos concretos (genérico, mayordomo/bartender, registro de visitas), extensibilidad para comportamientos futuros vía SDK (incluyendo bots controlados por modelos de IA, sin implementarlo aquí); y (2) todo lo demás que `plan/STATUS.md` dejó suelto en los realms `Room`/`Furniture & Items`/`Room Entities` que **no** está explícitamente diferido en otro documento (Pets, Wired, furniture de grupos/juegos quedan fuera, con su propia justificación ya escrita en otros planes) y que no es una expresión/efecto de avatar (eso es `plan/EXPRESSIONS-EFFECTS.md`, incluidos ya los furniture que otorgan efectos).

Investigado contra `legacy/Arcturus-Community` (`habbohotel/bots/{Bot,BotManager,ButlerBot,VisitorBot}.java`, `messages/{incoming,outgoing}/rooms/bots/**`, `plugin/events/bots/**`, `threading/runnables/{BotFollowHabbo,RoomUnitGiveHanditem,RoomUnitWalkToRoomUnit}.java`, y los `habbohotel/items/interactions/{InteractionPostIt,InteractionMannequin,InteractionBackgroundToner,InteractionLoveLock,InteractionJukeBox}.java` para la Parte 2) y `legacy/pixel-protocol` (headers reales confirmados para cada packet citado).

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding)

| Ya existe | Dónde | Cómo lo usa este plan |
| --- | --- | --- |
| Scheduler de proceso de 500ms (`Room.Schedule`/`ScheduledTask`) | `plan/interactions/MEGAPLAN-ESSENTIAL.md` Parte 1 | El ciclo de IA de un bot (caminar/hablar) es, en la referencia, literalmente un tick recurrente — se reusa este primitivo tal cual, no se construye un segundo scheduler. |
| `worldunit.Unit` (posición, rotación, statuses genéricos, `ControlKind`) | `internal/realm/room/world/unit` | Un bot es, en el mundo físico de la sala, **otro `Unit` más** — mismo pathfinding, mismo sistema de statuses (incluyendo los nuevos `StatusDance`/`StatusSign` de `plan/EXPRESSIONS-EFFECTS.md`), mismo hook `walkedon` ya generalizado "no por tipo" (`plan/interactions/TOGGLE-GATE.md` Parte 7). |
| `furnitureaccess.CanManage` (derechos de sala o staff) | `internal/realm/furniture/access` | Reusado directo para "¿puede este jugador colocar/recoger este bot?" — mismo criterio que colocar furniture. |
| `chat/filter`/wordfilter global | `plan/CHAT.md` | Reusado para nombre/motto/líneas de chat de un bot — un bot habla en la sala, así que sus textos pasan por el mismo filtro que cualquier jugador. |
| `essential.Source` (RNG inyectable) | `internal/realm/furniture/interactions/essential/service.go` | No aplica directo a bots (los bots no tienen aleatoriedad de premio), pero el patrón de "interfaz inyectable para poder testear determinísticamente" se reusa para el temporizador de caminata aleatoria (Parte 4.3). |
| `RoomUnitGiveHanditem`/`HandItemHolder` (primitivo de mano) | `plan/interactions/MEGAPLAN-ESSENTIAL.md` Parte 1.5 | El comportamiento Bartender (Parte 5.2) es, en la referencia, el caso de uso real que motivó ese primitivo — se reusa exactamente, no se construye un segundo sistema de "objeto en la mano". |
| `room.access.events.left`, `player.events.disconnected` | ya reusados en `plan/MESSENGER.md`/`plan/MARKET-TRADING.md` | El seguimiento de un bot a un jugador (Parte 4.4) debe cancelarse con el mismo criterio — si el jugador seguido sale de la sala o se desconecta, el bot deja de seguir. |
| `permission.RegisterNode` | `internal/permission` | Nodos ya establecidos como patrón — este plan agrega los propios de bots (Parte 8) sin inventar un mecanismo nuevo. |
| Ningún realm de `bots` hoy | confirmado, `find internal/realm -iname "*bot*"` vacío | Greenfield total — este documento es el primero en tocar el tema. |
| `plan/rooms/BUNDLE.md` Parte 8 (diferido) | Room Bundles necesitaba clonar bots de una plantilla | Este plan **cierra** esa dependencia — una vez que el modelo de bot exista, `BUNDLE.md` puede implementar `CloneRoomBots` con el mismo patrón que ya usa para furniture. |
| `plan/interactions/MEGAPLAN-ESSENTIAL.md` Parte 7 (`Cannon`) | Reusa `leavecmd.Handler` para expulsión no discrecional | Mencionado aquí solo como precedente de "un bot puede disparar acciones de sala sin pasar por autorización discrecional de jugador" — no se reusa directo, se anota el paralelismo. |

---

# PARTE A — BOTS (diseño exhaustivo)

## Parte 1 — Investigación exhaustiva (Arcturus)

### 1.1 `Bot.java` — el modelo base, campo por campo

```java
chatLines: ArrayList<String>       // líneas configuradas, cicladas secuencial o al azar
name, motto, figure, gender        // idénticos en forma a un HabboInfo/perfil de jugador
ownerId, ownerName                 // dueño actual (cambia al recoger/regalar)
room, roomUnit                     // referencias transitorias, null en inventario
chatAuto, chatRandom, chatDelay     // configuración de charla automática
chatTimeOut, chatTimestamp, lastChatIndex  // estado de ciclo de charla
bubble                             // estilo de burbuja de chat — MISMO sistema que CHAT.md ya construyó
type                                // discriminador de comportamiento ("generic", "bartender", "visitor_log", ...)
effect                             // efecto de avatar aplicado al bot — MISMO sistema de plan/EXPRESSIONS-EFFECTS.md
canWalk                            // si el bot puede moverse libremente
followingHabboId                   // 0 = no sigue a nadie
needsUpdate                        // flag de escritura diferida (mismo patrón lazy-write que CatalogItem/MarketPlaceOffer)
```
`Bot implements Runnable` — el mismo patrón de "flush diferido a BD" ya visto en el catálogo y el Marketplace (`run()` solo escribe si `needsUpdate`). Se replica igual: un `needsUpdate` en memoria, un flush asíncrono, no una escritura por cada mutación de campo.

### 1.2 El ciclo de IA (`cycle(allowBotsWalk)`) — la parte más importante de todo este documento

Llamado periódicamente (confirmado: mismo patrón de tick que `MEGAPLAN-ESSENTIAL.md` ya generalizó). Dos sub-comportamientos independientes, ambos condicionados a que el bot esté efectivamente en una sala:

**(a) Caminata aleatoria**, si `canWalk` y no está siguiendo a nadie (`followingHabboId==0`) y no está ya caminando y el timeout de espera venció:
```java
goal = BOT_LIMIT_WALKING_DISTANCE
    ? room.getRandomWalkableTilesAround(unit, currentTile, BOT_WALKING_DISTANCE_RADIUS) // radio 5, config real
    : room.getRandomWalkableTile();                                                      // toda la sala
unit.setGoalLocation(goal);
timeout = max(5, random(0,38)) segundos hasta el próximo intento              // fórmula real: random(20)*2, piso 5 si <10
```
**(b) Charla automática cíclica**, si hay líneas configuradas y `chatAuto` y venció el `chatTimeOut`:
```java
mensaje = chatLines[lastChatIndex]
    .replace("%owner%", room.ownerName)
    .replace("%item_count%", room.itemCount())
    .replace("%name%", bot.name)
    .replace("%roomname%", room.name)
    .replace("%user_count%", room.userCount());
// intenta que Wired intercepte el mensaje (WiredTriggerType.SAY_SOMETHING) — si Wired lo maneja, el bot NO habla por su cuenta
if (!Wired.handle(SAY_SOMETHING, ...)) bot.talk(mensaje);
lastChatIndex = chatRandom ? random(chatLines.size()) : (lastChatIndex+1) % chatLines.size();
chatTimeOut = now + chatDelay;
```
**Hallazgo real importante — punto de integración con Wired, hoy diferido**: cada vez que el bot va a decir algo (ciclo automático, saludo de colocación, o respuesta de comportamiento), Arcturus primero le da la oportunidad a un wired de tipo `SAY_SOMETHING` de interceptar y reemplazar esa línea — si el wired la maneja, el bot **no** habla por su cuenta. Como Wired está diferido (`plan/STATUS.md`), este plan **no** implementa la intercepción — pero diseña el punto de extensión explícitamente (Parte 6.3) para que conectar Wired el día de mañana sea agregar una función, no reabrir este documento.

### 1.3 Variables de plantilla en el chat — tabla completa confirmada

| Variable | Valor |
| --- | --- |
| `%owner%` | Nombre del dueño de la sala |
| `%item_count%` | Cantidad de muebles en la sala |
| `%name%` | Nombre del propio bot |
| `%roomname%` | Nombre de la sala |
| `%user_count%` | Cantidad de jugadores presentes |

### 1.4 Chat con tres alcances, cada uno con su propio evento de plugin cancelable

```java
talk(message)    → BotTalkEvent    → RoomUserTalkComposer (burbuja normal, todos en la sala)
shout(message)   → BotShoutEvent   → RoomUserShoutComposer (burbuja de grito, todos en la sala)
whisper(message, habbo) → BotWhisperEvent → RoomUserWhisperComposer (solo al destinatario)
```
**Detalle real curioso**: si el mensaje es exactamente `"o/"` o `"_o/"`, se dispara además un gesto de saludo (`RoomUserActionComposer(WAVE)`) sincronizado con el mensaje — un bot que "dice adiós con la mano" literalmente agita la mano. Se replica tal cual, reusando el mecanismo de gesto ya diseñado en `plan/EXPRESSIONS-EFFECTS.md` Parte 2.

### 1.5 Ciclo de vida: colocar y recoger

**Colocar** (`BotManager.placeBot`, orden real de validación):
1. Evento de plugin cancelable (`BotPlacedEvent`).
2. Autorización: dueño de la sala, o `ACC_ANYROOMOWNER`, o `ACC_PLACEFURNI`.
3. Tope de bots por sala (`Room.MAXIMUM_BOTS`), bypasseable por `ACC_UNLIMITED_BOTS`.
4. Tile de destino: no debe haber un jugador parado ahí, debe ser caminable o sit/lay; no debe haber ya otro bot ahí (chequeo separado del de jugadores — dos bots nunca comparten tile, aunque un bot y un jugador temporalmente sí durante el frame de colocación).
5. Crea el `RoomUnit` (tipo `BOT`, rotación sur, altura resuelta del tile), lo agrega a la sala, lo saca del inventario del dueño, confirma con un composer de "bot colocado".
6. `onPlace`: aplica el efecto configurado del bot (`plan/EXPRESSIONS-EFFECTS.md`) y dice un saludo aleatorio de una lista configurable (`PLACEMENT_MESSAGES`).
7. **Dispara `onWalkOn` del mueble debajo del tile de colocación** — un bot activa triggers de piso exactamente igual que un jugador (confirma que el hook `walkedon` ya generalizado en `plan/interactions/TOGGLE-GATE.md` debe aceptar unidades de bot sin distinción).

**Recoger** (`BotManager.pickUpBot`):
1. Evento de plugin cancelable (`BotPickUpEvent`).
2. Autorización: dueño del bot, o `ACC_ANYROOMOWNER`.
3. Tope de inventario de bots del receptor (`MAXIMUM_BOT_INVENTORY_SIZE=25`), bypasseable por `ACC_UNLIMITED_BOTS`.
4. Detiene cualquier seguimiento activo, reasigna el dueño, lo quita de la sala, lo agrega al inventario del receptor.

**Eliminar** (`BotManager.deleteBot`): borrado permanente (no soft-delete, no recuperable) — una acción explícitamente distinta de "recoger" (que sí es recuperable, vuelve al inventario).

### 1.6 Seguimiento (`startFollowingHabbo`/`BotFollowHabbo`)

Un runnable auto-reprogramable cada 500ms (mismo cadence que el `ScheduledTask` de 500ms ya generalizado en este proyecto):
```java
objetivo = tile detrás del jugador seguido (rotación del jugador + 180°, "+4 % 8")
si el objetivo cae fuera de la sala: usa el tile justo delante en su lugar
si la distancia entre bot y jugador < 2 y no se había marcado "alcanzado": dispara un hook (hoy Wired, diferido)
si el objetivo es válido: fija el goal del bot y se reprograma en 500ms
```
Se cancela solo cuando `followingHabboId` deja de coincidir con el jugador (asignado a 0 por `stopFollowingHabbo`, llamado al recoger el bot o — decisión de este plan, no confirmada byte a byte en la referencia — cuando el jugador seguido sale de la sala/se desconecta, Parte 4.4).

### 1.7 Configuración (`BotSaveSettingsEvent`, legacy multiplexado; protocolo real ya lo separa)

Arcturus multiplexa todo en un solo packet con un `settingId` (1=look copiado del dueño, 2=chat con validación anti-XSS-loop + wordfilter + longitud, 3=toggle caminar, 4=ciclar tipo de baile, 5=nombre con wordfilter, 9=motto). **El protocolo real moderno ya lo separa en packets propios**, confirmado:

| Header | Nombre | Reemplaza |
| --- | --- | --- |
| c2s `1986` | `BOT_CONFIGURATION` (`botId`) | apertura del panel — responde `1618` |
| s2c `1618` | `BOT_COMMAND_CONFIGURATION` | panel de configuración actual |
| c2s `2624` | `BOT_SKILL_SAVE` (`botId`, `skillId`, `data` string JSON/texto libre) | el `settingId`-switch legacy completo, generalizado |

Se diseña sobre el shape moderno (`skillId`+`data` genérico), no sobre el switch legacy — más extensible (agregar una skill nueva no requiere un valor mágico más en un switch, solo un nuevo `skillId` documentado y un parser de `data`).

### 1.8 Validaciones/límites reales confirmados

```
MINIMUM_CHAT_SPEED = 7s          MAXIMUM_CHAT_SPEED = 604800s (7 días)
MAXIMUM_CHAT_LENGTH = 120         MAXIMUM_NAME_LENGTH = 15
MAXIMUM_BOT_INVENTORY_SIZE = 25   Room.MAXIMUM_BOTS (tope por sala, valor no confirmado en los archivos leídos)
BOT_WALKING_DISTANCE_RADIUS = 5   BOT_LIMIT_WALKING_DISTANCE = true (config)
```
Anti-XSS real en el guardado de chat: un ciclo que aplica `Jsoup.parse(s).text()` repetidamente hasta que el resultado se estabiliza (protección contra HTML doblemente codificado), con un tope de 5 iteraciones antes de vaciar todo el chat del bot como medida defensiva — se replica el criterio (sanitizar hasta punto fijo, con un tope), usando la librería de sanitización que Pixels ya use en el resto del proyecto para texto de usuario, no Jsoup específicamente.

### 1.9 Los 10 eventos de plugin reales — mapeo a `pkg/bus`

```
BotPlacedEvent, BotPickUpEvent            → bot.placed, bot.picked_up
BotSavedLookEvent, BotSavedNameEvent, BotSavedChatEvent → bot.settings.look_saved / name_saved / chat_saved
BotTalkEvent, BotShoutEvent, BotWhisperEvent → bot.talked / bot.shouted / bot.whispered
BotServerItemEvent                          → bot.serve_item.requested (Parte 5.2, Bartender)
BotEvent (base)                             → no se replica como evento propio, es la clase base abstracta
```

### 1.10 Comportamientos reales concretos — los tres tipos registrados

Arcturus registra exactamente tres bajo un registro reflectivo `Map<String, Class<? extends Bot>>` (`BotManager.botDefenitions`) — **este registro es, en sí mismo, la prueba de que el patrón de extensibilidad ya existía en la referencia**; este plan lo formaliza como el punto de inserción de comportamientos personalizados (Parte 6).

1. **`"generic"` → `Bot` (clase base)** — el bot decorativo estándar: camina, charla en ciclo, sin ningún comportamiento especial en `onUserSay`/`onPickUp`.
2. **`"bartender"` → `ButlerBot`** — Parte 5.2.
3. **`"visitor_log"` → `VisitorBot`** — Parte 5.3.

---

## Parte 2 — Decisiones de diseño

1. **Un bot es un `RoomUnit` más**, no un sistema paralelo de movimiento/pathfinding — reusa exactamente el mismo motor físico que un jugador (Parte 0). La única diferencia estructural es de dónde vienen sus decisiones (jugador = input de red; bot = ciclo de IA local).
2. **El ciclo de IA (caminar+hablar) es un `Behavior` intercambiable**, no una función hardcodeada — esto es lo que permite (a) los tres comportamientos reales de Arcturus, y (b) dejar la puerta abierta a comportamientos futuros (incluida IA real) sin tocar el núcleo (Parte 6).
3. **El chat de un bot pasa por el mismo wordfilter que un jugador** — un bot no es una excepción de moderación solo porque lo controla el servidor; su texto lo configuró un jugador (o, a futuro, un modelo externo) y debe cumplir las mismas reglas.
4. **Se diseña sobre el shape moderno de configuración** (`BOT_SKILL_SAVE` con `skillId`+`data`), no sobre el switch legacy de Arcturus — más fácil de extender con nuevas skills sin tocar un switch central.
5. **La integración con Wired se deja como un punto de extensión explícito, no implementado** (Parte 6.3) — consistente con que Wired está diferido en todo el proyecto (`plan/STATUS.md`), pero sin bloquear el diseño de bots a que Wired exista primero.

---

## Parte 3 — Esquema

```sql
--liquibase formatted sql
--changeset pixels:pixels-bot-0001-create-bots
create table bots (
    id bigint generated always as identity primary key,
    owner_player_id bigint not null,
    room_id bigint null references rooms(id), -- null = en inventario
    behavior_type text not null default 'generic',
    name text not null,
    motto text not null default '',
    figure text not null,
    gender smallint not null, -- espejo de playermodel.Gender
    x integer null, y integer null, z double precision null, rotation smallint null,
    can_walk boolean not null default true,
    dance_type smallint not null default 0,
    chat_auto boolean not null default false,
    chat_random boolean not null default false,
    chat_delay_seconds integer not null default 10,
    bubble_style integer not null default 0,
    effect_id integer null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1,
    constraint bots_name_length_chk check (char_length(name) between 1 and 15),
    constraint bots_chat_delay_chk check (chat_delay_seconds between 7 and 604800)
);
create index bots_owner_idx on bots (owner_player_id) where room_id is null;
create index bots_room_idx on bots (room_id) where room_id is not null;

create table bot_chat_lines (
    bot_id bigint not null references bots(id) on delete cascade,
    order_num integer not null,
    line text not null,
    primary key (bot_id, order_num)
);
--rollback drop table if exists bot_chat_lines; drop table if exists bots;
```
`behavior_type` es el discriminador que resuelve qué `Behavior` (Parte 6) gobierna el ciclo de IA de esta fila — mismo rol que la columna `type` real, pero explícitamente documentado como el punto de extensión, no un detalle incidental.

```sql
--changeset pixels:pixels-bot-0002-create-butler-serves
create table bot_serve_items (
    id bigint generated always as identity primary key,
    keyword text not null,
    definition_id bigint not null,
    constraint bot_serve_items_keyword_uidx unique (keyword)
);
```
Simplificación deliberada sobre la referencia: Arcturus permite un **conjunto** de keywords sinónimas por ítem (`THashSet<String>` como clave) — este plan usa una fila por keyword individual (varias filas pueden apuntar al mismo `definition_id`) porque es más simple de administrar vía CRUD HTTP sin perder ninguna capacidad real.

---

## Parte 4 — Servicio core

### 4.1 Colocar / recoger

```go
// internal/realm/bot/service/lifecycle.go
func (service *Service) Place(ctx context.Context, params PlaceParams) (Bot, error) {
    bot, found, err := service.store.FindInventoryBot(ctx, params.BotID, params.OwnerPlayerID)
    if err != nil || !found { return Bot{}, ErrBotNotFound }
    active, found := service.rooms.Find(params.RoomID)
    if !found { return Bot{}, ErrRoomNotFound }
    allowed, err := furnitureaccess.CanManage(ctx, service.permissions, active, params.ActorPlayerID) // mismo criterio que furniture
    if err != nil { return Bot{}, err }
    if !allowed { return Bot{}, ErrNoRights }
    if service.countRoomBots(active) >= service.config.MaxBotsPerRoom && !service.permissions.Has(ctx, params.ActorPlayerID, BotUnlimited) {
        return Bot{}, ErrRoomBotLimit
    }
    if active.HasOccupantAt(params.Point) || active.HasBotAt(params.Point) || !placeable(active, params.Point) {
        return Bot{}, ErrTileNotFree
    }
    unit := worldunit.NewBot(params.Point, worldunit.RotationSouth)
    active.AddBot(bot.ID, unit)
    // dispara walkedon del fixture debajo, igual que un jugador (Parte 1.5 paso 7)
    service.behaviors.For(bot.BehaviorType).OnPlace(ctx, &bot, active) // efecto + saludo, Parte 6
    return bot, service.store.MarkPlaced(ctx, bot.ID, params.RoomID, params.Point)
}

func (service *Service) Pickup(ctx context.Context, params PickupParams) (Bot, error) {
    // simétrico: autorización, tope de inventario (25, configurable), detiene seguimiento, reasigna dueño
}
```

### 4.2 Configuración (`BOT_SKILL_SAVE`)

```go
// internal/realm/bot/service/skills.go
func (service *Service) SaveSkill(ctx context.Context, botID int64, actorPlayerID int64, skillID int32, data string) error {
    // autorización igual que Place/Pickup
    switch skillID {
    case SkillLook:      return service.saveLook(ctx, botID, data)      // copia look+gender+effect del actor
    case SkillChat:      return service.saveChat(ctx, botID, data)      // parsea, sanitiza hasta punto fijo, filtra, trunca a 120
    case SkillWalk:      return service.toggleWalk(ctx, botID)
    case SkillDance:     return service.cycleDance(ctx, botID)
    case SkillName:      return service.saveName(ctx, botID, data)      // filtra, max 15, sin '<'/'>'
    case SkillMotto:     return service.saveMotto(ctx, botID, data)
    default:             return service.behaviors.For(bot.BehaviorType).SaveCustomSkill(ctx, botID, skillID, data) // Parte 6, extensión
    }
}
```
`SkillLook` copia `figure`/`gender`/`effect` del jugador que configura — **"hacé que el bot se vea como yo"**, confirmado real, no una decisión de diseño de este plan.

### 4.3 Ciclo de IA — reusa el scheduler de 500ms

```go
// internal/realm/bot/runtime/cycle.go
func (runtime *Runtime) Tick(ctx context.Context, active *roomlive.Room) {
    for _, bot := range active.Bots() {
        behavior := runtime.behaviors.For(bot.BehaviorType)
        behavior.OnCycle(ctx, bot, active) // Parte 6 — caminar/hablar es responsabilidad del Behavior, no de este loop
    }
}
```
El loop de sala solo itera y delega — toda la lógica de "cuándo caminar"/"cuándo hablar" vive dentro de cada `Behavior.OnCycle`, para que un comportamiento personalizado (Parte 6) pueda reemplazar sin tocar el runtime.

```go
// genericBehavior.OnCycle — el comportamiento "generic" reproduce el algoritmo real de la Parte 1.2
func (generic) OnCycle(ctx context.Context, bot *Bot, active *roomlive.Room) error {
    if bot.CanWalk && bot.FollowingPlayerID == 0 && !bot.Unit.IsWalking() && bot.WalkTimeout.Before(now()) {
        goal := active.RandomWalkableTileAround(bot.Unit.Point(), BotWalkRadius) // radio 5, mismo valor real
        bot.Unit.SetGoal(goal)
        bot.WalkTimeout = now().Add(randomDuration(5, 38, source)) // misma fórmula real, con Source inyectable para tests
    }
    if bot.ChatAuto && len(bot.ChatLines) > 0 && bot.ChatTimeout.Before(now()) {
        message := substituteTemplate(bot.ChatLines[bot.ChatIndex], active) // %owner%/%item_count%/%name%/%roomname%/%user_count%
        if !runtime.wiredHook(SayHomething, bot, active, message) { // Parte 6.3 — hoy siempre false (no-op), punto de extensión
            bot.Talk(ctx, message)
        }
        bot.AdvanceChatIndex()
        bot.ChatTimeout = now().Add(bot.ChatDelay)
    }
    return nil
}
```

### 4.4 Seguimiento

```go
// internal/realm/bot/runtime/follow.go — mismo cadence de 500ms que BotFollowHabbo real
func (runtime *Runtime) followTick(ctx context.Context, bot *Bot, active *roomlive.Room) {
    target, found := active.OccupantByPlayerID(bot.FollowingPlayerID)
    if !found { bot.StopFollowing(); return } // decisión de este plan: jugador ya no está en la sala → deja de seguir
    goal := tileBehind(active, target.Unit) // rotación+180°, con fallback al tile de enfrente si cae fuera de la sala
    bot.Unit.SetGoal(goal)
}
```
Enganchado también a `room.access.events.left`/`player.events.disconnected` (Parte 0) para detener el seguimiento inmediatamente, no solo en el próximo tick de 500ms.

### Tests
- Colocar en un tile ocupado por un jugador o por otro bot rechaza, con el mismo criterio para ambos casos por separado.
- El tope de bots por sala rechaza salvo con el permiso de bypass.
- El tope de inventario (25) rechaza recoger un bot adicional, salvo bypass.
- `SkillChat` sanitiza HTML hasta punto fijo con tope de 5 iteraciones, vaciando el chat si no converge — mismo criterio defensivo real.
- El ciclo de caminata nunca se dispara mientras `FollowingPlayerID != 0`.
- Seguir a un jugador que sale de la sala detiene el seguimiento inmediatamente, no en el próximo tick.

---

## Parte 5 — Comportamientos concretos

### 5.1 `generic` — decorativo estándar

Ya completamente descrito en 4.3 — camina al azar, charla en ciclo, sin reacciones a eventos de sala. Es el comportamiento por defecto y el que sirve de referencia para cualquier comportamiento nuevo.

### 5.2 `bartender` — servicio por palabra clave (el ejemplo más rico de la referencia)

```go
// internal/realm/bot/behavior/bartender.go
func (bartender) OnUserSay(ctx context.Context, bot *Bot, message ChatMessage, active *roomlive.Room) error {
    if bot.Unit.IsWalking() { return nil }
    distance := bot.Unit.Point().Distance(message.Actor.Point())
    if distance > bartender.config.CommandDistance { return nil } // "hotel.bot.butler.commanddistance" real
    for _, entry := range bartender.serveItems.MatchWholeWord(message.Text) { // \bkeyword\b, mismo criterio real
        // evento cancelable equivalente a BotServerItemEvent
        if bot.Unit.CanWalk() {
            bot.LookAt(message.Actor)
            if distance > bartender.config.ReachDistance { // "hotel.bot.butler.reachdistance", default 3
                runtime.WalkToUnit(bot.Unit, message.Actor.Unit, onReached(bot, message.Actor, entry.DefinitionID))
            } else {
                giveHandItemSequence(ctx, bot, message.Actor, entry.DefinitionID) // Parte 0, primitivo ya existente
            }
        } else {
            active.GiveHandItem(ctx, message.Actor, entry.DefinitionID) // sin caminar, entrega inmediata
        }
        bot.Talk(ctx, i18n("bots.bartender.given", entry.Keyword, message.Actor.Username()))
        return nil
    }
    return nil
}
```
Reusa **directamente** `RoomUnitGiveHanditem`/`HandItemHolder` de `plan/interactions/MEGAPLAN-ESSENTIAL.md` (dar el ítem a sí mismo primero, caminar, luego transferirlo al jugador y limpiar la propia mano) — es, de hecho, el caso de uso real que justificó ese primitivo en la referencia. Ningún primitivo nuevo se construye para este comportamiento.

**`bot_serve_items`** (Parte 3) es la tabla admin-configurable de `palabra clave → definition_id` — un admin decide qué dice "tea"/"coffee"/etc y qué mueble entrega, sin tocar código.

### 5.3 `visitor_log` — registro de visitas mientras el dueño no miraba

```go
// internal/realm/bot/behavior/visitorlog.go
func (visitorLog) OnUserEnter(ctx context.Context, bot *Bot, entering *Occupant, active *roomlive.Room) error {
    if bot.State.ShowedLog { return nil } // una vez por colocación, no se repite
    visits := runtime.roomVisits.Since(active.RoomID, entering.Player.LastOnline()) // Parte 5.3.1
    if len(visits) == 0 {
        bot.Talk(ctx, i18n("bots.visitor.no_visits"))
    } else {
        bot.Talk(ctx, i18n("bots.visitor.visits", len(visits)))
    }
    return nil
}

func (visitorLog) OnUserSay(ctx context.Context, bot *Bot, message ChatMessage, active *roomlive.Room) error {
    if bot.State.ShowedLog || !isAffirmative(message.Text) { return nil } // "sí"/"yes", localizado
    bot.State.ShowedLog = true
    bot.Talk(ctx, formatVisitList(bot.State.PendingVisits)) // lista con sala + fecha/hora por visita
    return nil
}
```
**Dependencia real no trivial**: este comportamiento necesita un registro de "quién entró a qué sala y cuándo" (`ModToolRoomVisit` en la referencia) — un log de visitas por sala. Pixels no tiene ese log todavía. **Decisión de este plan**: implementar `visitor_log` como un comportamiento completo requiere primero una tabla mínima `room_visits(room_id, player_id, entered_at)`, poblada desde el mismo evento `room.access.events.entered` que ya existe (Parte 0) — es una tabla nueva pequeña, no un sistema de moderación completo; se especifica aquí en vez de diferirse porque el costo real es bajo y el comportamiento pierde todo su sentido sin ella.

```sql
--changeset pixels:pixels-bot-0003-create-room-visits
create table room_visits (
    room_id bigint not null,
    player_id bigint not null,
    entered_at timestamptz not null default now()
);
create index room_visits_room_since_idx on room_visits (room_id, entered_at);
```

### Tests
- Bartender: mensaje con la keyword exacta como palabra completa dispara la entrega; como substring de otra palabra (ej. "teapot" conteniendo "tea") no dispara — mismo criterio real de límite de palabra.
- Bartender fuera de `commandDistance` no reacciona.
- Visitor log: primera entrada de un jugador tras ausencia muestra el resumen; responder afirmativamente muestra la lista una sola vez, una segunda afirmación no repite.

---

## Parte 6 — Extensibilidad: comportamientos futuros y SDK (diseño de la interfaz, sin implementación)

**Esto es intencionalmente una interfaz, no una feature completa.** El objetivo de esta parte es que agregar un comportamiento nuevo — incluido uno controlado por un modelo de IA externo — sea "implementar esta interfaz y registrarla", no "reabrir este documento y tocar el runtime".

### 6.1 La interfaz

```go
// internal/realm/bot/behavior/behavior.go
type Behavior interface {
    // Type identifica el discriminador persistido en bots.behavior_type.
    Type() string

    // OnPlace se ejecuta una vez, al colocar el bot en una sala.
    OnPlace(ctx context.Context, bot *Bot, room *roomlive.Room) error

    // OnPickup se ejecuta una vez, al recoger el bot de vuelta al inventario.
    OnPickup(ctx context.Context, bot *Bot) error

    // OnCycle se ejecuta en cada tick del scheduler de sala (500ms) mientras el bot esté colocado.
    OnCycle(ctx context.Context, bot *Bot, room *roomlive.Room) error

    // OnUserSay se ejecuta cuando cualquier jugador habla dentro del alcance de audición del bot.
    OnUserSay(ctx context.Context, bot *Bot, message ChatMessage, room *roomlive.Room) error

    // SaveCustomSkill maneja un skillId no reconocido por el núcleo (Parte 4.2) — el punto de extensión de configuración.
    SaveCustomSkill(ctx context.Context, bot *Bot, skillID int32, data string) error
}
```

### 6.2 Registro (mismo patrón que `BotManager.botDefenitions`, en Go idiomático)

```go
// internal/realm/bot/behavior/registry.go
type Registry struct {
    mutex   sync.RWMutex
    factory map[string]func() Behavior
}

// Register es el punto de extensión — un plugin/SDK de Pixels llama esto una vez, en el arranque,
// para agregar un tipo de bot que el núcleo no conoce.
func (registry *Registry) Register(botType string, factory func() Behavior) error {
    registry.mutex.Lock()
    defer registry.mutex.Unlock()
    if _, exists := registry.factory[botType]; exists {
        return ErrBehaviorAlreadyRegistered // mismo criterio real: Arcturus también rechaza un tipo duplicado
    }
    registry.factory[botType] = factory
    return nil
}

func (registry *Registry) For(botType string) Behavior {
    registry.mutex.RLock()
    defer registry.mutex.RUnlock()
    if factory, found := registry.factory[botType]; found {
        return factory()
    }
    return registry.factory["generic"]() // fallback seguro, nunca un bot sin comportamiento
}
```
Registrado vía `fx.Provide`/`fx.Invoke` en `module.go` — los tres comportamientos reales (Parte 5) se auto-registran al arrancar el realm `bot`, exactamente como cualquier comportamiento futuro externo lo haría.

### 6.3 Puntos de enganche ya previstos pero no conectados (Wired, diferido)

`OnCycle`/`OnUserSay` reciben una función `wiredHook(triggerType, bot, room, args) bool` (Parte 4.3) que hoy **siempre retorna `false`** (Wired no existe) — el día que Wired se implemente, conectar bots a triggers como `SAY_SOMETHING`/`BOT_REACHED_STF`/`BOT_REACHED_AVTR` (los tres confirmados reales en Arcturus, Parte 1.2/1.6) es reemplazar esa función, no tocar ningún `Behavior`.

### 6.4 Por qué esto es suficiente para un futuro "bot con IA", sin construirlo ahora

Un comportamiento controlado por un modelo externo (LLM u otro) sería, en este diseño, exactamente una implementación más de `Behavior`:
- `OnUserSay` podría reenviar el mensaje a un servicio externo y hablar con la respuesta — de forma **asíncrona**, sin bloquear el tick de sala (el contrato de la interfaz no exige que `OnUserSay` responda antes de que retorne; una implementación de IA legítimamente puede lanzar la llamada externa en una goroutine propia y hablar más tarde vía el mismo `bot.Talk` que cualquier otro comportamiento usa).
- `OnCycle` podría decidir movimiento/charla con lógica arbitraria en vez de la fórmula aleatoria de `generic` — el runtime (Parte 4.3) no sabe ni le importa cómo decide un `Behavior`, solo lo invoca.
- `SaveCustomSkill` permite que ese comportamiento tenga su propia configuración (ej. "prompt del sistema", "temperatura", "clave de API") sin que el núcleo de bots sepa nada de eso — persistida como el `data` string libre de `BOT_SKILL_SAVE`, interpretada solo por esa implementación.

**Explícitamente fuera de alcance de este documento**: la implementación real de un comportamiento de IA (qué proveedor, qué prompt, límites de costo/latencia, moderación de la salida de un modelo antes de que un bot la diga en una sala pública). Eso es un plan propio, o directamente un plugin/SDK externo construido sobre esta interfaz — este documento solo garantiza que la interfaz ya alcanza para no necesitar tocar el núcleo cuando llegue el momento.

### Tests
- Registrar dos veces el mismo `botType` rechaza.
- Un `botType` desconocido en una fila de `bots` (dato corrupto o de una versión anterior con un comportamiento removido) resuelve a `generic` en vez de fallar — nunca un bot "sin comportamiento".
- Un `Behavior` de prueba cuyo `OnUserSay` tarda artificialmente 2 segundos no bloquea el tick de otros bots en la misma sala — prueba explícita del contrato asíncrono de 6.4.

---

## Parte 7 — Permisos

```go
var (
    AnyRoomOwnerBots = permission.RegisterNode("bot.any_room_owner", "")   // equivalente a ACC_ANYROOMOWNER para bots
    PlaceAnywhere    = permission.RegisterNode("bot.place_anywhere", "")    // equivalente a ACC_PLACEFURNI
    UnlimitedBots    = permission.RegisterNode("bot.unlimited", "")         // bypass de topes de sala/inventario
)
```

---

## Parte 8 — Rutas HTTP admin

```
GET/POST/PATCH/DELETE /api/admin/bots/serve-items      — CRUD de bot_serve_items (Bartender)
GET    /api/admin/bots/:id                              — inspección de soporte (dueño, sala, comportamiento, estado)
POST   /api/admin/bots/:id/force-pickup                  — recoger administrativamente (bot problemático/abandonado)
```

---

## Parte 9 — Eventos y packets (resumen)

Ver Parte 1.9 para el mapeo completo de eventos de plugin → `pkg/bus`.

| Header | Nombre | Dirección | Parte |
| --- | --- | --- | --- |
| 1592 | `BOT_PLACE` (`botId`,`x`,`y`) | c2s | 4.1 |
| 3323 | `BOT_PICKUP` (`botId`) | c2s | 4.1 |
| 1986 → 1618 | `BOT_CONFIGURATION` → `BOT_COMMAND_CONFIGURATION` | c2s / s2c | 1.7 |
| 2624 | `BOT_SKILL_SAVE` (`botId`,`skillId`,`data`) | c2s | 4.2 |
| — / 69 | `BOT_SKILL_LIST_UPDATE` | — / s2c | 4.2 (push tras guardar) |
| — / 296 | `BOT_FORCE_OPEN_CONTEXT_MENU` | — / s2c | interacción de click en bot |
| — / 639 | `BOT_ERROR` | — / s2c | errores de 4.1 (tope de sala/inventario, tile ocupado, nombre rechazado) |
| — / 3684 | `BOT_RECEIVED` | — / s2c | confirmación de recibir un bot (regalo/trade — cross-ref `plan/MARKET-TRADING.md` si algún día los bots son transferibles por trade, hoy no está en alcance de ese plan) |
| c2s 3848 / s2c 3086 | `USER_BOTS` (inventario) | c2s / s2c | listado de inventario |
| — / 1352 | `ADD_BOT_TO_INVENTORY` | — / s2c | confirma recoger |
| — / 233 | `USER_BOT_REMOVE` | — / s2c | confirma colocar (sale del inventario) |

---

## Parte 10 — Hot paths

El ciclo de IA (Parte 4.3) corre cada 500ms **por bot colocado**, no por sala — en una sala con muchos bots decorativos, esto es el camino más caliente de todo este documento. Mitigado por: (a) el mismo scheduler de proceso ya usado en el resto del proyecto, sin un timer nuevo por bot; (b) `OnCycle` de `generic` es aritmética + una consulta de tile aleatorio en memoria (`RandomWalkableTileAround`, ya existente en el motor de mundo), sin I/O; (c) el flush a base de datos (`needsUpdate`) es asíncrono y solo ocurre cuando algo realmente cambió, no en cada tick.

```go
func BenchmarkBotCycleTick(b *testing.B) { ... } // N bots en una sala, mide el costo total de un tick
```

---

## Parte 11 — Seeding

```sql
--changeset pixels:pixels-bot-seed-development-0001-examples context:development
insert into bots (id, owner_player_id, behavior_type, name, motto, figure, gender, chat_auto, chat_random) values
    (1, 1, 'generic', 'Party Bot', 'Yo!', '...', 1, true, true),
    (2, 1, 'bartender', 'Frank', 'Need a drink?', '...', 1, false, false)
on conflict do nothing;
insert into bot_chat_lines (bot_id, order_num, line) values
    (1, 1, 'Yo!'), (1, 2, 'Hello I''m a real party animal!'), (1, 3, 'Hello!')
on conflict do nothing;
insert into bot_serve_items (keyword, definition_id) values ('tea', 1), ('coffee', 1)
on conflict do nothing;
```

---

## Parte 12 — Milestones (bots)

1. **BT1 — Esquema + colocar/recoger** (Parte 3, 4.1): sin dependencias externas.
2. **BT2 — Ciclo de IA genérico** (Parte 4.3, 5.1): depende de BT1 y del scheduler de proceso ya existente.
3. **BT3 — Configuración/skills** (Parte 4.2): depende de BT1.
4. **BT4 — Seguimiento** (Parte 4.4): depende de BT2.
5. **BT5 — Registro de comportamientos + extensibilidad** (Parte 6): depende de BT2 — es la generalización del propio BT2, no una feature aparte.
6. **BT6 — Bartender** (Parte 5.2): depende de BT5 y del primitivo de hand-item ya existente.
7. **BT7 — Visitor log** (Parte 5.3): depende de BT5 y de la nueva tabla `room_visits`.
8. **BT8 — Cierre de `plan/rooms/BUNDLE.md` Parte 8**: una vez BT1 exista, ese plan puede implementar `CloneRoomBots` sin volver a este documento.

---

# PARTE B — Remanentes de Room / Furniture & Items / Room Entities

Todo lo que sigue está fuera de alcance de: Pets, Wired, furniture de grupos/juegos (diferidos explícitamente en otros documentos), y expresiones/efectos de avatar (`plan/EXPRESSIONS-EFFECTS.md`, que ya incluye el furniture que otorga efectos). Cada ítem se diseña a un nivel de detalle proporcional a su tamaño real — no todos ameritan el mismo desarrollo que Bots.

## Parte 13 — Room: remanentes

| Packet | Diseño |
| --- | --- |
| `ROOM_DELETE`(532) | El propio dueño borra su sala. Reusa `room/record/service.SoftDelete` (ya existe, hoy solo invocado desde admin) — se agrega un comando de jugador con la misma autorización que editar settings (dueño o `ACC_ANYROOMOWNER`), y una confirmación explícita del lado del cliente antes de mandar el packet (el cliente ya la maneja; el servidor no necesita un "¿estás seguro?" de dos pasos). |
| `ROOM_STAFF_PICK`(1918) | Toggle de "sala destacada por staff" — la columna `staff_picked` ya existe en el modelo de sala (`plan/REMAINING-ROOMS.md` la marcó "columna+lectura existen, falta comando de toggle"). Este documento cierra ese gap puntual: un comando admin-only que voltea el flag, sin esquema nuevo. |
| `VOTE_FOR_ROOM`(143) | **Ambigüedad marcada en `plan/STATUS.md` Parte 2.2, resuelta aquí por descarte razonado**: sin más contenido en el spec que un resumen genérico, y sin ninguna clase equivalente en Arcturus (que solo tiene el sistema de `ROOM_LIKE`/`ROOM_SCORE` ya implementado por `plan/rooms/UPVOTES.md`), la hipótesis más simple es que es un alias legacy del mismo "like" bajo un nombre de header distinto de una generación anterior de cliente. **Decisión**: no se implementa un segundo sistema de votos — si una captura real futura demuestra que es una mecánica genuinamente distinta (ej. votación de sala destacada de temporada), se revisita puntualmente. |
| `ROOM_AD_*`(9 packets: purchase/info/search/event-tab/etc.), `ROOM_PROMOTION`(2274) | Sistema de "anuncios de sala" (pagar créditos para promocionar la sala en el navegador) — depende conceptualmente del realm `catalog`/`currency` ya existentes (comprar promoción = un cargo de moneda + un período de visibilidad elevada), pero es una feature de tamaño propio (búsqueda de slots, calendario de disponibilidad, panel de resultados). **Se difiere a un plan propio** (`plan/rooms/ROOM-ADS.md`, no escrito) — no por ser inalcanzable, sino porque mezclarlo aquí diluiría el foco de este documento; se anota la dependencia (catalog+currency) para cuando se aborde. |
| `ROOM_EVENT`/`CANCEL_ROOM_EVENT`/`EDIT_ROOM_EVENT`/`CAN_CREATE_ROOM_EVENT` | Anuncio de "evento en esta sala ahora" en el navegador — más simple que los Ads (sin pago), pero depende de la misma superficie de navegador. Se agrupa con el mismo futuro `ROOM-ADS.md` por cercanía funcional, no se diseña aquí. |
| `CHANGE_QUEUE`/`ROOM_QUEUE_STATUS`/`ROOM_SPECTATOR` | Cola de espera cuando una sala está llena + modo espectador. Depende de un concepto de "capacidad excedida" que hoy simplemente rechaza la entrada (`plan/rooms/ENTRY.md`) — implementar una cola real es una extensión de ese flujo, no de este documento; se anota como dependencia de `ENTRY.md`, no se rediseña aquí. |
| `ROOM_PAINT`(2454) | Sin columnas `wall_paint`/`floor_paint` en el modelo de sala hoy (confirmado en `plan/rooms/BUNDLE.md` Parte 0) — bloqueado por la misma ausencia de esquema que ese documento ya señaló. Se difiere junto con esa dependencia. |
| `SET_ITEM_DATA`/`GET_ITEM_DATA`/`SET_OBJECT_DATA` | Genéricos de configuración de ítem — en la práctica, la mayoría de sus usos reales están cubiertos por Wired (condiciones/triggers, diferido) o por interacciones puntuales ya diseñadas en otros documentos (`plan/interactions/MEGAPLAN-ESSENTIAL.md`). No se identificó un uso real huérfano que justifique un diseño propio — se dejan anotados como "cubiertos indirectamente", no como pendiente activo. |
| `ROOM_DIRECTORY_ROOM_NETWORK_OPEN_CONNECTION`, `SHOW_ENFORCE_ROOM_CATEGORY`, `ROOM_SETTINGS_ERROR`, `ROOM_AMBASSADOR_ALERT` | Cuatro packets de superficie muy chica/muy específica de un flujo administrativo puntual (mensajes de error genéricos, un mecanismo de red de directorio no confirmado en la referencia). Se implementan trivialmente cuando se necesiten como parte de otros flujos (`ROOM_SETTINGS_ERROR`, por ejemplo, es el error genérico que el propio `plan/rooms/CONFIG.md` ya debería emitir) — no ameritan una sección de diseño propia. |

### Tests
- `ROOM_DELETE` por el dueño elimina y notifica a los ocupantes actuales (mismo flujo de expulsión que un `leavecmd` ya existente).
- `ROOM_STAFF_PICK` solo accesible con nodo admin, alterna el flag existente sin tocar ninguna otra columna.

---

## Parte 14 — Furniture & Items: remanentes

### 14.1 Postit (nota adhesiva)

Simple: un mueble de pared con texto libre (extradata) y flag de "es una edición limitada" — sin lógica de caminar. El "sticky pole" (`FURNITURE_POSTIT_SAVE_STICKY_POLE`/`FURNITURE_POSTIT_STICKY_POLE_OPEN`) es una variante que agrupa varias notas en un solo poste, abierto/cerrado como un mueble de dos estados (mismo patrón `ExtraData`-only ya usado en Toggle). Diseño:
```go
// internal/realm/furniture/interactions/essential/postit.go
func (service *Service) SavePostit(ctx context.Context, itemID int64, actorPlayerID int64, text string) error {
    // valida longitud, pasa por wordfilter, persiste en ExtraData — mismo patrón que trofeos/badge-displays de STORE-FINAL.md
}
```

### 14.2 Mannequin (maniquí de outfit)

Guarda un look completo (género+figura+nombre del outfit) elegido por el dueño; **cualquier** jugador del mismo género puede hacer clic para copiarse ese look instantáneamente, sujeto a las mismas reglas de validación de vestimenta que el resto del proyecto ya aplica (HC-exclusivo, etc. — reusa el validador de look existente, no uno nuevo).
```go
func (service *Service) ApplyMannequinLook(ctx context.Context, itemID int64, actorPlayerID int64) error {
    saved, found := service.store.MannequinLook(ctx, itemID)
    if !found || saved.Gender != actorGender { return ErrGenderMismatch }
    newLook := mergeBodyPartsOnly(actorCurrentLook, saved.Figure) // mismo criterio real: solo cabeza+piernas del propio jugador se preservan, el resto viene del maniquí
    return service.players.SetLook(ctx, actorPlayerID, service.lookValidator.Validate(newLook))
}
```
`MANNEQUIN_SAVE_LOOK`(2209)/`MANNEQUIN_SAVE_NAME`(2850) — dos packets separados para guardar el outfit y renombrarlo, ambos solo accesibles al dueño del mueble.

### 14.3 Jukebox / Trax / YouTube Display (reproducción de medios en sala)

El grupo más grande de este remanente (10+ packets: `ADD_JUKEBOX_DISK`, `REMOVE_JUKEBOX_DISK`, `GET_JUKEBOX_PLAYLIST`, `GET_NOW_PLAYING`, `GET_SONG_INFO`, `GET_OFFICIAL_SONG_ID`, `GET_USER_SONG_DISKS`, `GET_SOUND_MACHINE_PLAYLIST`, `SET_YOUTUBE_DISPLAY_PLAYLIST`, `CONTROL_YOUTUBE_DISPLAY_PLAYBACK`, `GET_YOUTUBE_DISPLAY_STATUS`). Modelo conceptual: un mueble "jukebox" tiene una playlist ordenada de "discos" (ítems de inventario específicos, cada uno representando una canción), reproduce una pista a la vez, difunde qué está sonando a toda la sala. El YouTube Display es la variante moderna (reproduce un video de una URL/ID en vez de un disco de catálogo). **Diseño de tamaño propio** (modelo de playlist, códigos de pista, sincronización de "now playing" entre clientes que entran a mitad de una canción) — se anota como candidato a su propio documento (`plan/rooms/MEDIA-FURNITURE.md`, no escrito) dado que mezclar un sistema de reproducción sincronizada con el resto de este remanente le restaría profundidad a ambos. Se dejan aquí solo los headers identificados para que quede constancia de que no se pasaron por alto.

### 14.4 Item Dimmer / Room Toner

Ambos son "herramientas de ambientación": el dimmer controla las luces de mood-light de la sala (color+intensidad+patrón de parpadeo, mismo tipo de dato que un `room_toner` de fondo pero aplicado a la iluminación, no al color de fondo); el toner (`plan/EXPRESSIONS-EFFECTS.md` Parte 1.3 ya documentó su curiosidad del efecto 1337) es el recolor de pared/piso on/off con 3 valores HSL. Ambos comparten el mismo patrón: `ExtraData` con 3-4 valores numéricos, toggle/set vía clic, sin necesidad de un `Behavior` de MEGAPLAN nuevo — se implementan como interacciones simples en `essential/`, análogas a Toggle.

### 14.5 Lovelock (candado del amor) + `FRIEND_FURNI_CONFIRM_LOCK`

Minijuego de dos jugadores: el primero hace clic (queda registrado, se le muestra un diálogo de "esperando a tu pareja"), un segundo jugador **distinto** hace clic (queda registrado como el segundo), y el primero debe **confirmar** (`FRIEND_FURNI_CONFIRM_LOCK`, 3775) para sellar el candado — a partir de ahí el mueble queda permanentemente grabado con ambos nombres, ambos looks (como snapshot visual) y la fecha de sellado, sin poder deshacerse.
```sql
--changeset pixels:pixels-furniture-000X-add-lovelock
alter table furniture_items
    add column lovelock_player_one_id bigint null,
    add column lovelock_player_two_id bigint null,
    add column lovelock_sealed_at timestamptz null;
```
```go
func (service *Service) ClickLovelock(ctx context.Context, itemID int64, actorPlayerID int64) (LovelockState, error) {
    // primer clic: registra player_one, responde "esperando"
    // segundo clic de un player distinto: registra player_two, responde "esperando confirmación del iniciador"
    // el clic de un tercer jugador una vez que ambos slots están llenos: no-op, el candado ya está en proceso o sellado
}
func (service *Service) ConfirmLovelock(ctx context.Context, itemID int64, actorPlayerID int64) error {
    // solo player_one puede confirmar; sella con looks+fecha, vuelve inmutable
}
```

### 14.6 Mystery Box

Un mueble que, al colocarse/abrirse, entrega un premio aleatorio de un pool configurado — conceptualmente muy similar al `InteractionEffectGiver` de `plan/EXPRESSIONS-EFFECTS.md` Parte 4.1 (mismo patrón de "pool aleatorio vía `Source` inyectable"), salvo que el premio es furniture en vez de un efecto de avatar, y hay una espera configurable ("wait message") antes de revelarlo — de ahí los packets `SHOWMYSTERYBOXWAITMESSAGE`/`CANCELMYSTERYBOXWAITMESSAGE`/`GOTMYSTERYBOXPRIZEMESSAGE` (Notifications & Landing, cross-referenciados aquí porque no tienen sentido sin el mueble). Diseño:
```go
func (service *Service) OpenMysteryBox(ctx context.Context, itemID int64, actorPlayerID int64) error {
    // notifica "esperando" (Notifications), espera N segundos (ScheduledTask, mismo patrón de 500ms-aligned de MEGAPLAN),
    // luego concede un premio elegido por Source y notifica el resultado
}
```
Reusa el mismo primitivo de pool aleatorio que effect-giver — ningún mecanismo nuevo, solo un premio de tipo distinto (furniture vs. efecto).

### 14.7 Builders Club — colocación real (`BUILDERS_CLUB_PLACE_WALL_ITEM`/`_ROOM_ITEM`)

`plan/STORE-FINAL.md` Parte 11 ya diseñó el **stub de estado** de Builders Club (siempre "no aplica", tier discontinuado sin implementación real en Arcturus). Estos dos packets son la variante de colocación de furniture bajo ese tier — dado que el stub ya responde "sin límite/no aplica" en todo lo demás, la colocación bajo Builders Club se resuelve **reusando el flujo normal de colocación de furniture**, sin ninguna rama especial — el cliente puede seguir enviando estos headers (compatibilidad legacy) pero el servidor los trata como el `FURNITURE_WALL_UPDATE`/place normal ya implementado. No hace falta diseño propio, solo el wire de aceptar estos dos headers como alias.

### 14.8 Otros headers de superficie chica

`FURNITURE_ALIASES`(c2s+s2c) y `FURNITURE_GROUP_INFO`(c2s) — sin suficiente contexto en la documentación local para diseñar con confianza (no se encontró una clase Arcturus equivalente clara); se anotan como pendientes de una captura real antes de diseñar, mismo criterio de honestidad que `plan/STORE-FINAL.md` Parte 15 ya estableció para casos así.

### Tests
- Mannequin rechaza aplicar el look si el género no coincide, sin mutar nada.
- Lovelock: un tercer jugador haciendo clic tras el segundo no altera el estado; solo `player_one` puede confirmar.
- Mystery box: el premio se revela solo tras el tiempo de espera configurado, nunca antes.

---

## Parte 15 — Room Entities: remanentes no-bot, no-expresión

| Packet | Diseño |
| --- | --- |
| `HAND_ITEM_RECEIVED`(s2c 354), `UNIT_DROP_HAND_ITEM`(c2s 2814), `UNIT_GIVE_HANDITEM`(c2s 2941) | El lado de **unidad** del primitivo de mano que `plan/interactions/MEGAPLAN-ESSENTIAL.md` Parte 1.5 ya construyó para vending/handitem — ese documento diseñó `HandItemHolder` desde la perspectiva del furniture que entrega; estos tres packets son el lado del jugador (soltar voluntariamente lo que tiene en la mano, dárselo a otro jugador adyacente) que ese documento no cubrió porque su foco era la furniture emisora. Se cierra aquí: `UNIT_DROP_HAND_ITEM` vacía el `HandItemHolder` del actor sin destino; `UNIT_GIVE_HANDITEM` lo transfiere al `HandItemHolder` de otro unit adyacente (reusa el mismo primitivo, solo cambia el origen de la transferencia — de "furniture" a "otro jugador"). |
| `UNIT_NUMBER`(s2c 2324) | Numeración de unidades visible en el cliente (ej. overlay de "jugador #3" en ciertos minijuegos/colas) — sin un consumidor real identificado todavía en Pixels (no hay minijuegos ni colas implementadas); se deja como un encoder simple (`roomIndex`+`number`) listo para cuando `ROOM_QUEUE_STATUS`(Parte 13) o un futuro `plan/GAMES.md` lo necesiten, sin lógica de negocio propia hoy. |
| `UNIT_INFO`(s2c 3920) | Empuje genérico de "refrescar toda la info de esta unidad" — un encoder de conveniencia que reusa los mismos datos que ya se sirven en el snapshot de entrada a sala, sin estado nuevo. |

### Tests
- `UNIT_GIVE_HANDITEM` entre dos jugadores no adyacentes rechaza — mismo criterio de adyacencia que el resto de interacciones de mano.
- `UNIT_DROP_HAND_ITEM` limpia el `HandItemHolder` del actor sin afectar al de nadie más.

---

## Parte 16 — Packets: resumen completo de la Parte B

| Header | Nombre | Parte |
| --- | --- | --- |
| 532 | `ROOM_DELETE` | 13 |
| 1918 | `ROOM_STAFF_PICK` | 13 |
| 143 | `VOTE_FOR_ROOM` | 13 (resuelto por descarte razonado) |
| 2248 / 3283 / 2366 | Postit / sticky pole save / sticky pole open | 14.1 |
| 2209 / 2850 | `MANNEQUIN_SAVE_LOOK` / `MANNEQUIN_SAVE_NAME` | 14.2 |
| 3336 / — | `REMOVE_WALL_ITEM` (aliasing de colocación legacy) | 14.7 |
| 2765 | `ONE_WAY_DOOR_CLICK` | ya cubierto por `plan/interactions/MEGAPLAN-ESSENTIAL.md` (Onewaygate) — anotado aquí solo para confirmar que no falta nada |
| 3775 | `FRIEND_FURNI_CONFIRM_LOCK` | 14.5 |
| 1648/2296/2813/2710 | Item dimmer (save/toggle/settings) | 14.4 |
| 711 | `ITEM_PAINT` (room toner) | 14.4 |
| 2833 / 3201 / 596 / 3712 / 2012 / 3074 | Mystery box (llaves, wait/cancel/got prize) | 14.6 |
| 462 / 1051 | Builders Club place wall/room item | 14.7 |
| 2814 / 2941 / 354 | Hand item drop/give/received (lado de unidad) | 15 |
| 2324 / 3920 | `UNIT_NUMBER` / `UNIT_INFO` | 15 |

Diferido explícitamente en otro lugar, no repetido aquí: jukebox/trax/youtube (14.3, candidato a plan propio), room ads/eventos/cola (13, candidato a plan propio), rentable de espacio (`plan/STORE-FINAL.md` Parte 13 / `plan/interactions/TOGGLE-GATE.md`), Wired/Pets/furniture de grupos/juegos (ya diferidos en `plan/STATUS.md`).

---

## Parte 17 — Milestones (Parte B)

1. **RB1 — Room: `ROOM_DELETE` + `ROOM_STAFF_PICK`** (Parte 13): triviales, sin dependencias.
2. **RB2 — Postit + Mannequin + Dimmer + Toner** (Parte 14.1/14.2/14.4): pequeños, sin dependencias entre sí.
3. **RB3 — Lovelock** (Parte 14.5): esquema propio pequeño, sin dependencias.
4. **RB4 — Mystery box** (Parte 14.6): depende del mismo primitivo de pool aleatorio que `plan/EXPRESSIONS-EFFECTS.md` Parte 4.1 ya define — implementar ese primitivo una vez, reusarlo en ambos.
5. **RB5 — Hand item de unidad** (Parte 15): depende de que exista el `HandItemHolder` de `MEGAPLAN-ESSENTIAL.md` (ya implementado según `plan/STATUS.md`).
6. **RB6 — Builders Club, colocación** (Parte 14.7): depende del stub ya implementado de `STORE-FINAL.md` Parte 11.

### Fuera de este documento, candidatos a plan propio
- `plan/rooms/MEDIA-FURNITURE.md` (Parte 14.3, jukebox/trax/youtube display).
- `plan/rooms/ROOM-ADS.md` (Parte 13, anuncios/eventos/cola de sala).
- `FURNITURE_ALIASES`/`FURNITURE_GROUP_INFO` (Parte 14.8) — pendientes de captura real antes de diseñar.
