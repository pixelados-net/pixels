package i18n

import (
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestModuleProvidesTranslator verifies Fx wiring.
func TestModuleProvidesTranslator(t *testing.T) {
	var invoked bool
	app := fxtest.New(
		t,
		Module,
		fx.Provide(
			func() Config {
				return Config{Path: filepathForMissingTest(t)}
			},
			zap.NewNop,
		),
		fx.Invoke(func(translator Translator) {
			invoked = translator != nil
		}),
	)

	app.RequireStart()
	app.RequireStop()

	if !invoked {
		t.Fatal("expected translator invocation")
	}
}

// filepathForMissingTest returns a missing catalog path.
func filepathForMissingTest(t *testing.T) string {
	t.Helper()

	return t.TempDir() + "/missing.json"
}
