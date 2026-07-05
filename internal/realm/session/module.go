package session

import (
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"go.uber.org/fx"
)

// Module provides session realm runtime state.
var Module = fx.Module(
	"realm-session",
	fx.Provide(binding.NewRegistry),
)
