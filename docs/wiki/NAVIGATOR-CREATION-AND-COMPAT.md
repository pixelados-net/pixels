# Room Creation and Navigator Compatibility

Third page of the Navigator section: how new rooms are born, and which navigator packets are deliberately answered with empty results, with the reasoning, since "deliberately empty" is a design decision worth defending.

## Creating a room

Room creation flows through the navigator realm (`internal/realm/navigator/create`), because it starts from the navigator's "create room" dialog:

1. **Name check.** The client can pre-validate the name; the server applies the same length and filtering rules it will enforce at creation, so the dialog's feedback matches reality.
2. **Creation.** The request carries name, description, the chosen layout model from the layout catalog, initial door mode, category, and visitor limit. Validation reuses the same primitives as room settings, one set of rules whether you're creating or editing later.
3. **Handoff.** The durable room record is created by the room realm's record service (the navigator only orchestrates), and the client is forwarded straight into the new room, arriving through the normal entry flow described in [[ROOMS-ENTRY]].

Layouts come from the room realm's layout catalog, the same models the floor plan editor can later modify, so a "created" room and an "edited" room are never different kinds of things.

## The deliberately empty corners

Two families of navigator packets answer with valid-but-empty results on purpose. Both decisions came from auditing the real client rather than assuming, and both are documented in code comments at their handlers.

**Legacy named searches.** Old protocol revisions had one packet per search: "my favourite rooms search", "guild base search", "rooms where my friends are", and a dozen more. The modern client collapsed all of them into the single `NAVIGATOR_SEARCH` packet with a code (see [[NAVIGATOR-BROWSING]]), and an exhaustive audit of the shipped client confirmed none of the legacy headers are ever sent. The composer classes exist in the client's engine, but nothing constructs them, and one (`guild base search`) is structurally bypassed because the client visits a group's home room directly from the group panel without touching the navigator at all. Pixels decodes every one of these headers and returns an empty result list so nothing can hang, while the semantics they used to name are already served by the modern tabs: favorites through `myworld_view`, popular through `hotel_view`, official through `official_view`, free text through the query fallback.

**Competition searches.** The protocol specifies searches over competition-entry rooms, plus forwards into random competition rooms. There is no competition system with a client-side surface to feed these. The audit of the client found parsers but no UI capable of driving the flow, so they answer a valid empty page. If a real competition feature ever lands, these packets are its ready-made wire entry points; until then, an empty page is the correct answer.

The underlying principle is the one stated on [[Home]]: the wire contract is always honored, so traffic inspection never meets a mystery, but behavior exists only where the real client can actually reach it. An emulator that invents behavior for unreachable packets isn't more complete. It's less trustworthy, because none of that behavior can ever be exercised or verified from the actual game.

## What this means when extending the navigator

If you're adding a navigator feature, the checklist that follows from this page:

- New search semantics belong behind a **code** routed through the existing search dispatch, not behind a new packet.
- If the feature is personal-list shaped (favorites-like), it's a `myworld_view` sub-list plus card state.
- Before implementing any currently-empty packet "for completeness," find the client call site that would send it. If there isn't one, the empty answer is already correct.
