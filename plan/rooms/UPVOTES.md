# Plan: Likes / Upvotes de Room (`internal/realm/room/control/votes`)

## Estado de implementación

Implementado completamente. Las decisiones finales confirmadas contra Nitro son:

- Inbound `ROOM_LIKE` usa header `3582` y contiene el valor positivo `rating=1`; el room se resuelve desde la presencia activa y nunca desde el payload.
- Outbound `ROOM_SCORE` usa header `482` y contiene `score int32` seguido de `canLike bool`.
- La elegibilidad de todos los ocupantes se resuelve con una sola consulta batch y se reutilizan únicamente dos packets codificados por actualización.
- Las rutas protegidas son `GET /api/admin/room-votes/status`, `GET /api/admin/room-votes/list` y `POST /api/admin/room-votes/cast`.
- La persistencia vive en `control/votes` más `database/votes`, sin ampliar el ya completo package `record/service`.

Implementa completamente la **Parte 5** de `plan/REMAINING-ROOMS.md` — votar una sala sube su `Score` una sola vez por jugador, de por vida, nunca baja. El diseño original de Arcturus ya estaba bien pensado en `REMAINING-ROOMS.md`; este documento lo cruza contra la implementación real de un fork activo y mejorado (**Polaris-Emulator**, `github.com/duckietm/Polaris-Emulator`, Java, confirmado por lectura directa del código fuente vía GitHub) y corrige un problema de consistencia real que encontré ahí.

Es un plan solamente — no se escribió código Go todavía.

---

## Parte 0 — Punto de partida real (grounding, confirmado leyendo el código actual)

| Ya existe | Dónde | Nota |
| --- | --- | --- |
| `roommodel.Room.Score int` | `internal/realm/room/record/model/room.go` | Ya existe la columna/campo — nunca se escribe, solo se lee para ordenar (`ListHighestScore`). |
| `record.service.Manager.ListHighestScore(ctx, limit)` | `internal/realm/room/record/service/contract.go` | Ya ordena por `Score` — este plan es el primero en darle un mecanismo real para subir. |
| Ninguna tabla `room_votes` | grep sin resultados | Greenfield total. |
| `control.Actor`/`control.MatchRoom` | `internal/realm/room/control/commands/resolve/session.go` | Helper ya usado por moderación/derechos/settings para resolver actor + room actual y validar que el packet apunta al room correcto — se reusa acá, no se reimplementa. |
| `broadcast.RoomPacket(ctx, connections, active, packet, excludedPlayerID)` | `internal/realm/room/runtime/broadcast/broadcast.go` | Ya tolera fallos de envío individuales sin abortar el resto — se reusa para el broadcast de score. |
| `bus.Publisher`/eventos de settings/moderación ya establecidos | `internal/realm/room/control/events/*` | Mismo patrón de evento tras cada mutación exitosa — este plan agrega el suyo. |
| Patrón completo de comando de settings (`control/commands/settings/save.go`) | mismo paquete | Sirve de plantilla exacta de wiring: resolver actor+room, autorizar, mutar con versión optimista, broadcastear, publicar evento — este plan sigue la misma forma para votar. |

---

## Parte 1 — Lo que confirmé en Polaris-Emulator (no en Arcturus original, en el fork real)

Leí directamente `RoomManager.java` (método `voteForRoom`/`hasVotedForRoom`) y `RoomVoteEvent.java` del repo `duckietm/Polaris-Emulator`. Confirma exactamente lo que `REMAINING-ROOMS.md` ya había investigado, con un detalle adicional real:

```java
public void voteForRoom(Habbo habbo, Room room) {
    if (habbo.getHabboInfo().getCurrentRoom() != null && room != null && habbo.getHabboInfo().getCurrentRoom() == room) {
        if (this.hasVotedForRoom(habbo, room)) return;
        room.setScore(room.getScore() + 1);
        habbo.getHabboStats().votedRooms.add(room.getId());
        for (Habbo h : room.getHabbos()) {
            h.getClient().sendResponse(new RoomScoreComposer(room.getScore(), !this.hasVotedForRoom(h, room)));
        }
        // INSERT INTO room_votes (user_id, room_id) VALUES (?, ?) — recién acá, al final
    }
}

boolean hasVotedForRoom(Habbo habbo, Room room) {
    if (room.getOwnerId() == habbo.getHabboInfo().getId()) return true; // el dueño "ya votó" — truco para bloquearlo sin un branch aparte
    return habbo.getHabboStats().votedRooms.contains(room.getId());
}
```

**Confirmado, sin sorpresas respecto al diseño original**:
- Votar exige estar **físicamente presente en ESE room ahora mismo** (`getCurrentRoom() == room`) — no es "votar por ID desde cualquier lado", es un voto sobre la sala en la que estás parado. `REMAINING-ROOMS.md` ya decía "sin payload más que mi room actual" — esto lo confirma.
- El dueño queda bloqueado tratándolo como "ya votó" (mismo resultado que un chequeo aparte, truco de implementación, no una semántica distinta).
- Se broadcastea a **todos** los ocupantes, cada uno con su propio `canVote` recalculado individualmente (`!hasVotedForRoom(h, room)`) — exactamente como ya estaba diseñado.
- `room_votes(user_id, room_id)` sin ninguna columna de fecha — confirmado, es la tabla real.

**Problema real que encontré, y que este plan corrige**: el orden de operaciones en Polaris es **memoria primero, base de datos después** — `room.setScore(...)`, el broadcast, y el cache de sesión (`votedRooms.add(...)`) pasan **antes** del `INSERT` a `room_votes`, y si el `INSERT` falla (excepción SQL), el código solo lo loguea — **no revierte nada**. Eso significa que un fallo transitorio de base de datos deja el score en memoria ya incrementado, ya broadcasteado a todos, y el jugador ya no puede volver a votar en esa sesión (por el cache local), pero la fila nunca quedó en `room_votes` — el voto se "pierde" de forma silenciosa desde la perspectiva de la base, y si el room se recarga de memoria antes de que alguien note el problema, el score persistido en Postgres nunca reflejó ese voto. Este plan invierte el orden: **persistir primero, en una sola transacción, y solo si eso tiene éxito se muta el estado en memoria y se broadcastea** (Parte 3).

---

## Parte 2 — Esquema

```sql
create table room_votes (
    room_id bigint not null references rooms(id) on delete cascade,
    player_id bigint not null references players(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (room_id, player_id)
);
```
Igual al sketch original de `REMAINING-ROOMS.md` — `created_at` es la única adición sobre la tabla real de Polaris (`room_votes` no tiene fecha), siguiendo la convención ya establecida en el resto del proyecto de que toda tabla nueva registra cuándo se creó cada fila, aunque el propio voto sea "de por vida" y no necesite expirar nunca.

---

## Parte 3 — `record.service`: `Vote`, persistido primero, atómico

```go
// internal/realm/room/record/service/vote.go

// VoteResult contains the outcome of a room vote attempt.
type VoteResult struct {
    // Score stores the resulting room score after the vote.
    Score int
    // Voted reports whether this call actually registered a new vote (false when
    // the player had already voted, or is the owner — a no-op, not an error).
    Voted bool
}

// Vote records one player's vote for a room, exactly once per player, ever.
func (service *Service) Vote(ctx context.Context, roomID int64, playerID int64) (VoteResult, error) {
    room, found, err := service.store.FindRoomByID(ctx, roomID)
    if err != nil {
        return VoteResult{}, err
    }
    if !found {
        return VoteResult{}, ErrRoomNotFound
    }
    if room.OwnerPlayerID == playerID {
        return VoteResult{Score: room.Score}, nil // no-op silencioso, mismo criterio que Polaris: el dueño nunca vota su sala
    }

    return service.store.InsertVoteAndIncrementScore(ctx, roomID, playerID)
}
```

`Store.InsertVoteAndIncrementScore` corre en **una sola transacción de Postgres**:
```sql
begin;
insert into room_votes (room_id, player_id) values ($1, $2) on conflict (room_id, player_id) do nothing;
-- si no insertó ninguna fila (ya había votado), rollback y retornar Voted: false con el score actual sin tocar
update rooms set score = score + 1 where id = $1 returning score;
commit;
```
`INSERT ... ON CONFLICT DO NOTHING` reemplaza el chequeo en memoria de Polaris (`hasVotedForRoom` contra un `Set` de sesión) por la propia restricción de unicidad de la base — no hace falta ningún cache de "ya votó" del lado del proceso, la fuente de verdad es la tabla, siempre. Recién si el `INSERT` insertó una fila real se hace el `UPDATE` del score — todo en la misma transacción, así que un fallo a mitad de camino no deja el score incrementado sin el voto registrado (o viceversa), a diferencia del bug real encontrado en Polaris (Parte 1).

**Solo después de que la transacción confirma** el service muta el estado en memoria (ninguno, en este diseño — a diferencia de Polaris, Pixels no necesita un score "en memoria" separado del de Postgres, ya que `roomlive.Room` no cachea el score, se lee siempre fresco del `roommodel.Room` de vuelta) y el handler dispara el broadcast (Parte 6).

---

## Parte 4 — Comando y handler (mismo wiring que `control/commands/settings/save.go`)

```go
// internal/realm/room/control/commands/votes/cast.go

const CastName command.Name = "room.vote.cast"

// CastCommand casts one vote for the player's current room — sin payload propio,
// mismo criterio que Polaris y que REMAINING-ROOMS.md ya fijaban ("mi room actual").
type CastCommand struct {
    Handler netconn.Context
}

func (CastCommand) CommandName() command.Name { return CastName }

func (handler CastHandler) Handle(ctx context.Context, envelope command.Envelope[CastCommand]) error {
    player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
    if err != nil {
        return err
    }
    result, err := handler.Rooms.Vote(ctx, roomID, player.ID())
    if err != nil {
        return err
    }
    active, activeFound := handler.Runtime.Find(roomID)
    if activeFound {
        if err = handler.broadcastScore(ctx, active, result.Score); err != nil {
            return err
        }
    }
    if !result.Voted || handler.Events == nil {
        return nil
    }

    return handler.Events.Publish(ctx, bus.Event{Name: cast.Name, Payload: cast.Payload{RoomID: roomID, PlayerID: player.ID(), Score: result.Score}})
}
```
Nótese que `Voted: false` (ya había votado, o es el dueño) **todavía re-broadcastea el score actual** — un click que no aplica no es un error duro, mismo criterio ya usado para otros "clicks que no aplican" en el proyecto (`INTERACTIONS.md`), pero tampoco publica el evento (nada cambió de verdad).

---

## Parte 5 — Packets

| Dirección | Paquete | Contenido | Header |
| --- | --- | --- | --- |
| Inbound | `room/vote/cast` | sin campos | TBD — a confirmar contra Nitro real |
| Outbound | `room/score` | `score int32`, `canVote bool` | TBD |

`room/score` se manda dos veces en la vida de una conexión: al entrar al room (mismo momento en que hoy se manda `ROOM_MODEL`/heightmap, `access/commands/enter`) y tras cada voto exitoso de cualquier ocupante (broadcast).

---

## Parte 6 — Broadcast: `canVote` recalculado por destinatario, no un booleano compartido

```go
// broadcastScore sends the room score to every occupant, each with their own
// canVote — mismo patrón exacto confirmado en Polaris (RoomManager.voteForRoom
// recalcula hasVotedForRoom por cada Habbo del loop, no manda un valor compartido).
func (handler CastHandler) broadcastScore(ctx context.Context, active *roomlive.Room, score int) error {
    for _, occupant := range active.Occupants() {
        canVote, err := handler.Rooms.CanVote(ctx, active.ID(), occupant.PlayerID)
        if err != nil {
            continue // best-effort, mismo criterio que broadcast.RoomPacket para un ocupante puntual
        }
        packet, err := outscore.Encode(int32(score), canVote)
        if err != nil {
            return err
        }
        connection, found := handler.Connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
        if !found {
            continue
        }
        _ = connection.Send(ctx, packet)
    }

    return nil
}
```
`CanVote(ctx, roomID, playerID) (bool, error)` en `record.service`: `dueño → false`, si no `exists(select 1 from room_votes where room_id = $1 and player_id = $2)` negado. Una consulta por ocupante en el broadcast — aceptable dado que votar es una acción rara (no es un hot path, a diferencia de chat/movimiento), y el propio `broadcastScore` ya es best-effort por conexión igual que el resto de los broadcasts del proyecto.

---

## Parte 7 — Eventos

```go
// internal/realm/room/control/events/votecast/event.go
type Payload struct {
    RoomID   int64
    PlayerID int64
    Score    int
}
```
Publicado solo cuando `Voted == true` (nunca en el no-op de dueño/voto repetido) — mismo criterio que el resto de los eventos de `control/events/*`, consumible a futuro por cualquier sistema (ej. logros por "sala más votada"), sin que `votes.Service` sepa que existe ningún consumidor.

---

## Parte 8 — Hot paths y allocations

Votar **no es un hot path** — es una acción rara y deliberada del jugador, no algo que ocurra por tick ni por mensaje de chat. No hace falta ningún benchmark dedicado ni ninguna optimización de allocation más allá de lo que ya es estándar en el proyecto (reusar `broadcast.RoomPacket`-style best-effort delivery, no allocar más de lo necesario en el loop de broadcast). Se anota explícitamente para no gastar esfuerzo de optimización donde no hace falta — mismo criterio de "no construir para hipotéticos" ya aplicado en el resto de esta serie de planes.

---

## Parte 9 — Tests

- Votar suma exactamente 1 al score, persistido en Postgres, y hace broadcast a todos los ocupantes con su propio `canVote` recalculado (test de integración contra Postgres real de prueba, no un fake, para confirmar la transacción atómica).
- Votar dos veces el mismo jugador la segunda vez es un no-op: `Voted: false`, score sin cambios, **sin publicar el evento**, pero SÍ re-broadcastea el score actual (confirma el comportamiento "click que no aplica" documentado en Parte 4).
- El dueño del room nunca puede votar el suyo, incluso en su primer intento.
- Simular un fallo de conexión a Postgres a mitad de la transacción (ej. matar la conexión entre el `INSERT` y el `UPDATE` en un test con un executor fake que falla a propósito) confirma que ni el voto ni el incremento de score quedan aplicados — regresión directa contra el bug de consistencia encontrado en Polaris (Parte 1).
- `CanVote` refleja `false` para el dueño y para cualquier jugador que ya votó, `true` para el resto.
- Entrar a un room manda `room/score` con el score y `canVote` correctos para ese jugador puntual (integración con el flujo de entrada ya existente).

---

## Parte 10 — Milestones de implementación

1. **UV1 — Esquema**: migración de `room_votes`.
2. **UV2 — `record.service.Vote`/`CanVote`**: transacción atómica persistir-primero (Parte 3), corrigiendo el orden de operaciones del bug real encontrado en Polaris.
3. **UV3 — Comando, handler, packets**: `room/vote/cast` → `room/score`, mismo wiring que `control/commands/settings/save.go` (Parte 4-5).
4. **UV4 — Broadcast + evento**: `broadcastScore` con `CanVote` por destinatario (Parte 6), evento `votecast` (Parte 7).
5. **UV5 — Integración con la entrada a la sala**: mandar `room/score` en el bootstrap de `access/commands/enter`, mismo momento que hoy se manda el heightmap/modelo — depende de UV3.
