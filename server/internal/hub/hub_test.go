package hub

import (
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

func TestHub_CreateRoom_generatesSixCharCode(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}

	code, actor := h.CreateRoom(questions, 3*time.Second)
	t.Cleanup(actor.Stop)

	if len(code) != 6 {
		t.Errorf("len(code) = %d, want 6", len(code))
	}
}

func TestHub_GetRoom_findsCreatedRoom(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}

	code, actor := h.CreateRoom(questions, 3*time.Second)
	t.Cleanup(actor.Stop)

	got, ok := h.GetRoom(code)
	if !ok {
		t.Fatalf("GetRoom(%q) not found", code)
	}
	if got != actor {
		t.Errorf("GetRoom(%q) returned a different actor", code)
	}
}

func TestHub_GetRoom_unknownCodeNotFound(t *testing.T) {
	h := NewHub()

	_, ok := h.GetRoom("ZZZZZZ")
	if ok {
		t.Fatalf("GetRoom(ZZZZZZ) = found, want not found")
	}
}

func TestHub_CreateRoom_codesAreUnique(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		code, actor := h.CreateRoom(questions, 3*time.Second)
		t.Cleanup(actor.Stop)
		if seen[code] {
			t.Fatalf("duplicate room code generated: %s", code)
		}
		seen[code] = true
	}
}

func TestHub_JoinRoom_addsPlayerAndReturnsToken(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}
	code, actor := h.CreateRoom(questions, 3*time.Second)
	t.Cleanup(actor.Stop)

	id, token, err := h.JoinRoom(code, "Alice", "🦊|#FF5733")
	if err != nil {
		t.Fatalf("JoinRoom: %v", err)
	}
	if id == "" || token == "" {
		t.Fatalf("JoinRoom returned empty id/token: %q/%q", id, token)
	}

	resolved, ok := actor.ResolvePlayer(token)
	if !ok || resolved != id {
		t.Errorf("ResolvePlayer(token) = (%v, %v), want (%v, true)", resolved, ok, id)
	}
}

func TestHub_JoinRoom_unknownCodeErrors(t *testing.T) {
	h := NewHub()

	_, _, err := h.JoinRoom("ZZZZZZ", "Alice", "")
	if err == nil {
		t.Fatal("expected error joining unknown room code, got nil")
	}
}
