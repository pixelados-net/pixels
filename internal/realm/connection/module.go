package connection

import "go.uber.org/fx"

// Module provides connection-realm handlers.
var Module = fx.Module(
	"realm-connection",
	fx.Provide(NewHandlersWithPermissions),
	fx.Invoke(
		RegisterSecurityTranslations,
		RegisterEffectHandlers,
		RegisterAchievementHandlers,
		RegisterProtocolCompatibilityHandlers,
		RegisterPlayerCompatibilityHandlers,
		RegisterPlayerSettingsHandlers,
		RegisterPlayerProfileHandlers,
		RegisterPlayerWardrobeHandlers,
		RegisterPlayerIdentityHandlers,
	),
)
