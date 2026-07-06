package bus

import (
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestModuleProvidesBusContracts verifies Fx can resolve bus interfaces.
func TestModuleProvidesBusContracts(t *testing.T) {
	var concrete *Bus
	var publisher Publisher
	var subscriber Subscriber

	app := fxtest.New(
		t,
		fx.Provide(zap.NewNop),
		Module,
		fx.Populate(&concrete, &publisher, &subscriber),
	)
	app.RequireStart()
	app.RequireStop()

	if concrete == nil {
		t.Fatal("expected concrete bus")
	}
	if publisher != concrete {
		t.Fatal("expected publisher to use concrete bus")
	}
	if subscriber != concrete {
		t.Fatal("expected subscriber to use concrete bus")
	}
}
