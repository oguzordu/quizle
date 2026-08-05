package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

// pool builds n questions in one category, all in the given language pool.
func pool(category string, n int) []game.Question {
	out := make([]game.Question, n)
	for i := range out {
		out[i] = game.Question{
			ID:       category + "-" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			Category: category,
			Choices:  []string{"a", "b", "c", "d"},
			Duration: time.Second,
		}
	}
	return out
}

func serverWithPools(t *testing.T) *Server {
	t.Helper()
	srv := NewServer(NewHub())
	// Three categories in Turkish so a 10-question round can honour the
	// per-category cap (3 x 4 >= 10); celebrities is English-only.
	tr := append(append(pool("movies", 20), pool("music", 20)...), pool("history", 20)...)
	en := append(append([]game.Question{}, tr...), pool("celebrities", 20)...)
	srv.SetQuestionPools(map[string][]game.Question{"tr": tr, "en": en}, 10, time.Second)
	return srv
}

func TestServer_PickQuestions_filtersToRequestedCategory(t *testing.T) {
	srv := serverWithPools(t)

	picked, err := srv.pickQuestions("tr", "music")
	if err != nil {
		t.Fatalf("pickQuestions: %v", err)
	}
	if len(picked) != 10 {
		t.Fatalf("got %d questions, want 10", len(picked))
	}
	for _, q := range picked {
		if q.Category != "music" {
			t.Fatalf("got a %q question, want only music", q.Category)
		}
	}
}

// Without a category the existing mixed behaviour must survive: a spread
// across categories, capped per category.
func TestServer_PickQuestions_withoutCategoryStaysBalanced(t *testing.T) {
	srv := serverWithPools(t)

	for round := 0; round < 200; round++ {
		picked, err := srv.pickQuestions("tr", "")
		if err != nil {
			t.Fatalf("pickQuestions: %v", err)
		}
		counts := map[string]int{}
		for _, q := range picked {
			counts[q.Category]++
		}
		for cat, n := range counts {
			if n > maxQuestionsPerCategory {
				t.Fatalf("round %d: %q appeared %d times, want <= %d", round, cat, n, maxQuestionsPerCategory)
			}
		}
	}
}

// A category present in English but not Turkish must not leak into a
// Turkish room.
func TestServer_PickQuestions_rejectsCategoryMissingInLanguage(t *testing.T) {
	srv := serverWithPools(t)

	if _, err := srv.pickQuestions("tr", "celebrities"); err == nil {
		t.Fatal("expected an error for a category with no Turkish questions, got nil")
	}
	if _, err := srv.pickQuestions("en", "celebrities"); err != nil {
		t.Fatalf("celebrities should work in English: %v", err)
	}
}

func TestServer_CreateRoomHandler_honoursLangAndCategory(t *testing.T) {
	srv := serverWithPools(t)

	req := httptest.NewRequest(http.MethodPost, "/rooms?lang=en&category=celebrities", nil)
	w := httptest.NewRecorder()
	srv.CreateRoomHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp createRoomResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	actor, ok := srv.hub.GetRoom(resp.Code)
	if !ok {
		t.Fatalf("room %q not registered", resp.Code)
	}
	t.Cleanup(actor.Stop)
}

func TestServer_CreateRoomHandler_rejectsUnknownCategory(t *testing.T) {
	srv := serverWithPools(t)

	req := httptest.NewRequest(http.MethodPost, "/rooms?category=underwater-basketweaving", nil)
	w := httptest.NewRecorder()
	srv.CreateRoomHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// The client renders whatever this returns, so it must never offer a
// category that can't fill a full game in that language.
func TestServer_CategoriesHandler_listsOnlyPlayableCategories(t *testing.T) {
	srv := NewServer(NewHub())
	srv.SetQuestionPools(map[string][]game.Question{
		"tr": append(pool("movies", 20), pool("music", 3)...), // music can't fill 10
		"en": pool("celebrities", 20),
	}, 10, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/categories?lang=tr", nil)
	w := httptest.NewRecorder()
	srv.CategoriesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp categoriesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	got := map[string]bool{}
	for _, c := range resp.Categories {
		got[c.Slug] = true
	}
	if !got["movies"] {
		t.Error("movies (20 questions) should be listed")
	}
	if got["music"] {
		t.Error("music has only 3 questions and should not be listed")
	}
	if got["celebrities"] {
		t.Error("celebrities is English-only and should not appear for lang=tr")
	}
}
