package poll

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/networking/codec"
	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
	outoffer "github.com/niflaot/pixels/networking/outbound/progression/poll/offer"
)

// Generation stores one immutable DB poll cache.
type Generation struct {
	// ByID resolves enabled definitions by wire identifier.
	ByID map[int32]Definition
	// ByRoom resolves enabled definitions by assigned room.
	ByRoom map[int64]Definition
}

// WithDatabase attaches durable poll persistence and optional badge rewards.
func (service *Service) WithDatabase(store Store, badges BadgeGranter) *Service {
	service.store, service.badges = store, badges
	return service
}

// ReloadDatabase atomically replaces the enabled DB poll cache.
func (service *Service) ReloadDatabase(ctx context.Context) error {
	if service.store == nil {
		service.generation.Store(&Generation{ByID: map[int32]Definition{}, ByRoom: map[int64]Definition{}})
		return nil
	}
	definitions, err := service.store.Polls(ctx)
	if err != nil {
		return err
	}
	generation := &Generation{ByID: make(map[int32]Definition, len(definitions)), ByRoom: make(map[int64]Definition, len(definitions))}
	for _, definition := range definitions {
		generation.ByID[definition.ID] = definition
		if definition.RoomID > 0 {
			generation.ByRoom[definition.RoomID] = definition
		}
	}
	service.generation.Store(generation)
	return nil
}

// databasePoll resolves one immutable cached definition.
func (service *Service) databasePoll(pollID int32) (Definition, bool) {
	generation := service.generation.Load()
	if generation == nil {
		return Definition{}, false
	}
	definition, found := generation.ByID[pollID]
	return definition, found
}

// DatabaseContents returns one unanswered durable poll.
func (service *Service) DatabaseContents(ctx context.Context, playerID int64, pollID int32) (codec.Packet, bool, error) {
	if service.store == nil {
		return codec.Packet{}, false, nil
	}
	definition, found := service.databasePoll(pollID)
	if !found {
		return codec.Packet{}, false, nil
	}
	completed, err := service.store.Completed(ctx, playerID, pollID)
	if err != nil {
		return codec.Packet{}, false, err
	}
	if completed {
		return codec.Packet{}, false, ErrInvalidAnswer
	}
	packet, err := outcontents.Encode(outcontents.Data{ID: definition.ID, StartMessage: definition.StartMessage, EndMessage: definition.ThanksMessage, Questions: definition.Questions})
	return packet, err == nil, err
}

// HasDatabasePoll reports whether an enabled durable poll exists.
func (service *Service) HasDatabasePoll(ctx context.Context, pollID int32) (bool, error) {
	if service.store == nil {
		return false, nil
	}
	_, found := service.databasePoll(pollID)
	return found, nil
}

// AnswerDatabase persists one durable answer and grants completion rewards once.
func (service *Service) AnswerDatabase(ctx context.Context, playerID int64, pollID int32, questionID int32, values []string) error {
	if service.store == nil || questionID <= 0 || len(values) == 0 {
		return ErrInvalidAnswer
	}
	definition, found := service.databasePoll(pollID)
	if !found || !validAnswer(definition.Questions, questionID, values) {
		return ErrInvalidAnswer
	}
	completed, badge, err := service.store.SaveAnswer(ctx, playerID, pollID, questionID, values)
	if err != nil {
		return err
	}
	if completed && badge != "" && service.badges != nil {
		_, err = service.badges.GrantBadge(ctx, playerID, badge, "poll")
	}
	if completed && err == nil {
		service.databaseResponses.Add(1)
	}
	return err
}

// validAnswer verifies one submission against immutable cached question choices.
func validAnswer(questions []outcontents.Question, questionID int32, values []string) bool {
	var selected *outcontents.Question
	for index := range questions {
		if questions[index].ID == questionID {
			selected = &questions[index]
			break
		}
		for childIndex := range questions[index].Children {
			if questions[index].Children[childIndex].ID == questionID {
				selected = &questions[index].Children[childIndex]
				break
			}
		}
	}
	if selected == nil || len(values) > 64 {
		return false
	}
	if selected.Type != 1 && selected.Type != 2 {
		return len(values) == 1 && strings.TrimSpace(values[0]) != "" && len(values[0]) <= 1024
	}
	if selected.Type == 1 && len(values) != 1 {
		return false
	}
	allowed := make(map[string]struct{}, len(selected.Choices))
	for _, choice := range selected.Choices {
		allowed[choice.Value] = struct{}{}
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := allowed[value]; !exists {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

// DatabaseResponses returns the lock-free completed response count.
func (service *Service) DatabaseResponses() uint64 { return service.databaseResponses.Load() }

// RejectDatabase persists one durable poll rejection.
func (service *Service) RejectDatabase(ctx context.Context, playerID int64, pollID int32) error {
	if service.store == nil {
		return ErrUnavailable
	}
	return service.store.RejectPoll(ctx, playerID, pollID)
}

// OfferForRoom creates an offer only for an unanswered assigned poll.
func (service *Service) OfferForRoom(ctx context.Context, playerID int64, roomID int64) (codec.Packet, bool, error) {
	if service.store == nil {
		return codec.Packet{}, false, nil
	}
	generation := service.generation.Load()
	if generation == nil {
		return codec.Packet{}, false, nil
	}
	definition, found := generation.ByRoom[roomID]
	if !found {
		return codec.Packet{}, false, nil
	}
	completed, err := service.store.Completed(ctx, playerID, definition.ID)
	if err != nil || completed {
		return codec.Packet{}, false, err
	}
	packet, err := outoffer.Encode(definition.ID, "ROOM", definition.Headline, definition.Summary)
	return packet, err == nil, err
}
