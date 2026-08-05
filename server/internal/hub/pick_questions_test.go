package hub

import (
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

// buildUnbalancedPool creates a pool dominated by one category, mirroring
// the real question bank where "movies" (89 questions) vastly outnumbers
// every other category.
func buildUnbalancedPool() []game.Question {
	var pool []game.Question
	for i := 0; i < 50; i++ {
		pool = append(pool, game.Question{ID: id("movies", i), Category: "movies", Choices: []string{"a", "b", "c", "d"}})
	}
	for _, cat := range []string{"science", "history", "sports", "geography"} {
		for i := 0; i < 6; i++ {
			pool = append(pool, game.Question{ID: id(cat, i), Category: cat, Choices: []string{"a", "b", "c", "d"}})
		}
	}
	return pool
}

func id(cat string, i int) string {
	return cat + "-" + string(rune('a'+i))
}

func TestServer_PickQuestions_NeverExceedsMaxPerCategory(t *testing.T) {
	srv := NewServer(NewHub())
	srv.SetQuestionPool(buildUnbalancedPool(), 10, time.Second)

	for round := 0; round < 500; round++ {
		picked := srv.pickQuestions()
		if len(picked) != 10 {
			t.Fatalf("round %d: got %d questions, want 10", round, len(picked))
		}
		counts := map[string]int{}
		for _, q := range picked {
			counts[q.Category]++
			if counts[q.Category] > maxQuestionsPerCategory {
				t.Fatalf("round %d: category %q appeared %d times, want <= %d", round, q.Category, counts[q.Category], maxQuestionsPerCategory)
			}
		}
	}
}
