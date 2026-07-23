// Package poll owns bounded room word-quiz sessions.
package poll

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outanswered "github.com/niflaot/pixels/networking/outbound/progression/poll/answered"
	outfinished "github.com/niflaot/pixels/networking/outbound/progression/poll/finished"
	outquestion "github.com/niflaot/pixels/networking/outbound/progression/poll/question"
)

var (
	// ErrUnavailable reports a missing room or active poll.
	ErrUnavailable = errors.New("word poll unavailable")
	// ErrForbidden reports a launch without room rights.
	ErrForbidden = errors.New("word poll forbidden")
	// ErrInvalidAnswer reports an unsupported or duplicate vote.
	ErrInvalidAnswer = errors.New("word poll invalid answer")
)

// Broadcaster projects packets to room occupants.
type Broadcaster func(context.Context, *roomlive.Room, codec.Packet, int64) error

// active stores one bounded live poll.
type active struct {
	// id identifies the poll on the wire.
	id int32
	// room stores the active audience.
	room *roomlive.Room
	// question stores display content.
	question string
	// duration stores widget visibility time.
	duration time.Duration
	// answers stores one value per player.
	answers map[int64]string
	// yes stores affirmative votes.
	yes int32
	// no stores negative votes.
	no int32
	// timer closes the poll.
	timer *time.Timer
}

// Service coordinates live room word polls.
type Service struct {
	// rooms resolves active rooms and occupants.
	rooms *roomlive.Registry
	// broadcast sends room-wide packets.
	broadcast Broadcaster
	// next generates process-local poll identifiers.
	next atomic.Int32
	// mutex protects active polls.
	mutex sync.Mutex
	// activeByRoom stores at most one poll per room.
	activeByRoom map[int64]*active
	// activeByID resolves client requests without scanning rooms.
	activeByID map[int32]*active
	// store persists optional DB-backed polls.
	store Store
	// badges grants optional completion rewards.
	badges BadgeGranter
	// generation stores immutable enabled DB polls by id and room.
	generation atomic.Pointer[Generation]
	// databaseResponses counts completed durable poll responses.
	databaseResponses atomic.Uint64
}

// New creates an empty word-poll runtime.
func New(rooms *roomlive.Registry, broadcast Broadcaster) *Service {
	service := &Service{rooms: rooms, broadcast: broadcast, activeByRoom: make(map[int64]*active), activeByID: make(map[int32]*active)}
	return service
}

// Start launches one word poll after checking live room rights.
func (service *Service) Start(ctx context.Context, roomID int64, actorID int64, question string, duration time.Duration) (int32, error) {
	room, found := service.rooms.Find(roomID)
	if !found || question == "" || duration <= 0 {
		return 0, ErrUnavailable
	}
	if !room.HasRights(actorID) {
		return 0, ErrForbidden
	}
	poll := &active{id: service.next.Add(1), room: room, question: question, duration: duration, answers: make(map[int64]string)}
	service.mutex.Lock()
	if previous := service.activeByRoom[roomID]; previous != nil {
		service.removeLocked(previous)
	}
	service.activeByRoom[roomID] = poll
	service.activeByID[poll.id] = poll
	poll.timer = time.AfterFunc(duration, func() { _ = service.Finish(context.Background(), poll.id) })
	service.mutex.Unlock()
	packet, err := questionPacket(poll)
	if err != nil {
		service.Cancel(poll.id)
		return 0, err
	}
	if err = service.broadcast(ctx, room, packet, 0); err != nil {
		service.Cancel(poll.id)
		return 0, err
	}
	return poll.id, nil
}

// Cancel removes one live poll without broadcasting a final result.
func (service *Service) Cancel(pollID int32) {
	service.mutex.Lock()
	if poll := service.activeByID[pollID]; poll != nil {
		service.removeLocked(poll)
	}
	service.mutex.Unlock()
}

// Current returns the active poll requested by an occupant.
func (service *Service) Current(playerID int64, pollID int32) (codec.Packet, bool, error) {
	service.mutex.Lock()
	poll := service.activeByID[pollID]
	service.mutex.Unlock()
	if poll == nil {
		return codec.Packet{}, false, nil
	}
	if _, found := poll.room.Occupant(playerID); !found {
		return codec.Packet{}, false, nil
	}
	packet, err := questionPacket(poll)
	return packet, err == nil, err
}

// Answer records and broadcasts one occupant vote exactly once.
func (service *Service) Answer(ctx context.Context, playerID int64, pollID int32, questionID int32, values []string) error {
	if questionID != -1 || len(values) != 1 || values[0] != "0" && values[0] != "1" {
		return ErrInvalidAnswer
	}
	service.mutex.Lock()
	poll := service.activeByID[pollID]
	if poll == nil {
		service.mutex.Unlock()
		return ErrUnavailable
	}
	_, found := poll.room.Occupant(playerID)
	if !found {
		service.mutex.Unlock()
		return ErrForbidden
	}
	if _, answered := poll.answers[playerID]; answered {
		service.mutex.Unlock()
		return ErrInvalidAnswer
	}
	poll.answers[playerID] = values[0]
	if values[0] == "1" {
		poll.yes++
	} else {
		poll.no++
	}
	counts := pollCounts(poll)
	room := poll.room
	service.mutex.Unlock()
	packet, err := outanswered.Encode(int32(playerID), values[0], counts)
	if err != nil {
		return err
	}
	return service.broadcast(ctx, room, packet, 0)
}

// Reject records no answer while keeping the shared poll active.
func (service *Service) Reject(playerID int64, pollID int32) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	poll := service.activeByID[pollID]
	if poll == nil {
		return ErrUnavailable
	}
	if _, found := poll.room.Occupant(playerID); !found {
		return ErrForbidden
	}
	if _, answered := poll.answers[playerID]; answered {
		return ErrInvalidAnswer
	}
	poll.answers[playerID] = ""
	return nil
}

// Finish closes and broadcasts final aggregate results idempotently.
func (service *Service) Finish(ctx context.Context, pollID int32) error {
	service.mutex.Lock()
	poll := service.activeByID[pollID]
	if poll == nil {
		service.mutex.Unlock()
		return nil
	}
	counts := pollCounts(poll)
	room := poll.room
	service.removeLocked(poll)
	service.mutex.Unlock()
	packet, err := outfinished.Encode(-1, counts)
	if err != nil {
		return err
	}
	return service.broadcast(ctx, room, packet, 0)
}

// Close cancels every live poll timer.
func (service *Service) Close() {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	for _, poll := range service.activeByRoom {
		service.removeLocked(poll)
	}
}

// removeLocked removes one poll while the service mutex is held.
func (service *Service) removeLocked(poll *active) {
	if poll.timer != nil {
		poll.timer.Stop()
	}
	delete(service.activeByID, poll.id)
	delete(service.activeByRoom, poll.room.Snapshot().ID)
}

// questionPacket encodes one quick word-quiz question.
func questionPacket(poll *active) (codec.Packet, error) {
	return outquestion.Encode(outquestion.Data{PollType: poll.question, PollID: poll.id, QuestionID: -1, DurationSeconds: int32(poll.duration.Milliseconds()), ID: -1, Type: 3, Content: poll.question})
}

// pollCounts returns stable yes and no aggregates.
func pollCounts(poll *active) []outanswered.Count {
	return []outanswered.Count{{Value: "0", Count: poll.no}, {Value: "1", Count: poll.yes}}
}
