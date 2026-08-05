package hub

import (
	"sync"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

// RoomActor owns one game.Room and serializes every mutation through a single
// goroutine, so the underlying Room never needs its own locking. It also
// drives the wall-clock timers that turn "time ran out" and "reveal is done,
// show the next question" into actual state transitions, and maps
// reconnect tokens back to stable PlayerIDs so a dropped WebSocket doesn't
// cost a player their score or streak.
type RoomActor struct {
	room        *game.Room
	revealDelay time.Duration

	cmds chan func(now time.Time) []game.Event
	stop chan struct{}
	done chan struct{}

	subMu       sync.Mutex
	subscribers map[chan game.Event]struct{}

	tokenMu sync.Mutex
	tokens  map[string]game.PlayerID

	sourceMu sync.Mutex
	source   QuestionSource
}

// QuestionSource records which slice of the question bank a room was built
// from, so "play again" draws a fresh set from the same language and
// category instead of silently switching the game out from under everyone.
type QuestionSource struct {
	Lang     string
	Category string // empty means a mixed round
}

// SetSource records the room's question source. Called once, right after the
// room is created.
func (a *RoomActor) SetSource(src QuestionSource) {
	a.sourceMu.Lock()
	defer a.sourceMu.Unlock()
	a.source = src
}

// Source returns the room's recorded question source.
func (a *RoomActor) Source() QuestionSource {
	a.sourceMu.Lock()
	defer a.sourceMu.Unlock()
	return a.source
}

// NewRoomActor creates an actor for the given players and question set.
// revealDelay controls how long the reveal phase is shown before the actor
// automatically advances to the next question (or finishes the game).
func NewRoomActor(players []game.Player, questions []game.Question, revealDelay time.Duration) *RoomActor {
	a := &RoomActor{
		room:        game.NewRoom(players, questions),
		revealDelay: revealDelay,
		cmds:        make(chan func(now time.Time) []game.Event),
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
		subscribers: make(map[chan game.Event]struct{}),
		tokens:      make(map[string]game.PlayerID),
	}
	go a.loop()
	return a
}

// Start begins the game, publishing the first QuestionStarted event.
func (a *RoomActor) Start() {
	a.exec(func(now time.Time) []game.Event {
		return a.room.Start(now)
	})
}

// SubmitAnswer records a player's answer for the current question.
func (a *RoomActor) SubmitAnswer(id game.PlayerID, choice int) {
	a.exec(func(now time.Time) []game.Event {
		return a.room.SubmitAnswer(id, choice, now)
	})
}

// Roster returns a snapshot of every currently enrolled player, so a newly
// connecting client can render everyone who joined before it did.
func (a *RoomActor) Roster() []game.Player {
	var roster []game.Player
	a.exec(func(now time.Time) []game.Event {
		roster = a.room.Players()
		return nil
	})
	return roster
}

// AddPlayer enrolls a new player while the room is still in its lobby.
func (a *RoomActor) AddPlayer(id game.PlayerID, name, avatar string) error {
	var err error
	a.exec(func(now time.Time) []game.Event {
		var events []game.Event
		events, err = a.room.AddPlayer(id, name, avatar)
		return events
	})
	return err
}

// Reset restarts the room in place (same roster, same actor, same join
// code) with a fresh question set, for a "play again" flow.
func (a *RoomActor) Reset(questions []game.Question) {
	a.exec(func(now time.Time) []game.Event {
		return a.room.Reset(questions)
	})
}

// exec runs fn on the actor's goroutine and blocks until it has completed,
// so callers observe a consistent room state immediately after this returns.
func (a *RoomActor) exec(fn func(now time.Time) []game.Event) {
	ack := make(chan struct{})
	select {
	case a.cmds <- func(now time.Time) []game.Event {
		events := fn(now)
		close(ack)
		return events
	}:
		<-ack
	case <-a.done:
	}
}

// Subscribe registers a new listener for room events. The returned channel is
// buffered generously enough that a slow reader won't stall the room; the
// unsubscribe func must be called to release it.
func (a *RoomActor) Subscribe() (<-chan game.Event, func()) {
	ch := make(chan game.Event, 32)
	a.subMu.Lock()
	a.subscribers[ch] = struct{}{}
	a.subMu.Unlock()

	unsubscribe := func() {
		a.subMu.Lock()
		delete(a.subscribers, ch)
		a.subMu.Unlock()
	}
	return ch, unsubscribe
}

func (a *RoomActor) publish(events []game.Event) {
	if len(events) == 0 {
		return
	}
	a.subMu.Lock()
	defer a.subMu.Unlock()
	for ch := range a.subscribers {
		for _, ev := range events {
			select {
			case ch <- ev:
			default:
				// Slow subscriber; drop rather than block the room.
			}
		}
	}
}

// RegisterToken associates a reconnect token with a stable player identity.
func (a *RoomActor) RegisterToken(token string, id game.PlayerID) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	a.tokens[token] = id
}

// ResolvePlayer looks up the PlayerID previously registered for token.
func (a *RoomActor) ResolvePlayer(token string) (game.PlayerID, bool) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	id, ok := a.tokens[token]
	return id, ok
}

// Stop terminates the actor's goroutine. Safe to call more than once.
func (a *RoomActor) Stop() {
	select {
	case <-a.done:
	default:
		close(a.stop)
	}
}

func (a *RoomActor) loop() {
	defer close(a.done)

	var deadline *time.Timer
	var next *time.Timer
	defer func() {
		stopTimer(deadline)
		stopTimer(next)
	}()

	timerC := func(t *time.Timer) <-chan time.Time {
		if t == nil {
			return nil
		}
		return t.C
	}

	// reschedule reads the events just produced by the room and arms
	// whichever timer should fire next. The events themselves (not a
	// separate query into game.Room) are the source of truth for timing,
	// so this package never needs game.Room to expose its internal clock.
	reschedule := func(events []game.Event) {
		for _, ev := range events {
			switch e := ev.(type) {
			case game.QuestionStarted:
				stopTimer(deadline)
				stopTimer(next)
				deadline = time.NewTimer(time.Until(e.Deadline))
				next = nil
			case game.QuestionRevealed:
				stopTimer(deadline)
				deadline = nil
				stopTimer(next)
				next = time.NewTimer(a.revealDelay)
			case game.GameFinished:
				stopTimer(deadline)
				stopTimer(next)
				deadline, next = nil, nil
			}
		}
	}

	for {
		select {
		case <-a.stop:
			return
		case fn := <-a.cmds:
			events := fn(time.Now())
			a.publish(events)
			reschedule(events)
		case <-timerC(deadline):
			events := a.room.Tick(time.Now())
			a.publish(events)
			reschedule(events)
		case <-timerC(next):
			events := a.room.NextQuestion(time.Now())
			a.publish(events)
			reschedule(events)
		}
	}
}

func stopTimer(t *time.Timer) {
	if t != nil {
		t.Stop()
	}
}
