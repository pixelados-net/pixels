package guide

import (
	"sync"
	"sync/atomic"
	"time"

	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
)

// Manager owns guide duty and ephemeral sessions.
type Manager struct {
	// mutex protects duty and session indexes.
	mutex sync.RWMutex
	// duty stores guide availability.
	duty map[int64]Duty
	// sessions stores sessions by id.
	sessions map[int64]*Session
	// byPlayer indexes either participant.
	byPlayer map[int64]int64
	// completed stores one feedback-eligible session per requester.
	completed map[int64]Session
	// nextID generates runtime identities.
	nextID atomic.Int64
	// filter applies the hotel dictionary.
	filter *chatfilter.Service
	// now supplies deterministic timestamps.
	now func() time.Time
}

// New creates an empty guide manager.
func New(filter *chatfilter.Service) *Manager {
	return &Manager{duty: make(map[int64]Duty), sessions: make(map[int64]*Session), byPlayer: make(map[int64]int64), completed: make(map[int64]Session), filter: filter, now: time.Now}
}

// SetDuty replaces one guide's queue participation.
func (manager *Manager) SetDuty(playerID int64, guide bool, bully bool, guardian bool) Duty {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if !guide && !bully && !guardian {
		delete(manager.duty, playerID)
		return Duty{PlayerID: playerID}
	}
	value, found := manager.duty[playerID]
	if !found {
		value = Duty{PlayerID: playerID, Since: manager.now()}
	}
	value.Guide, value.Bully, value.Guardian = guide, bully, guardian
	manager.duty[playerID] = value
	return value
}

// DutyCount returns current normal, bully, and guardian pool sizes.
func (manager *Manager) DutyCount() (int32, int32, int32) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	var guides, bullies, guardians int32
	for id, duty := range manager.duty {
		if _, busy := manager.byPlayer[id]; busy {
			continue
		}
		if duty.Guide {
			guides++
		}
		if duty.Bully {
			bullies++
		}
		if duty.Guardian {
			guardians++
		}
	}
	return guides, bullies, guardians
}

// Create matches the oldest available guide.
func (manager *Manager) Create(requesterID int64, topic int32, description string) (Session, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if _, busy := manager.byPlayer[requesterID]; busy {
		return Session{}, ErrBusy
	}
	guideID := manager.oldestGuide(requesterID, 0)
	if guideID == 0 {
		return Session{}, ErrUnavailable
	}
	id := manager.nextID.Add(1)
	session := &Session{ID: id, RequesterPlayerID: requesterID, GuidePlayerID: guideID, Topic: topic, Description: manager.clean(description), State: StateAttached, CreatedAt: manager.now(), Transcript: make([]Message, 0, 16)}
	manager.sessions[id], manager.byPlayer[requesterID], manager.byPlayer[guideID] = session, id, id
	return clone(session), nil
}

// Decide accepts or rejects an attached request.
func (manager *Manager) Decide(guideID int64, accepted bool) (Session, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	session := manager.sessionFor(guideID)
	if session == nil || session.GuidePlayerID != guideID {
		return Session{}, ErrUnauthorized
	}
	if session.State != StateAttached {
		return Session{}, ErrInvalidState
	}
	if accepted {
		session.State = StateStarted
		return clone(session), nil
	}
	requesterID := session.RequesterPlayerID
	manager.remove(session)
	next := manager.oldestGuide(guideID, requesterID)
	if next == 0 {
		return clone(session), ErrUnavailable
	}
	session.GuidePlayerID, session.State = next, StateAttached
	manager.sessions[session.ID], manager.byPlayer[requesterID], manager.byPlayer[next] = session, session.ID, session.ID
	return clone(session), nil
}

// Send appends one bounded filtered transcript message.
func (manager *Manager) Send(playerID int64, text string) (Session, Message, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	session := manager.sessionFor(playerID)
	if session == nil {
		return Session{}, Message{}, ErrUnavailable
	}
	if session.State != StateStarted {
		return Session{}, Message{}, ErrInvalidState
	}
	message := Message{SenderPlayerID: playerID, Text: manager.clean(text), CreatedAt: manager.now()}
	if message.Text == "" {
		return clone(session), message, nil
	}
	if len(session.Transcript) == 200 {
		copy(session.Transcript, session.Transcript[1:])
		session.Transcript = session.Transcript[:199]
	}
	session.Transcript = append(session.Transcript, message)
	return clone(session), message, nil
}

// End completes or detaches one participant's session.
func (manager *Manager) End(playerID int64) (Session, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	session := manager.sessionFor(playerID)
	if session == nil {
		return Session{}, false
	}
	session.State = StateEnded
	value := clone(session)
	manager.completed[session.RequesterPlayerID] = value
	manager.remove(session)
	return value, true
}

// TakeCompleted consumes one feedback-eligible completed session.
func (manager *Manager) TakeCompleted(requesterID int64) (Session, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	value, found := manager.completed[requesterID]
	if found {
		delete(manager.completed, requesterID)
	}
	return value, found
}

// SessionFor returns one participant's detached session snapshot.
func (manager *Manager) SessionFor(playerID int64) (Session, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	session := manager.sessionFor(playerID)
	if session == nil {
		return Session{}, false
	}
	return clone(session), true
}

// oldestGuide returns the first idle normal guide.
func (manager *Manager) oldestGuide(excludeID int64, secondExcludeID int64) int64 {
	var selected Duty
	for id, duty := range manager.duty {
		if id == excludeID || id == secondExcludeID || !duty.Guide {
			continue
		}
		if _, busy := manager.byPlayer[id]; busy {
			continue
		}
		if selected.PlayerID == 0 || duty.Since.Before(selected.Since) || duty.Since.Equal(selected.Since) && duty.PlayerID < selected.PlayerID {
			selected = duty
		}
	}
	return selected.PlayerID
}

// sessionFor resolves one indexed participant.
func (manager *Manager) sessionFor(playerID int64) *Session {
	return manager.sessions[manager.byPlayer[playerID]]
}

// remove clears all session indexes.
func (manager *Manager) remove(session *Session) {
	delete(manager.sessions, session.ID)
	delete(manager.byPlayer, session.RequesterPlayerID)
	delete(manager.byPlayer, session.GuidePlayerID)
}
