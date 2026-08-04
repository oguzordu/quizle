// Package game implements Quizle's core multiplayer quiz logic as a pure,
// network-free state machine. A Room advances only in response to explicit
// method calls carrying a caller-supplied timestamp, and reports what
// happened via returned events. Callers (the WebSocket hub) own goroutines,
// timers, and wall-clock time; this package owns none of that, which is what
// makes it deterministically testable.
package game

import (
	"errors"
	"time"
)

// ErrGameAlreadyStarted is returned by AddPlayer once the Room has left
// PhaseLobby.
var ErrGameAlreadyStarted = errors.New("game already started")

// Phase is where a Room currently sits in its lifecycle.
type Phase int

const (
	PhaseLobby Phase = iota
	PhaseQuestion
	PhaseReveal
	PhaseFinished
)

func (p Phase) String() string {
	switch p {
	case PhaseLobby:
		return "lobby"
	case PhaseQuestion:
		return "question"
	case PhaseReveal:
		return "reveal"
	case PhaseFinished:
		return "finished"
	default:
		return "unknown"
	}
}

// PlayerID uniquely identifies a player within a Room.
type PlayerID string

// Player tracks a participant's identity and running game state.
type Player struct {
	ID     PlayerID
	Name   string
	Score  int
	Streak int

	// totalElapsedCorrect and correctAnswerCount track response speed only
	// for the end-of-game FastestPlayerBonus award — they never affect
	// per-question scoring, so a slower player never loses points to a
	// faster one.
	totalElapsedCorrect time.Duration
	correctAnswerCount  int
}

// Question is one round of the quiz. Correct is the index into Choices.
type Question struct {
	ID       string
	Choices  []string
	Correct  int
	Duration time.Duration
}

// Event is a marker interface for anything a Room reports back to its caller.
// Concrete event types are declared in events.go.
type Event interface{}

type answer struct {
	choice       int
	at           time.Time
	streakBefore int // player's streak going into this answer, for scoring at reveal
}

// Room is the state machine for one game. Zero value is not usable; create
// with NewRoom.
type Room struct {
	phase           Phase
	questions       []Question
	currentIdx      int
	players         map[PlayerID]*Player
	answers         map[PlayerID]answer
	questionStarted time.Time
	questionEnds    time.Time
}

// NewRoom creates a Room in PhaseLobby, ready for Start.
func NewRoom(players []Player, questions []Question) *Room {
	m := make(map[PlayerID]*Player, len(players))
	for i := range players {
		p := players[i]
		m[p.ID] = &p
	}
	return &Room{
		phase:     PhaseLobby,
		questions: questions,
		players:   m,
		answers:   make(map[PlayerID]answer),
	}
}

// Phase reports the Room's current phase.
func (r *Room) Phase() Phase {
	return r.phase
}

// Players returns a snapshot of every enrolled player, for building an
// initial roster when a new connection joins mid-lobby.
func (r *Room) Players() []Player {
	out := make([]Player, 0, len(r.players))
	for _, p := range r.players {
		out = append(out, *p)
	}
	return out
}

// AddPlayer enrolls a new player while the Room is still in the lobby. It is
// rejected once the game has started so mid-game joiners can't dodge
// questions everyone else already answered.
func (r *Room) AddPlayer(id PlayerID, name string) ([]Event, error) {
	if r.phase != PhaseLobby {
		return nil, ErrGameAlreadyStarted
	}
	r.players[id] = &Player{ID: id, Name: name}
	return []Event{PlayerJoined{PlayerID: id, Name: name}}, nil
}

// Start moves the Room from PhaseLobby into the first question.
func (r *Room) Start(now time.Time) []Event {
	return r.beginQuestion(0, now)
}

func (r *Room) beginQuestion(idx int, now time.Time) []Event {
	r.currentIdx = idx
	r.phase = PhaseQuestion
	r.answers = make(map[PlayerID]answer)
	q := r.questions[idx]
	r.questionStarted = now
	r.questionEnds = now.Add(q.Duration)
	return []Event{QuestionStarted{Question: q, Deadline: r.questionEnds}}
}

// SubmitAnswer records a player's choice for the current question.
// Correctness (and its effect on streak) is decided immediately, but the
// points it earns depend on this round's final answer order, so they aren't
// computed until reveal. If every player has now answered, the question is
// revealed immediately.
func (r *Room) SubmitAnswer(id PlayerID, choice int, now time.Time) []Event {
	if r.phase != PhaseQuestion {
		return nil
	}
	if _, already := r.answers[id]; already {
		return nil
	}

	player := r.players[id]
	r.answers[id] = answer{choice: choice, at: now, streakBefore: player.Streak}

	q := r.questions[r.currentIdx]
	correct := choice == q.Correct
	if correct {
		player.Streak++
	} else {
		player.Streak = 0
	}

	events := []Event{AnswerAccepted{PlayerID: id, Correct: correct}}

	if len(r.answers) >= len(r.players) {
		events = append(events, r.reveal()...)
	}
	return events
}

// Tick lets the caller drive timeout-based transitions. If the current
// question's deadline has passed, the question is revealed. Otherwise it is
// a no-op.
func (r *Room) Tick(now time.Time) []Event {
	if r.phase != PhaseQuestion {
		return nil
	}
	if now.Before(r.questionEnds) {
		return nil
	}
	return r.reveal()
}

func (r *Room) reveal() []Event {
	r.phase = PhaseReveal
	q := r.questions[r.currentIdx]

	pointsAwarded := make(map[PlayerID]int)
	for id, a := range r.answers {
		if a.choice != q.Correct {
			continue
		}
		player := r.players[id]
		points := Score(true, q.Duration, r.questionEnds.Sub(a.at), a.streakBefore)
		pointsAwarded[id] = points
		player.Score += points
		player.totalElapsedCorrect += a.at.Sub(r.questionStarted)
		player.correctAnswerCount++
	}

	scores := make(map[PlayerID]int, len(r.players))
	for id, p := range r.players {
		scores[id] = p.Score
	}

	return []Event{QuestionRevealed{CorrectChoice: q.Correct, PointsAwarded: pointsAwarded, Scores: scores}}
}

// NextQuestion advances from PhaseReveal to the next question, or to
// PhaseFinished if the question just revealed was the last one.
func (r *Room) NextQuestion(now time.Time) []Event {
	if r.phase != PhaseReveal {
		return nil
	}
	next := r.currentIdx + 1
	if next >= len(r.questions) {
		r.phase = PhaseFinished

		fastestID, hasFastest := r.fastestPlayer()
		if hasFastest {
			r.players[fastestID].Score += FastestPlayerBonus
		}

		scores := make(map[PlayerID]int, len(r.players))
		for id, p := range r.players {
			scores[id] = p.Score
		}
		return []Event{GameFinished{
			FinalScores:      scores,
			HasFastestPlayer: hasFastest,
			FastestPlayerID:  fastestID,
		}}
	}
	return r.beginQuestion(next, now)
}

// fastestPlayer finds whoever had the lowest average response time among
// their correct answers. Players with no correct answers don't qualify.
// Exact ties are broken by PlayerID so the result is deterministic
// regardless of Go's randomized map iteration order.
func (r *Room) fastestPlayer() (PlayerID, bool) {
	var bestID PlayerID
	var bestAvg time.Duration
	found := false

	for id, p := range r.players {
		if p.correctAnswerCount == 0 {
			continue
		}
		avg := p.totalElapsedCorrect / time.Duration(p.correctAnswerCount)
		if !found || avg < bestAvg || (avg == bestAvg && id < bestID) {
			bestID, bestAvg, found = id, avg, true
		}
	}

	return bestID, found
}
