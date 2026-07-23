# Navigator Browsing

First page of the Navigator section: the tabbed browser, the one search packet everything rides on, categories, favorites, history, and saved searches. [[NAVIGATOR-ROOM-ADS]] covers promotions; [[NAVIGATOR-CREATION-AND-COMPAT]] covers room creation and the deliberately empty corners.

## The four tabs

When a client opens the navigator, Pixels sends the top-level context metadata, exactly four tabs, matching what the Nitro UI renders:

```go
// metadataContexts returns top-level Nitro navigator contexts.
return []outmetadata.Context{
	{Code: "official_view", SavedSearches: …},
	{Code: "hotel_view",    SavedSearches: …},
	{Code: "roomads_view",  SavedSearches: …},
	{Code: "myworld_view",  SavedSearches: …},
}
```

| Tab | What it serves |
|---|---|
| `official_view` | Staff-curated rooms, in thumbnail mode |
| `hotel_view` | Popular rooms by live occupancy, or free-text results when the player typed a query |
| `roomads_view` | Rooms with an active promotion (see [[NAVIGATOR-ROOM-ADS]]) |
| `myworld_view` | Personal sub-lists: your rooms, your favorites |

## One search packet to rule them all

Every search afterwards is a single generic packet, `NAVIGATOR_SEARCH`, carrying a `code` (which tab or sub-list) and a `data` string (the typed query, if any). The client never hardcodes category codes; it echoes back codes the server previously sent, either a tab code from the metadata above or a sub-list code from a previous result. Server-side, one dispatch point in `internal/realm/navigator/browse/search` routes codes to result builders:

```go
case "favorites":       favoriteLists(…)
case "history":         historyLists(ctx, playerID, false)
case "frequent":        historyLists(ctx, playerID, true)
case "official_view":   officialLists(…)
case "hotel_view":      hotelLists(ctx, query)
default:                queryLists(ctx, code, query) // free-text fallback
```

Result lists carry a display mode (list versus thumbnail grid), an action flag (whether "see more" is offered), and room cards. Cards are enriched in one bounded batch with social-group badges for rooms owned by groups, a deliberate single query, not one lookup per card.

**Official rooms** are simply rooms with the `StaffPicked` flag set. Staff toggle it from the room info panel in-client, and operators can do the same through `POST/DELETE /api/admin/navigator/official/:roomId`. There's no separate curation table to maintain.

**Popular rooms** rank by live occupancy read from the room runtime registry (see [[ROOMS-RUNTIME]]). The count is the number of units actually standing in each room right now, not a cached statistic.

## Favorites, history, saved searches

- **Favorites** are durable per-player rows with a capped count, toggled by the star button. They surface twice: as a `myworld_view` sub-list and as state on every room card.
- **History** tracks recently visited and frequently visited rooms per player, fed by room entry events rather than by the navigator itself. Entering a room is what writes history; opening the navigator only reads it.
- **Saved searches** are the bookmarks under the search bar, stored per player and echoed back inside the tab metadata, each bound to the tab it was saved from.

## Categories

Rooms carry a flat category chosen at creation or in room settings. The navigator exposes the category list and per-category user counts through dedicated request packets, and per-player category preferences (which sections are collapsed) persist across sessions. Categories are data: they live in navigator-owned tables and seeds, so adding one is an insert, not a deploy.
