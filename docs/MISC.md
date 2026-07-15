# MISC catalog implementation

All six milestones from `plan/MISC.md` are implemented on `feature/misc`.

## Generated development snapshot

The checked-in snapshot is cross-matched against the Nitro asset pack served by
this development hotel. It contains:

- 563 wall definitions, including windows, posters, stickers, curtains, and
  other asset-safe wall decoration.
- 495 floor definitions across plants, carpets, lighting, dividers, Lodge,
  Iced, Plasto, Silo, Polyfon, Mode, Area, and three classic trophies.
- 96 poster offers backed by the single `poster` definition and distinct
  `extra_data` art ids.
- 1,257 enabled offers, including two multi-product decoration packs.
- Six pack product rows: two window products and four classic living products.

Mode/Polyfon and Area/Silo are public/technical aliases in the source data. The
catalog exposes both names deliberately and points their offers at the same
definitions; it does not duplicate inventory definitions or sprites.

The import review queue is stored in `MISC_IMPORT_REVIEW.txt`. Asset-only rows
remain unseeded until their server behavior can be verified. Excluded pets,
wired, guild, crafting, playable games, media furniture, and rentable families
remain deferred exactly as the plan requires.

## Trophy behavior

Nitro's `trophies` layout sends the textarea in catalog purchase `extraData`.
Pixels now carries that field through normal and gift packet adapters. Only a
definition with `interaction_type='trophy'` consumes it. At grant time the
server resolves the buyer, replaces tabs/newlines/control separators, applies
the immutable global word filter, truncates to 300 Unicode code points, and
persists `username\tDD-MM-YYYY\tmessage`. Other offers continue using their
server-owned catalog `extra_data`, so clients cannot inject arbitrary state.

## Database validation baseline

Liquibase schema and development seed changelogs were validated, the three MISC
changesets were rolled back, and then reapplied successfully on 2026-07-14.
Post-apply queries returned:

- 1,058 generated active definitions.
- 1,257 generated enabled offers.
- Zero generated definitions without an active offer.
- Zero duplicate generated sprite ids.
- Zero generated interactions outside `default`, `bed`, `gate`, and `trophy`.
- Two hotel-wide sanitize-list rows, both pre-existing and intentional:
  `teleport_tile` (internal paired-teleport component) and `credit_furni_50`
  (redeemable not currently sold directly).

## Required client smoke tests

1. Restart Pixels so its immutable catalog cache loads the new rows.
2. Open Shop and confirm `wall_decoration`, `plants`, `carpets`, `lighting`,
   `dividers`, `classics`, `decoration_ideas`, and `trophies` are visible.
3. Open every classics child page. Confirm Mode, Area, Lodge, Iced, Plasto,
   Silo, and Polyfon each show products and their thumbnails render.
4. Buy and place `window_basic` and `window_double_default`; move and pick them
   up, then reconnect and verify both wall positions persist.
5. Buy at least two poster variants, place them side by side, and verify their
   artwork differs according to the selected offer.
6. Buy a curtain and lamp with two states. Click each repeatedly and verify the
   client and every room occupant observe the complete state cycle.
7. Place a carpet, walk across every tile, stack one allowed item where
   applicable, and verify the avatar never floats or becomes blocked.
8. Place a classic chair, two-seat sofa, and double bed in all four rotations.
   Sit/lay on each valid slot and verify body position and rotation.
9. Purchase gold, silver, and bronze trophies with plain text, filtered text,
   accents, tabs/newlines, and more than 300 characters. Place and click each;
   verify username, date, filtered message, and truncation render correctly.
10. Gift a trophy to an online and offline user. Verify the buyer inscription
    survives wrapping/opening and uses the buyer name, not the recipient name.
11. Buy both decoration packs. Verify one charge creates every configured
    product and inventory refresh/unseen counts remain correct.
12. Run `GET /api/admin/catalog/sanitize-list` with `X-API-Key`; confirm only the
    two documented baseline rows remain.
