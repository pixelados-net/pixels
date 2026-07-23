# Room Ads

Second page of the Navigator section. Room ads (promotions) are the navigator family's one paid feature, and a single wire mechanism serves two UI surfaces: the catalog page where an owner buys a promotion, and the in-room event banner every occupant sees. Both are views over the same durable record, implemented in `internal/realm/room/promotion`.

## The flow

**1. Purchase info.** From the catalog's Room Ads page, the client asks which rooms the actor may promote. Pixels answers with rooms where the player is the owner, or holds the promotion manage-any permission node, which is how staff promote on behalf of others.

**2. Purchase.** The client sends the catalog page and offer being bought, plus the title, description, and category the owner typed:

```go
case inpurchase.Header:
	value, err := inpurchase.Decode(packet)
	input.Action = PurchaseAd
	input.Purchase = PurchaseParams{RoomID: …, PageID: …, OfferID: …,
		Title: value.Title, Description: value.Description,
		CategoryID: value.CategoryID, Extended: value.Extended}
```

The service then validates in order: the actor's rights over the room, the copy (length-limited, filtered like any player-authored text), and, critically, that the offer genuinely belongs to a catalog page using the Room Ads layout, so no other product can be replayed into a promotion purchase. Payment goes through the regular catalog purchase flow, and the promotion row is created **in the same transaction as the charge**. No charge without a promotion, no promotion without a charge; a duplicate purchase while one is active extends the existing promotion instead of stacking a second one.

**3. Projection.** The event banner (title, description, category, minutes since start, minutes remaining) is pushed to everyone currently in the room and to each player entering while the promotion runs. The `roomads_view` navigator tab lists rooms whose promotion is still active.

**4. Edit.** The owner can rename and re-describe the live promotion from the banner without paying again, same copy validation, no new charge.

**5. Cancel.** Cancelling ends the promotion and clears the banner for everyone. Worth a historical note: the classic reference servers never implemented cancellation at all, even though the client can ask for it; Pixels implements it because the packet pair exists and the semantics are unambiguous.

## Design details worth knowing

**Expiry is computed, not ticked.** A promotion stores its end timestamp; the banner and the navigator tab both compare against the clock when asked. There is no background job decrementing minutes, nothing to fall behind, and a promotion "ends" simultaneously everywhere by definition.

**Telemetry packets are accepted and dropped.** The client emits several instrumentation packets around the Room Ads catalog page (tab clicked, tab viewed, purchase initiated, ad search). They decode and dispatch as explicit `Telemetry` actions that do nothing further: accepted so the client is never confused, ignored because an emulator has no ad-metrics pipeline to feed.

**Configuration.** Promotion duration and extension length are config values with defaults matching classic behavior (about two hours per purchase), and the copy limits live beside them in the promotion package's config. See [[CONFIGURATION]] for the pattern.

## Administration

For moderation cases (a promotion with copy that slipped through filters), operators can inspect and force-cancel through the admin API, with the standard acting staff authorization and audit trail described in [[USERS-PERMISSIONS]].
