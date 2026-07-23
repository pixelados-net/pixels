package behavior

import sdkbot "github.com/niflaot/pixels/sdk/bot"

// RegisterBuiltins registers every core behavior exactly once.
func RegisterBuiltins(registry *Registry) error {
	if err := registry.Register("generic", func() sdkbot.Behavior { return Generic{} }); err != nil {
		return err
	}
	if err := registry.Register("bartender", func() sdkbot.Behavior { return Bartender{} }); err != nil {
		return err
	}
	return registry.Register("visitor_log", func() sdkbot.Behavior { return VisitorLog{} })
}
