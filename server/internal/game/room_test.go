package game

import (
	"testing"
	"time"
)

func twoPlayerQuiz() (players []Player, questions []Question) {
	players = []Player{
		{ID: "alice", Name: "Alice"},
		{ID: "bob", Name: "Bob"},
	}
	questions = []Question{
		{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 1, Duration: 10 * time.Second},
		{ID: "q2", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: 10 * time.Second},
	}
	return
}

func TestRoom_Start_beginsFirstQuestion(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	events := r.Start(now)

	if r.Phase() != PhaseQuestion {
		t.Fatalf("Phase() = %v, want PhaseQuestion", r.Phase())
	}
	qs := findEvent[QuestionStarted](t, events)
	if qs.Question.ID != "q1" {
		t.Errorf("started question = %q, want q1", qs.Question.ID)
	}
	wantDeadline := now.Add(10 * time.Second)
	if !qs.Deadline.Equal(wantDeadline) {
		t.Errorf("deadline = %v, want %v", qs.Deadline, wantDeadline)
	}
}

func TestRoom_SubmitAnswer_correctIsConfirmedImmediatelyButUnscoredUntilReveal(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)

	events := r.SubmitAnswer("alice", 1, now)

	aa := findEvent[AnswerAccepted](t, events)
	if !aa.Correct {
		t.Errorf("Correct = false, want true")
	}
	// Points depend on the final answer order, so nothing is awarded yet.
	if r.players["alice"].Score != 0 {
		t.Errorf("player score = %d, want 0 before reveal", r.players["alice"].Score)
	}
}

func TestRoom_SubmitAnswer_incorrectResetsStreakImmediately(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.players["alice"].Streak = 4

	events := r.SubmitAnswer("alice", 2, now)

	aa := findEvent[AnswerAccepted](t, events)
	if aa.Correct {
		t.Errorf("Correct = true, want false")
	}
	if r.players["alice"].Streak != 0 {
		t.Errorf("streak = %d, want 0 after wrong answer", r.players["alice"].Streak)
	}
}

func TestRoom_Reveal_soleCorrectAnswererGetsTopRankPoints(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now)

	events := r.SubmitAnswer("bob", 2, now) // wrong, and last player -> triggers reveal

	qr := findEvent[QuestionRevealed](t, events)
	if qr.PointsAwarded["alice"] != 110 {
		t.Errorf("alice PointsAwarded = %d, want 110 (sole correct answerer, also fastest)", qr.PointsAwarded["alice"])
	}
	if _, wrongPlayerScored := qr.PointsAwarded["bob"]; wrongPlayerScored {
		t.Errorf("bob should not appear in PointsAwarded, got %d", qr.PointsAwarded["bob"])
	}
}

func TestRoom_Reveal_firstCorrectAnswererGetsBonusButKnowledgeDominates(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now) // correct, first

	events := r.SubmitAnswer("bob", 1, now.Add(time.Second)) // correct, a full second later

	qr := findEvent[QuestionRevealed](t, events)
	if qr.PointsAwarded["alice"] != 110 {
		t.Errorf("alice PointsAwarded = %d, want 110 (answered first, gets small bonus)", qr.PointsAwarded["alice"])
	}
	// Being slower never costs Bob the base points — knowing the answer is
	// what mostly counts. The gap between them is only the small speed bonus.
	if qr.PointsAwarded["bob"] != 100 {
		t.Errorf("bob PointsAwarded = %d, want 100 (answered second, still correct, no bonus)", qr.PointsAwarded["bob"])
	}
}

func TestRoom_Reveal_nearSimultaneousCorrectAnswersBothGetTheBonus(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now)

	events := r.SubmitAnswer("bob", 1, now.Add(100*time.Millisecond)) // within the tie window

	qr := findEvent[QuestionRevealed](t, events)
	if qr.PointsAwarded["alice"] != 110 || qr.PointsAwarded["bob"] != 110 {
		t.Errorf("PointsAwarded = %+v, want both at 110 (tied for fastest)", qr.PointsAwarded)
	}
}

func TestRoom_AllPlayersAnswered_advancesToRevealEarly(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)

	r.SubmitAnswer("alice", 1, now)
	if r.Phase() != PhaseQuestion {
		t.Fatalf("Phase() = %v after 1/2 answers, want still PhaseQuestion", r.Phase())
	}

	events := r.SubmitAnswer("bob", 0, now.Add(time.Second))

	if r.Phase() != PhaseReveal {
		t.Fatalf("Phase() = %v after 2/2 answers, want PhaseReveal", r.Phase())
	}
	findEvent[QuestionRevealed](t, events)
}

func TestRoom_Tick_timeoutAdvancesToRevealWithUnansweredScoringZero(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now)

	events := r.Tick(now.Add(10 * time.Second))

	if r.Phase() != PhaseReveal {
		t.Fatalf("Phase() = %v after timeout, want PhaseReveal", r.Phase())
	}
	findEvent[QuestionRevealed](t, events)
	if r.players["bob"].Score != 0 {
		t.Errorf("bob score = %d, want 0 (never answered)", r.players["bob"].Score)
	}
}

func TestRoom_Tick_beforeDeadlineDoesNothing(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)

	events := r.Tick(now.Add(time.Second))

	if r.Phase() != PhaseQuestion {
		t.Fatalf("Phase() = %v, want still PhaseQuestion", r.Phase())
	}
	if len(events) != 0 {
		t.Errorf("events = %v, want none", events)
	}
}

func TestRoom_NextQuestion_advancesToSecondQuestion(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now)
	r.SubmitAnswer("bob", 1, now)

	events := r.NextQuestion(now.Add(3 * time.Second))

	if r.Phase() != PhaseQuestion {
		t.Fatalf("Phase() = %v, want PhaseQuestion", r.Phase())
	}
	qs := findEvent[QuestionStarted](t, events)
	if qs.Question.ID != "q2" {
		t.Errorf("started question = %q, want q2", qs.Question.ID)
	}
}

func TestRoom_NextQuestion_afterLastQuestionFinishesGame(t *testing.T) {
	players, questions := twoPlayerQuiz()
	questions = questions[:1] // single-question game
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now)
	r.SubmitAnswer("bob", 1, now)

	events := r.NextQuestion(now.Add(3 * time.Second))

	if r.Phase() != PhaseFinished {
		t.Fatalf("Phase() = %v, want PhaseFinished", r.Phase())
	}
	gf := findEvent[GameFinished](t, events)
	if gf.FinalScores["alice"] != 110 {
		t.Errorf("alice final score = %d, want 110 (correct, tied for fastest)", gf.FinalScores["alice"])
	}
}

func TestRoom_StreakBuildsAcrossQuestions(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now) // correct -> streak 1
	r.SubmitAnswer("bob", 1, now)
	r.NextQuestion(now.Add(3 * time.Second))

	r.SubmitAnswer("alice", 0, now.Add(3*time.Second)) // correct again -> streak 2

	if r.players["alice"].Streak != 2 {
		t.Errorf("alice streak = %d, want 2", r.players["alice"].Streak)
	}
}

func TestRoom_AddPlayer_succeedsDuringLobby(t *testing.T) {
	_, questions := twoPlayerQuiz()
	r := NewRoom(nil, questions)

	err := r.AddPlayer("alice", "Alice")

	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if _, ok := r.players["alice"]; !ok {
		t.Fatalf("player alice not present after AddPlayer")
	}
}

func TestRoom_AddPlayer_rejectedOnceGameStarted(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	r.Start(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	err := r.AddPlayer("carol", "Carol")

	if err == nil {
		t.Fatal("expected error adding player after game started, got nil")
	}
}

// findEvent returns the first event of type T in events, failing the test if none found.
func findEvent[T any](t *testing.T, events []Event) T {
	t.Helper()
	for _, e := range events {
		if v, ok := e.(T); ok {
			return v
		}
	}
	var zero T
	t.Fatalf("no event of type %T found in %#v", zero, events)
	return zero
}
