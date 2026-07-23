package guardian

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/pkg/redis"
)

var (
	// ErrUnavailable reports insufficient reviewers or missing ticket.
	ErrUnavailable = errors.New("guardian review unavailable")
	// ErrInvalidState reports an action outside its review phase.
	ErrInvalidState = errors.New("invalid guardian review state")
	// ErrUnauthorized reports a player outside the reviewer set.
	ErrUnauthorized = errors.New("guardian review unauthorized")
)

// Pool supplies currently eligible guardian ids.
type Pool interface {
	// Guardians returns bounded available guardian ids.
	Guardians(int64, int) []int64
}

// Manager owns review sessions and vote aggregation.
type Manager struct {
	// mutex protects ticket and player indexes.
	mutex sync.RWMutex
	// config stores reviewer count and vote window.
	config moderationconfig.Config
	// pool supplies guardian candidates.
	pool Pool
	// redis stores ignored-offer counters and temporary exclusions.
	redis *redis.Client
	// store persists tickets, votes, and final results.
	store moderationrecord.Store
	// tickets stores active and recently closed reviews.
	tickets map[int64]*Ticket
	// byPlayer indexes reviewers and reporters.
	byPlayer map[int64]int64
	// nextID generates runtime identities.
	nextID atomic.Int64
	// now supplies deterministic timestamps.
	now func() time.Time
}

// New creates an in-memory guardian review manager for focused use.
func New(config moderationconfig.Config, pool Pool) *Manager {
	return newManager(config, pool, nil, nil)
}

// NewPersistent creates a guardian manager with distributed exclusions and durable audit.
func NewPersistent(config moderationconfig.Config, pool Pool, redisClient *redis.Client, store moderationrecord.Store) *Manager {
	return newManager(config, pool, redisClient, store)
}

// newManager composes guardian review dependencies.
func newManager(config moderationconfig.Config, pool Pool, redisClient *redis.Client, store moderationrecord.Store) *Manager {
	return &Manager{config: config, pool: pool, redis: redisClient, store: store, tickets: make(map[int64]*Ticket), byPlayer: make(map[int64]int64), now: time.Now}
}

// Create freezes anonymized evidence and offers it to guardians.
func (manager *Manager) Create(ctx context.Context, reporterID int64, reportedID int64, chatlog []moderationrecord.ChatEntry) (Ticket, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if manager.byPlayer[reporterID] != 0 {
		return Ticket{}, ErrInvalidState
	}
	ids := manager.eligible(ctx, reporterID, reportedID)
	if len(ids) < manager.config.GuardianCount {
		return Ticket{}, ErrUnavailable
	}
	now := manager.now()
	id := manager.nextID.Add(1)
	if manager.store != nil {
		persistedID, err := manager.store.CreateGuardianTicket(ctx, reporterID, reportedID, now.Add(manager.config.GuardianVoteWindow))
		if err != nil {
			return Ticket{}, err
		}
		id = persistedID
	}
	ticket := &Ticket{ID: id, ReporterPlayerID: reporterID, ReportedPlayerID: reportedID, State: StateOffered, CreatedAt: now, ClosesAt: now.Add(manager.config.GuardianVoteWindow), Reviewers: make(map[int64]*Reviewer, len(ids)), Chatlog: anonymize(chatlog)}
	for _, playerID := range ids {
		ticket.Reviewers[playerID] = &Reviewer{PlayerID: playerID}
		manager.byPlayer[playerID] = id
	}
	manager.tickets[id], manager.byPlayer[reporterID] = ticket, id
	return clone(ticket), nil
}

// Decide records acceptance and starts voting after every explicit decision.
func (manager *Manager) Decide(playerID int64, accepted bool) (Ticket, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	ticket := manager.ticketFor(playerID)
	if ticket == nil || ticket.State != StateOffered {
		return Ticket{}, ErrInvalidState
	}
	reviewer := ticket.Reviewers[playerID]
	if reviewer == nil {
		return Ticket{}, ErrUnauthorized
	}
	reviewer.Accepted, reviewer.Decided = accepted, true
	all := true
	acceptedCount := 0
	for _, candidate := range ticket.Reviewers {
		if !candidate.Decided {
			all = false
		}
		if candidate.Accepted {
			acceptedCount++
		}
	}
	if all {
		if acceptedCount == 0 {
			manager.close(context.Background(), ticket)
		} else {
			ticket.State, ticket.ClosesAt = StateVoting, manager.now().Add(manager.config.GuardianVoteWindow)
		}
	}
	return clone(ticket), nil
}

// Vote records one accepted guardian vote and closes on complete quorum.
func (manager *Manager) Vote(playerID int64, verdict Verdict) (Ticket, bool, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if verdict < VerdictAcceptable || verdict > VerdictHorrible {
		return Ticket{}, false, ErrInvalidState
	}
	ticket := manager.ticketFor(playerID)
	if ticket == nil || ticket.State != StateVoting {
		return Ticket{}, false, ErrInvalidState
	}
	reviewer := ticket.Reviewers[playerID]
	if reviewer == nil || !reviewer.Accepted {
		return Ticket{}, false, ErrUnauthorized
	}
	copy := verdict
	reviewer.Vote = &copy
	if manager.store != nil {
		if err := manager.store.SaveGuardianVote(context.Background(), ticket.ID, playerID, int32(verdict)); err != nil {
			reviewer.Vote = nil
			return Ticket{}, false, err
		}
	}
	complete := true
	for _, candidate := range ticket.Reviewers {
		if candidate.Accepted && candidate.Vote == nil {
			complete = false
		}
	}
	if complete {
		manager.close(context.Background(), ticket)
	}
	return clone(ticket), complete, nil
}

// Tick closes expired voting tickets and returns detached results.
func (manager *Manager) Tick(ctx context.Context, now time.Time) []Ticket {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	closed := make([]Ticket, 0)
	for _, ticket := range manager.tickets {
		if ticket.State != StateClosed && !ticket.ClosesAt.After(now) {
			if ticket.State == StateOffered {
				for playerID, reviewer := range ticket.Reviewers {
					if !reviewer.Decided {
						manager.recordIgnored(ctx, playerID)
					}
				}
			}
			manager.close(ctx, ticket)
			closed = append(closed, clone(ticket))
		}
	}
	return closed
}

// Detach removes one guardian from an active review.
func (manager *Manager) Detach(playerID int64) (Ticket, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	ticket := manager.ticketFor(playerID)
	if ticket == nil {
		return Ticket{}, false
	}
	delete(ticket.Reviewers, playerID)
	delete(manager.byPlayer, playerID)
	if len(ticket.Reviewers) == 0 {
		manager.close(context.Background(), ticket)
	} else if ticket.State == StateVoting && votingComplete(ticket) {
		manager.close(context.Background(), ticket)
	}
	return clone(ticket), true
}

// TicketFor returns one participant's detached review.
func (manager *Manager) TicketFor(playerID int64) (Ticket, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	value := manager.ticketFor(playerID)
	if value == nil {
		return Ticket{}, false
	}
	return clone(value), true
}

// close aggregates a strict majority or mixed result.
func (manager *Manager) close(ctx context.Context, ticket *Ticket) {
	counts := [3]int{}
	total := 0
	for _, reviewer := range ticket.Reviewers {
		if reviewer.Accepted && reviewer.Vote != nil {
			counts[int(*reviewer.Vote)]++
			total++
		}
		delete(manager.byPlayer, reviewer.PlayerID)
	}
	ticket.Result, ticket.State = VerdictMixed, StateClosed
	for index, count := range counts {
		if count > total/2 {
			ticket.Result = Verdict(index)
			break
		}
	}
	delete(manager.byPlayer, ticket.ReporterPlayerID)
	delete(manager.tickets, ticket.ID)
	if manager.store != nil {
		_ = manager.store.CloseGuardianTicket(ctx, ticket.ID, int32(ticket.Result))
	}
}

// ticketFor resolves one player index.
func (manager *Manager) ticketFor(playerID int64) *Ticket {
	return manager.tickets[manager.byPlayer[playerID]]
}
