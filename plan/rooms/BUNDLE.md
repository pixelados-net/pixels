# Plan: Room Bundles (comprar una sala completa desde el catálogo)

Retoma `plan/STORE-FINAL.md` Parte 4, que dejó Room Bundles diferido con un análisis breve. Esta vuelta lo diseña completo, al mismo nivel que el resto de `plan/rooms/*` — con una sola pieza realmente diferida (clonar bots, porque no existe ningún realm de bots en Pixels) en vez del defer total que `STORE-FINAL.md` había aplicado. La investigación de esta vuelta confirma que las otras dos razones que motivaron ese defer (falta de un servicio de clonado de salas, incertidumbre del propio código real) ya no bloquean nada: `world/layout.RoomManager.SaveCustom`/`FindCustomByRoomID` y `furniture/repository.ItemReader.ListRoomItems` ya existen y alcanzan para clonar layout y furniture sin construir infraestructura nueva de bajo nivel.

Investigado contra `legacy/Arcturus-Community` (`habbohotel/catalog/layouts/{RoomBundleLayout,SingleBundle}.java`, ya leídos completos en la investigación de `STORE-FINAL.md`, releídos aquí con foco en el clonado) y el estado real actual de Pixels (`internal/realm/room/{record,world/layout}`, `internal/realm/furniture/repository`, `internal/realm/navigator/create/room`).

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding, confirmado leyendo el código actual)

| Ya existe | Dónde | Cómo lo usa este plan |
| --- | --- | --- |
| `furniture/repository.ItemReader.ListRoomItems(ctx, roomID)` | `internal/realm/furniture/repository/contract.go` | Enumera el furniture (piso+pared) de la sala plantilla — sin esto habría que construir un listado desde cero; ya existe. |
| `world/layout.RoomManager.FindCustomByRoomID`/`SaveCustom` | `internal/realm/room/world/layout/custom.go`, `database/layout/custom.go` | Lee el layout custom de la plantilla (si lo tiene, `plan/rooms/FLOORPLAN.md`) y lo re-aplica tal cual a la sala nueva — ninguna lógica de clonado de heightmap que inventar. |
| `room/record/service.Create(ctx, CreateParams)` | `internal/realm/room/record/service/room.go` | Base para crear la sala destino — este plan solo agrega un paso posterior de clonado, no reemplaza la creación normal. |
| `navigator/create/room` (`ListByOwner` + `MaxRoomsPerPlayer`) | `internal/realm/navigator/create/room/command.go`, `room/record/service/errors.go` | Mismo chequeo de tope de salas ya aplicado a la creación manual — Room Bundles lo reusa tal cual, mismo límite, mismo error. Nota: el comentario real del código (`"the room ownership limit before subscriptions exist"`) ya anticipa que este tope se volverá club-aware el día que el realm `subscription` (`plan/STORE-FINAL.md`) lo consuma — **no es una tarea de este plan**, se señala para que quede constancia. |
| Pipeline de compra unificado (`purchaseOffer`/`commitPurchase`) | `internal/realm/catalog/service/purchase.go`, `plan/STORE-FINAL.md` Parte 1.1/3.2 | Room Bundles se integra como una tercera rama de `commitPurchase` (junto a "grant simple" y "grant multi-producto") — mismo principio ya establecido: un único pipeline de compra, no uno nuevo por feature. |
| `catalog_items.bundle_discount_enabled`, `catalog_item_products` | `plan/STORE-FINAL.md` Parte 3.1/3.2 | Confirmado: **no aplican** a Room Bundles — un bundle de sala nunca es elegible a descuento por cantidad ni es un "multi-producto" (es una rama de negocio distinta, ver Parte 3). |
| `sharedmodel.Base`, `roommodel.Room` (sin `WallPaint`/`FloorPaint` como columnas propias) | `internal/realm/room/record/model/room.go` | Confirma que no hay pintura de pared/piso a nivel de columna de sala que clonar aparte de lo que `CreateRecordParams` ya cubre — nada que inventar aquí. |
| Ningún realm de `bots` | confirmado por `find internal/realm -iname "*bot*"` (vacío) | Único punto real de este plan que se difiere — Parte 8. |
| `BubbleAlertComposer`/`GENERIC_ALERT` ya existentes | `networking/outbound/session/{alert,bubblealert}` | Reusado para la notificación de "sala creada" tras comprar un bundle — mismo mecanismo que Arcturus, sin packet nuevo. |

---

## Parte 1 — Investigación (Arcturus, `RoomBundleLayout`/`SingleBundle`)

### 1.1 Qué es realmente un "room bundle"

No es un mueble ni un conjunto de muebles — es una **sala plantilla completa** (`roomId` fijo, configurado por un admin) que se clona íntegra a nombre del comprador: fila de `rooms` (dueño reasignado), cada fila de `items` (piso y pared, heightmap/layout custom incluido si la plantilla lo tiene), y — solo si `bundle.bots.enabled` — cada bot de la plantilla.

### 1.2 `getCatalogItems()` — el mueble sintético (no se replica tal cual)

Arcturus recalcula el contenido del "bundle" cada vez que se lista la página (con throttle de 120s): cuenta cada instancia de furniture en la sala plantilla, la codifica en un string compuesto (`itemId:cantidad;...`) y lo reasigna al único `CatalogItem` sintético de esa página, para que el cliente muestre "esto es lo que incluye" antes de comprar. Es un mecanismo de **preview**, no de compra — la compra real vuelve a leer la sala plantilla en el momento de `buyRoom`, no usa el string cacheado. Pixels no necesita el recálculo periódico con throttle: al no cachear una "oferta" separada del catálogo (el catálogo de Pixels ya sirve `Offer.Products[]` en tiempo real desde `projection.Offer`, `plan/STORE-FINAL.md` Parte 3.2), el preview se resuelve leyendo `ListRoomItems(templateRoomID)` en el momento de servir `CATALOG_PAGE`, agrupado por `DefinitionID` — sin cache propia, sin throttle, siempre exacto. Si esto resulta demasiado costoso en producción (poco probable: una plantilla tiene a lo sumo un puñado de docenas de ítems), se cachea con el mismo mecanismo de cache de catálogo que ya existe (`catalogservice.cache`, invalidado por `Refresh`), no uno nuevo.

### 1.3 `buyRoom(habbo)` — el clonado real, paso a paso

1. Si la plantilla no está cargada en memoria, la carga (`loadItems`).
2. **Chequeo de tope de salas — con una condición real notable**: `if (habbo != null) { chequear tope }` — es decir, **el chequeo de tope se salta por completo si `habbo` es null** (invocación vía RCON/consola, sin un jugador real detrás). Pixels no tiene un camino equivalente de "compra sin jugador" — se documenta el hallazgo pero no se replica la condición (siempre hay un `PlayerID` real en el flujo de compra de Pixels).
3. Persiste cualquier cambio pendiente de la plantilla (`room.save()`) y ejecuta el `run()` de cada `HabboItem` (flush de estado) — en Pixels no hace falta: la plantilla ya está persistida en Postgres en todo momento, no hay estado "sucio" en memoria que sincronizar primero.
4. Clona `rooms` (`INSERT ... SELECT ... WHERE id=plantilla`, reasignando `owner_id`/`owner_name`), clona `items` (`INSERT ... SELECT`, reasignando `user_id`/`room_id`, forzando `wired_data`/posición tal cual, y **reseteando `limited_data` a `"0:0"`** — un mueble clonado nunca hereda el número de serie LTD de la plantilla), clona `room_models_custom` si la plantilla tiene layout personalizado, y clona bots si `bundle.bots.enabled`.
5. Reaplica visualmente algunos campos de la sala nueva (`wallHeight`, `floorSize`, `wallPaint`, `floorPaint`) — confirma que estos SÍ son conceptos reales de Habbo aunque Pixels no los modele hoy como columnas propias de `rooms` (Parte 0) — no bloquea este plan, se nota para cuando esas columnas existan.
6. Notifica al comprador con un `BubbleAlertComposer` de "sala comprada".

**Incertidumbre admitida en el propio código real** (`CatalogBuyItemEvent`, hallazgo ya citado en `STORE-FINAL.md` Parte 4): un comentario `// not sure if this is how it should be handled :S` en la rama de error cuando se alcanza el tope de salas — la propia referencia no está segura de su manejo de errores acá. Pixels no hereda esa incertidumbre: define un error explícito y un packet de fallo claro (Parte 6), sin ambigüedad.

### 1.4 Qué NO dispara este flujo (fragmentación real, ya documentada en `STORE-FINAL.md` 1.1)

`buyRoom` no dispara `UserCatalogItemPurchasedEvent` ni `UserCatalogFurnitureBoughtEvent`, y no escribe `CatalogPurchaseLogEntry` — es uno de los 4 caminos de compra fragmentados de Arcturus. Pixels lo unifica: la compra de un room bundle pasa por el mismo `commitPurchase` transaccional que cualquier otra oferta (Parte 6), así que **sí** dispara `catalog.purchased` y queda auditada — una mejora deliberada, no una discrepancia a explicar.

---

## Parte 2 — Decisiones de diseño

1. **Una plantilla es una sala real, marcada** (`rooms.is_bundle_template`), no un concepto separado. Arcturus no marca nada — cualquier `roomId` sirve de plantilla, lo cual permitiría por accidente clonar la sala de un jugador real. Pixels agrega el flag como mejora deliberada: una plantilla se excluye de navegador/búsqueda/tope de salas del admin que la construyó, y solo una sala marcada puede referenciarse desde un `catalog_item`.
2. **El clonado es puramente de persistencia**, no toca el runtime de sala (`roomlive`/`world`). La sala nueva no tiene ningún occupant ni fixture cache viva en el momento de la compra — se construye normalmente la primera vez que alguien entra (mismo camino que cualquier sala recién creada). No hace falta invocar `Room.LoadWorld` ni `ReloadFixtures` desde este plan.
3. **El preview del catálogo no cachea nada propio** — lee `ListRoomItems` en el momento de armar `CATALOG_PAGE`, sin el throttle de 120s de la referencia (innecesario a esta escala, Parte 1.2).
4. **Bots: diferido**, no simplificado ni omitido en silencio — Parte 8 explica exactamente qué falta y qué hará un futuro plan cuando el realm de bots exista.
5. **El tope de salas siempre se chequea** (a diferencia de la referencia, que lo salta si `habbo==null`) — en Pixels toda compra tiene un jugador real detrás, así que no hay atajo que replicar.

---

## Parte 3 — Esquema

```sql
--liquibase formatted sql
--changeset pixels:pixels-room-000X-add-bundle-template
alter table rooms add column is_bundle_template boolean not null default false;
--rollback alter table rooms drop column is_bundle_template;
```

```sql
--changeset pixels:pixels-catalog-000X-add-room-bundle
alter table catalog_items add column room_bundle_template_room_id bigint null references rooms(id);
--rollback alter table catalog_items drop column room_bundle_template_room_id;
```
Un `catalog_items` con `room_bundle_template_room_id` no-nulo es, por definición, un room bundle — mutuamente excluyente con `DefinitionID`/`Amount` (que quedan en `0`/sentinela, igual que un ítem de `catalog_item_products`, `plan/STORE-FINAL.md` Parte 3.2) y con `bundle_discount_enabled` (siempre `false`, no elegible a descuento por cantidad — Parte 1.4 lo confirma real: `buyRoom` nunca pasa por `calculateDiscountedPrice`).

```sql
--changeset pixels:pixels-room-000X-create-bundle-purchase-log
create table room_bundle_purchases (
    id bigint generated always as identity primary key,
    catalog_item_id bigint not null references catalog_items(id),
    template_room_id bigint not null references rooms(id),
    created_room_id bigint not null references rooms(id),
    buyer_player_id bigint not null,
    purchased_at timestamptz not null default now(),
    furniture_item_count integer not null
);
create index room_bundle_purchases_buyer_idx on room_bundle_purchases (buyer_player_id);
--rollback drop table if exists room_bundle_purchases;
```
Auditoría dedicada (a diferencia de un `catalog.purchased` genérico, acá vale la pena guardar explícitamente qué plantilla generó qué sala — soporte/investigación de abuso de plantillas).

---

## Parte 4 — Administración de plantillas

Una plantilla se construye **con las herramientas normales** (un admin/builder entra a una sala común, la decora con FLOORPLAN.md/furniture normal) y luego se marca como plantilla — no hay un "editor de plantillas" separado que construir.

```
POST   /api/admin/rooms/:roomId/bundle-template          — marca una sala existente como plantilla (is_bundle_template=true)
DELETE /api/admin/rooms/:roomId/bundle-template          — desmarca (falla si algún catalog_item activo todavía la referencia)
GET    /api/admin/rooms/bundle-templates                  — lista todas las plantillas activas
```
`POST /api/admin/catalog/items` (ya existente, `plan/CATALOG.md`) se extiende para aceptar `RoomBundleTemplateRoomID` en `ItemInput`/`ItemPatch` — sin una ruta nueva de catálogo, solo un campo más en el CRUD que ya existe.

### Tests
- Marcar como plantilla una sala con jugadores dentro no falla (una plantilla puede seguir siendo una sala "real" mientras tanto) pero sí excluye la sala de resultados de navegador desde ese momento.
- Desmarcar una plantilla todavía referenciada por un `catalog_item` habilitado rechaza con un error explícito, no falla en silencio.

---

## Parte 5 — Clonado

### 5.1 Servicio

```go
// internal/realm/room/record/service/bundle.go
func (service *Service) CloneAsBundle(ctx context.Context, params CloneBundleParams) (roommodel.Room, error) {
    template, found, err := service.store.FindRoomByID(ctx, params.TemplateRoomID)
    if err != nil || !found || !template.IsBundleTemplate {
        return roommodel.Room{}, ErrInvalidBundleTemplate
    }

    owned, err := service.ListByOwner(ctx, params.BuyerPlayerID)
    if err != nil { return roommodel.Room{}, err }
    if len(owned) >= MaxRoomsPerPlayer { // mismo tope y mismo error que la creación manual, Parte 0
        return roommodel.Room{}, ErrRoomLimitReached
    }

    var created roommodel.Room
    err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
        created, err = service.store.CreateRoom(txCtx, cloneRoomParams(template, params))
        if err != nil { return err }

        if customLayout, found, err := service.layouts.FindCustomByRoomID(txCtx, template.ID); err != nil {
            return err
        } else if found {
            if _, err := service.layouts.SaveCustom(txCtx, cloneCustomParams(customLayout, created.ID)); err != nil {
                return err
            }
        }

        clonedItems, err := service.furniture.CloneRoomItems(txCtx, furnitureservice.CloneParams{
            SourceRoomID: template.ID, TargetRoomID: created.ID, TargetOwnerID: params.BuyerPlayerID,
        })
        if err != nil { return err }

        return service.store.RecordBundlePurchase(txCtx, params.CatalogItemID, template.ID, created.ID, params.BuyerPlayerID, len(clonedItems))
    })

    return created, err
}
```

### 5.2 Clonado de furniture — reglas de reset explícitas

```go
// internal/realm/furniture/service/clone.go
func (service *Service) CloneRoomItems(ctx context.Context, params CloneParams) ([]furnituremodel.Item, error) {
    source, err := service.store.ListRoomItems(ctx, params.SourceRoomID)
    if err != nil { return nil, err }
    return service.store.CloneItems(ctx, params.TargetRoomID, params.TargetOwnerID, source)
}
```
Campos preservados exactamente: `DefinitionID`, `X`/`Y`/`Z`/`Rotation`/`WallPosition`, `ExtraData`. Campos reseteados (confirmado real, Parte 1.3 paso 4, mismo criterio que la referencia):

| Campo | Valor tras clonar | Por qué |
| --- | --- | --- |
| `OwnerPlayerID` | comprador | obvio |
| `RoomID` | sala nueva | obvio |
| `LimitedEditionNumber` | `null` | rompe la cadena de serie LTD, igual que `limited_data="0:0"` en la referencia |
| `GiftWrapped`/`GiftWrapBoxID`/`GiftWrapRibbonID`/`GiftSenderPlayerID`/`GiftMessage` | `false`/`null` | un clon nunca es en sí mismo un regalo, aunque la plantilla tuviera un ítem envuelto (caso de borde improbable pero explícito) |

Una sola sentencia `INSERT INTO furniture_items (...) SELECT ... FROM furniture_items WHERE room_id=$source` (mismo idioma SQL que la referencia, adaptado a los nombres de columna de Pixels) — no un loop de `Grant` por ítem, para mantenerlo en una sola operación atómica de escritura sin importar cuántos muebles tenga la plantilla.

### Tests
- Clonar una plantilla con layout custom reproduce el heightmap/puerta exactos en la sala nueva.
- Un ítem LTD en la plantilla se clona sin número de serie — no consume ni reclama ningún cupo de la serie limitada original.
- Un ítem envuelto como regalo en la plantilla (caso de borde) se clona desenvuelto.
- Clonar preserva `ExtraData` (color, estado de interacción) tal cual — un semáforo en verde en la plantilla aparece en verde en la sala nueva, no reseteado.

---

## Parte 6 — Integración con el pipeline de compra

```go
// service/purchase.go — commitPurchase, tercera rama
func (service *Service) commitPurchase(ctx context.Context, params PurchaseParams, item catalogmodel.Item, result *PurchaseResult) error {
    switch {
    case item.RoomBundleTemplateRoomID != nil:
        room, err := service.rooms.CloneAsBundle(ctx, roomservice.CloneBundleParams{
            TemplateRoomID: *item.RoomBundleTemplateRoomID, BuyerPlayerID: params.PlayerID, CatalogItemID: item.ID,
        })
        if err != nil { return err }
        result.CreatedRoomID = &room.ID
    case len(products) > 0:
        // rama multi-producto, ya diseñada en STORE-FINAL.md Parte 3.2
    default:
        // rama simple, ya existente
    }
    // charge() y publishPurchase() son comunes a las tres ramas — sin duplicar cobro ni evento
}
```
`purchaseOffer` (validación de página/club/habilitado) no cambia — un room bundle es una oferta de catálogo como cualquier otra hasta el momento de conceder. `ErrRoomLimitReached` se mapea al mismo packet de fallo de compra ya existente (`CATALOG_PURCHASE_ERROR`/`3770`, `plan/STORE-FINAL.md` Parte 1.4), con un código propio en vez del comentario de incertidumbre que la referencia deja (Parte 1.3).

### Tests
- Comprar un room bundle sin cupo de salas rechaza **antes** de cobrar — ningún crédito se descuenta si el clonado no puede completarse.
- Comprar un room bundle dispara `catalog.purchased` con `result.CreatedRoomID` poblado — a diferencia de la referencia, sí queda auditado (Parte 1.4).
- Comprar un room bundle nunca pasa por el cálculo de descuento por cantidad, incluso si `Amount` llegara a ser `>1` por error de cliente — se rechaza como cualquier oferta no elegible.

---

## Parte 7 — Tope de salas por jugador

Reusa `MaxRoomsPerPlayer`/`ListByOwner` exactamente como `navigator/create/room` ya lo hace (Parte 0) — mismo límite, mismo error, para que comprar un room bundle nunca sea una forma de esquivar el tope de salas manuales. Sin excepción tipo "si no hay jugador" (Parte 1.3/2.5) — en Pixels toda compra tiene un `PlayerID` real.

---

## Parte 8 — Bots (diferido, con justificación y hoja de ruta)

**Por qué se difiere**: no existe ningún realm de bots en Pixels (confirmado, Parte 0) — ni modelo, ni persistencia, ni runtime, ni ningún packet de bots implementado. Clonar bots de una plantilla no tiene destino posible hoy, literalmente no hay tabla `bots` ni concepto de bot vivo en una sala.

**Decisión para este plan**: las plantillas de Pixels **no incluyen bots**. Un admin puede construir una plantilla decorativa completa (furniture, layout custom) sin bots, y el clonado de Parte 5 simplemente no busca ninguno — no es una limitación silenciosa, es una ausencia de feature explícita y documentada.

**Lo que un futuro plan de bots necesitaría** para cerrar esta brecha (dejado documentado, no prescrito en detalle — pertenece a ese plan, no a este):
- Un modelo `bot.Bot` (nombre, motto, figure, gender, posición, líneas de chat, tipo) y su propio realm, análogo en estructura a `furniture`/`room`.
- Una extensión de `CloneRoomItems`-style: `CloneRoomBots(ctx, sourceRoomID, targetRoomID, targetOwnerID)` — mismo patrón de "una sola sentencia de clonado", nada nuevo que inventar en la mecánica una vez que el modelo exista.
- Una bandera de configuración equivalente a `bundle.bots.enabled` de la referencia, para permitir desactivar el clonado de bots por plantilla sin quitarlos de la plantilla misma.
- Actualizar `commitPurchase`'s rama de room bundle (Parte 6) para invocar el clonado de bots junto al de furniture, dentro de la misma transacción — un solo punto de cambio, ya que la rama ya existe.

---

## Parte 9 — Comandos y servicio Go

```
internal/realm/room/record/
  service/bundle.go            # CloneAsBundle, cloneRoomParams, cloneCustomParams
  admin/bundle_template.go      # marcar/desmarcar/listar plantillas

internal/realm/furniture/
  service/clone.go              # CloneRoomItems
  repository/clone.go           # CloneItems (INSERT...SELECT)

internal/realm/catalog/
  model/item.go                 # + RoomBundleTemplateRoomID *int64
  service/purchase.go           # tercera rama de commitPurchase
  admin/contract.go              # + RoomBundleTemplateRoomID en ItemInput/ItemPatch
```
Ningún comando/handler de red nuevo — la compra sigue siendo `catalog.item.buy` (ya existente), la administración de plantillas son rutas HTTP, no comandos de protocolo de juego.

---

## Parte 10 — Eventos (`pkg/bus`)

```
room.bundle.purchased  {CatalogItemID, TemplateRoomID, CreatedRoomID, BuyerPlayerID, FurnitureItemCount}
```
Emitido junto a (no en reemplazo de) `catalog.purchased` — un consumidor interesado solo en "se creó una sala nueva" no necesita filtrar el evento genérico de catálogo por tipo de ítem.

---

## Parte 11 — Permisos

```go
var (
    BundleTemplateManage = permission.RegisterNode("room.admin.bundle_template.manage", "")
)
```
Ninguno nuevo para la compra en sí — usa exactamente las mismas reglas de acceso de página/ítem que cualquier oferta de catálogo (`plan/CATALOG.md`).

---

## Parte 12 — Packets

Ninguno nuevo. La compra reusa `CATALOG_PURCHASE`(3492)/`CATALOG_PURCHASE_OK`(869)/`CATALOG_PURCHASE_ERROR`(1404) tal cual (`plan/STORE-FINAL.md`). La notificación de "sala creada" reusa `GENERIC_ALERT`/`BUBBLE_ALERT` ya existentes (Parte 0) — ningún composer dedicado, a diferencia de lo que un lector apurado de Arcturus podría asumir al ver un `BubbleAlertComposer` con una key propia (`PURCHASING_ROOM`): esa key es solo un texto localizado, no un packet distinto.

---

## Parte 13 — Seeding

```sql
--liquibase formatted sql
--changeset pixels:pixels-room-seed-development-0001-bundle-template context:development
insert into rooms (id, owner_player_id, owner_name, name, description, model_name, door_mode, max_users, is_bundle_template)
overriding system value values
    (100, 1, 'system', 'bundle_template_starter_loft', 'Plantilla: loft inicial', 'model_a', 0, 25, true)
on conflict do nothing;
-- + un puñado de catalog_items existentes (ya sembrados en CATALOG.md) colocados a mano en la sala 100 para poblar la plantilla.

insert into catalog_pages (id, parent_id, name, layout, order_num, visible, enabled)
overriding system value values
    (8, 1, 'room_bundles', 'default_3x3', 6, true, true)
on conflict do nothing;

insert into catalog_items (id, page_id, name, cost_credits, cost_points, points_type, room_bundle_template_room_id, order_num, enabled, extra_data)
overriding system value values
    (9, 8, 'loft_starter_bundle', 75, 0, -1, 100, 1, true, '0')
on conflict do nothing;
```
No se siembra ningún bot (Parte 8) — la plantilla de desarrollo es puramente decorativa.

---

## Parte 14 — Hot paths

Ninguno — comprar un room bundle es una acción deliberada y poco frecuente del jugador (mismo criterio ya aplicado en `STORE-FINAL.md`/`MESSENGER.md`). El único costo no trivial es la sentencia `INSERT...SELECT` de clonado de furniture, acotada por el tamaño de la plantilla (unas pocas docenas de ítems en la práctica) — no amerita benchmark dedicado.

---

## Parte 15 — Testing (resumen transversal)

Cubierto en detalle por cada Parte 4-7. Transversal:
- Ninguna prueba de este plan depende de bots — la ausencia de bots en una plantilla es el comportamiento esperado, no un caso de error a simular.
- Una plantilla puede reusarse para múltiples compras concurrentes sin interferencia — cada clonado lee la plantilla (solo lectura) y escribe exclusivamente en la sala nueva, sin ningún estado compartido mutable entre compras.

---

## Parte 16 — Milestones de implementación

1. **RB1 — Esquema** (Parte 3): columnas + tabla de auditoría, sin dependencias.
2. **RB2 — Administración de plantillas** (Parte 4): depende de RB1.
3. **RB3 — Clonado de furniture** (Parte 5.2): depende de RB1, independiente de RB2.
4. **RB4 — Clonado de sala + layout custom** (Parte 5.1): depende de RB3.
5. **RB5 — Integración con el pipeline de compra** (Parte 6): depende de RB2 y RB4.
6. **RB6 — Seeding + validación end-to-end** (Parte 13): depende de RB5.

### Milestones futuros confirmados (fuera de este documento, no descartados)
- **Clonado de bots** (Parte 8) — depende de que exista un realm de bots; cuando exista, se conecta a la misma rama de `commitPurchase` ya construida en RB5, sin rediseñar nada de este plan.
- **Columnas de pintura de pared/piso a nivel de sala** (si Pixels las modela alguna vez como columnas propias en vez de furniture) — el clonado de Parte 5.1 se extendería a copiarlas, mismo patrón.
