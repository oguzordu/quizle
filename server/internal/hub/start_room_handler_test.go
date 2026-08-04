package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

func TestServer_StartRoomHandler_startsTheGame(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}
	code, actor := h.CreateRoom(questions, time.Second)
	t.Cleanup(actor.Stop)

	srv := NewServer(h)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rooms/{code}/start", srv.StartRoomHandler)

	req := httptest.NewRequest(http.MethodPost, "/rooms/"+code+"/start", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if actor.room.Phase() != game.PhaseQuestion {
		t.Errorf("Phase() = %v, want PhaseQuestion", actor.room.Phase())
	}
}

func TestServer_RematchRoomHandler_resetsRoomToLobby(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}
	code, actor := h.CreateRoom(questions, time.Second)
	t.Cleanup(actor.Stop)
	actor.AddPlayer("alice", "Alice", "")
	actor.Start()
	actor.SubmitAnswer("alice", 0)

	srv := NewServer(h)
	srv.SetQuestionPool(questions, 0, time.Second)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rooms/{code}/rematch", srv.RematchRoomHandler)

	req := httptest.NewRequest(http.MethodPost, "/rooms/"+code+"/rematch", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if actor.room.Phase() != game.PhaseLobby {
		t.Errorf("Phase() = %v, want PhaseLobby", actor.room.Phase())
	}
	for _, p := range actor.room.Players() {
		if p.ID == "alice" && p.Score != 0 {
			t.Errorf("alice score = %d, want 0 after rematch reset", p.Score)
		}
	}
}

func TestServer_StartRoomHandler_unknownCodeReturns404(t *testing.T) {
	h := NewHub()
	srv := NewServer(h)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rooms/{code}/start", srv.StartRoomHandler)

	req := httptest.NewRequest(http.MethodPost, "/rooms/ZZZZZZ/start", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
