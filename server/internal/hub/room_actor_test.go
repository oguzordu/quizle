package hub

import (
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

const testTimeout = 2 * time.Second

func actorFixture(t *testing.T, questionDuration, revealDelay time.Duration) *RoomActor {
	t.Helper()
	players := []game.Player{
		{ID: "alice", Name: "Alice"},
		{ID: "bob", Name: "Bob"},
	}
	questions := []game.Question{
		{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 1, Duration: questionDuration},
		{ID: "q2", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: questionDuration},
	}
	a := NewRoomActor(players, questions, revealDelay)
	t.Cleanup(a.Stop)
	return a
}

func recvEvent[T any](t *testing.T, sub <-chan game.Event) T {
	t.Helper()
	for {
		select {
		case ev := <-sub:
			if v, ok := ev.(T); ok {
				return v
			}
		case <-time.After(testTimeout):
			var zero T
			t.Fatalf("timed out waiting for event of type %T", zero)
			return zero
		}
	}
}

func TestRoomActor_StartPublishesQuestionStarted(t *testing.T) {
	a := actorFixture(t, time.Hour, time.Hour)
	sub, unsub := a.Subscribe()
	defer unsub()

	a.Start()

	qs := recvEvent[game.QuestionStarted](t, sub)
	if qs.Question.ID != "q1" {
		t.Errorf("QuestionID = %q, want q1", qs.Question.ID)
	}
}

func TestRoomActor_SubmitAnswerBySecondPlayerTriggersReveal(t *testing.T) {
	a := actorFixture(t, time.Hour, time.Hour)
	sub, unsub := a.Subscribe()
	defer unsub()

	a.Start()
	recvEvent[game.QuestionStarted](t, sub)

	a.SubmitAnswer("alice", 1)
	recvEvent[game.AnswerAccepted](t, sub)

	a.SubmitAnswer("bob", 0)
	recvEvent[game.AnswerAccepted](t, sub)
	recvEvent[game.QuestionRevealed](t, sub)
}

func TestRoomActor_DeadlineTimerRevealsWithoutAllAnswers(t *testing.T) {
	a := actorFixture(t, 30*time.Millisecond, time.Hour)
	sub, unsub := a.Subscribe()
	defer unsub()

	a.Start()
	recvEvent[game.QuestionStarted](t, sub)

	// Nobody answers; the actor's own deadline timer must fire the reveal.
	recvEvent[game.QuestionRevealed](t, sub)
}

func TestRoomActor_AutoAdvancesToNextQuestionAfterRevealDelay(t *testing.T) {
	a := actorFixture(t, time.Hour, 20*time.Millisecond)
	sub, unsub := a.Subscribe()
	defer unsub()

	a.Start()
	recvEvent[game.QuestionStarted](t, sub)
	a.SubmitAnswer("alice", 1)
	recvEvent[game.AnswerAccepted](t, sub)
	a.SubmitAnswer("bob", 1)
	recvEvent[game.AnswerAccepted](t, sub)
	recvEvent[game.QuestionRevealed](t, sub)

	qs := recvEvent[game.QuestionStarted](t, sub)
	if qs.Question.ID != "q2" {
		t.Errorf("QuestionID = %q, want q2 (auto-advanced)", qs.Question.ID)
	}
}

func TestRoomActor_ReconnectTokenPreservesPlayerIdentity(t *testing.T) {
	a := actorFixture(t, time.Hour, time.Hour)

	id, ok := a.ResolvePlayer("token-alice")
	if ok {
		t.Fatalf("expected token-alice to be unregistered initially")
	}
	a.RegisterToken("token-alice", "alice")

	id, ok = a.ResolvePlayer("token-alice")
	if !ok || id != "alice" {
		t.Fatalf("ResolvePlayer(token-alice) = (%v, %v), want (alice, true)", id, ok)
	}

	// Simulate a dropped connection reconnecting with the same token: score
	// must have survived because the underlying game.Room keyed it by
	// PlayerID, and the actor never lost that state.
	sub, unsub := a.Subscribe()
	defer unsub()
	a.Start()
	recvEvent[game.QuestionStarted](t, sub)
	a.SubmitAnswer("alice", 1)
	recvEvent[game.AnswerAccepted](t, sub)

	id, ok = a.ResolvePlayer("token-alice")
	if !ok || id != "alice" {
		t.Fatalf("ResolvePlayer after answering = (%v, %v), want (alice, true)", id, ok)
	}
}
