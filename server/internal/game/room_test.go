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

func TestRoom_Reveal_instantCorrectAnswerGetsBaseplusMaxSpeedBonus(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now) // answered the instant the question opened

	events := r.SubmitAnswer("bob", 2, now) // wrong, and last player -> triggers reveal

	qr := findEvent[QuestionRevealed](t, events)
	if qr.PointsAwarded["alice"] != 150 {
		t.Errorf("alice PointsAwarded = %d, want 150 (100 base + full 50 speed bonus)", qr.PointsAwarded["alice"])
	}
	if _, wrongPlayerScored := qr.PointsAwarded["bob"]; wrongPlayerScored {
		t.Errorf("bob should not appear in PointsAwarded, got %d", qr.PointsAwarded["bob"])
	}
}

func TestRoom_Reveal_slowerCorrectAnswerStillGetsTheBaseGuarantee(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now) // instant -> full bonus

	// Question duration is 10s; answering at the 5s mark leaves half the
	// window, so bob should earn half the speed bonus on top of the base.
	events := r.SubmitAnswer("bob", 1, now.Add(5*time.Second))

	qr := findEvent[QuestionRevealed](t, events)
	if qr.PointsAwarded["alice"] != 150 {
		t.Errorf("alice PointsAwarded = %d, want 150", qr.PointsAwarded["alice"])
	}
	if qr.PointsAwarded["bob"] != 125 {
		t.Errorf("bob PointsAwarded = %d, want 125 (100 base + half the speed bonus)", qr.PointsAwarded["bob"])
	}
}

func TestRoom_GameFinished_awardsFastestPlayerBonusToBestAverageResponder(t *testing.T) {
	players, questions := twoPlayerQuiz()
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 1, now.Add(1*time.Second)) // correct, answered quickly
	r.SubmitAnswer("bob", 1, now.Add(9*time.Second))   // correct, answered slowly
	r.NextQuestion(now.Add(10 * time.Second))

	events := r.SubmitAnswer("alice", 0, now.Add(11*time.Second)) // correct again, quickly
	events = append(events, r.SubmitAnswer("bob", 0, now.Add(19*time.Second))...)
	events = append(events, r.NextQuestion(now.Add(20*time.Second))...)

	gf := findEvent[GameFinished](t, events)
	if !gf.HasFastestPlayer || gf.FastestPlayerID != "alice" {
		t.Fatalf("FastestPlayerID = %q (has=%v), want alice", gf.FastestPlayerID, gf.HasFastestPlayer)
	}
	// alice: two answers at the 90%-of-window mark (145 each) + FastestPlayerBonus.
	// bob: two answers at the 10%-of-window mark (105 each), no bonus.
	if gf.FinalScores["alice"] != 145+145+FastestPlayerBonus {
		t.Errorf("alice final score = %d, want %d", gf.FinalScores["alice"], 145+145+FastestPlayerBonus)
	}
	if gf.FinalScores["bob"] != 105+105 {
		t.Errorf("bob final score = %d, want %d (no bonus, slower on average)", gf.FinalScores["bob"], 105+105)
	}
}

func TestRoom_GameFinished_noFastestBonusIfNobodyAnsweredCorrectly(t *testing.T) {
	players, questions := twoPlayerQuiz()
	questions = questions[:1]
	r := NewRoom(players, questions)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.Start(now)
	r.SubmitAnswer("alice", 3, now) // wrong
	r.SubmitAnswer("bob", 3, now)   // wrong

	events := r.NextQuestion(now.Add(1 * time.Second))

	gf := findEvent[GameFinished](t, events)
	if gf.HasFastestPlayer {
		t.Errorf("HasFastestPlayer = true, want false when nobody answered correctly")
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
	// Both answered at the exact same instant, so it's a genuine tie; the
	// fastest-player bonus tiebreak falls to whichever PlayerID sorts first.
	gf := findEvent[GameFinished](t, events)
	if gf.FinalScores["alice"] != 150+FastestPlayerBonus {
		t.Errorf("alice final score = %d, want %d (instant answer, wins the tiebreak)", gf.FinalScores["alice"], 150+FastestPlayerBonus)
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
