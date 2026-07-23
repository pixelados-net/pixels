package storage

import "go.uber.org/fx"

// Module provides S3-compatible storage configuration and behavior.
var Module = fx.Module("storage", fx.Provide(LoadConfig, New))
