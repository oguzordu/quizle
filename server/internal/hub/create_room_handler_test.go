package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

func TestServer_CreateRoomHandler_returnsJoinCode(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: 15 * time.Second}}
	srv := NewServer(h)
	srv.SetDefaultQuestions(questions, 4*time.Second)

	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	w := httptest.NewRecorder()

	srv.CreateRoomHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp createRoomResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Code) != 6 {
		t.Errorf("code = %q, want 6 chars", resp.Code)
	}

	if _, ok := h.GetRoom(resp.Code); !ok {
		t.Errorf("room %q not registered in hub", resp.Code)
	}
}

func TestServer_CreateRoomHandler_rejectsNonPost(t *testing.T) {
	h := NewHub()
	srv := NewServer(h)
	srv.SetDefaultQuestions([]game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	w := httptest.NewRecorder()

	srv.CreateRoomHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}
