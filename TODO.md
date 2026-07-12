# TODO

Upon already coded things.

- currency/types.json must not be preserved, find another way to handle it via env
- Fix permissions of HC in room settings.
- Design and implement the complete avatar effect behavior system, including
  gender-aware furniture effects, unit runtime state, protocol projection,
  walk-on/walk-off continuity, persistence where required, and expiry rules.
  Toggle and gate furniture now publish the required footprint events, but must
  not grant partial effects until this behavior exists end to end.
- Complete room favorites. Persistence, initialization, packet definitions, and
  `navigator.favorite_changed` already exist, but commands, handlers, room
  validation, the favorite limit, packet `2524` projection, and handler
  registration are still missing for inbound add (`3817`) and remove (`309`).
  The vendored Nitro React client also has no room-favorite control, so backend
  completion alone will not expose this action in the current UI. Do not modify
  Nitro as part of the server implementation; treat client support as a separate
  compatibility decision.
