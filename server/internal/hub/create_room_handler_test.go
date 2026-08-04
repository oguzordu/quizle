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

func TestServer_CreateRoomHandler_samplesFromPool(t *testing.T) {
	h := NewHub()
	pool := make([]game.Question, 10)
	for i := range pool {
		pool[i] = game.Question{ID: string(rune('a' + i)), Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: 20 * time.Millisecond}
	}
	srv := NewServer(h)
	srv.SetQuestionPool(pool, 3, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	w := httptest.NewRecorder()
	srv.CreateRoomHandler(w, req)

	var resp createRoomResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	actor, ok := h.GetRoom(resp.Code)
	if !ok {
		t.Fatalf("room %q not found", resp.Code)
	}
	t.Cleanup(actor.Stop)

	sub, unsub := actor.Subscribe()
	defer unsub()
	actor.Start()

	started := 0
	for {
		select {
		case ev := <-sub:
			switch ev.(type) {
			case game.QuestionStarted:
				started++
			case game.GameFinished:
				if started != 3 {
					t.Errorf("QuestionStarted fired %d times, want 3 (sampled from pool of 10)", started)
				}
				return
			}
		case <-time.After(testTimeout):
			t.Fatal("timed out waiting for game to finish")
		}
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
