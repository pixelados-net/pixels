package i18n

import "go.uber.org/fx"

// Module provides i18n catalog dependencies.
var Module = fx.Module(
	"i18n",
	fx.Provide(
		LoadCatalog,
		NewTranslator,
	),
)
