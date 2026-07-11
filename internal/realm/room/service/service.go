package service

import (
	"github.com/niflaot/pixels/internal/realm/room/layout"
	"github.com/niflaot/pixels/internal/realm/room/repository"
)

// Service validates and coordinates room persistence behavior.
type Service struct {
	// store reads and writes room persistence records.
	store repository.Store

	// layouts validates room layout references.
	layouts layout.Manager

	// profanity validates user-facing room text when configured.
	profanity ProfanityChecker
}

// New creates a room service.
func New(store repository.Store, layouts layout.Manager) *Service {
	return &Service{store: store, layouts: layouts}
}

// WithProfanity configures optional global content validation.
func (service *Service) WithProfanity(checker ProfanityChecker) *Service {
	service.profanity = checker

	return service
}
