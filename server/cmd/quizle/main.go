// Command quizle runs the Quizle game server: a WebSocket endpoint, a room
// creation endpoint, and (until the real Next.js frontend in Faz 4 exists) a
// bare-bones manual test page so the game can actually be played end to end.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/oguzordu/quizle/internal/game"
	"github.com/oguzordu/quizle/internal/hub"
)

func main() {
	pool, err := loadQuestionPool("data/questions")
	if err != nil {
		log.Fatalf("soru paketi yüklenemedi: %v", err)
	}

	h := hub.NewHub()
	srv := hub.NewServer(h)
	srv.SetQuestionPool(pool, 20, 4*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/rooms", srv.CreateRoomHandler)
	mux.HandleFunc("POST /rooms/{code}/start", srv.StartRoomHandler)
	mux.HandleFunc("/ws", srv.ServeWS)
	mux.HandleFunc("/", serveTestPage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Quizle sunucusu %s portunda çalışıyor (%d soru havuzda)", addr, len(pool))
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

// withCORS allows the Next.js dev server (a different origin/port) to call
// the room HTTP endpoints. There's no cookie-based auth to protect here —
// identity comes from the explicit token query parameter — so a permissive
// dev policy doesn't expose anything sensitive.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loadQuestionPool reads every *.json pack in dir into one combined pool.
// CreateRoomHandler draws a random sample from this pool for each new room.
func loadQuestionPool(dir string) ([]game.Question, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var pool []game.Question
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		qs, err := loadQuestionPack(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		pool = append(pool, qs...)
	}
	return pool, nil
}

func loadQuestionPack(path string) ([]game.Question, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw []struct {
		ID      string `json:"id"`
		Correct int    `json:"correct"`
		TR      struct {
			Text    string   `json:"text"`
			Choices []string `json:"choices"`
		} `json:"tr"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	questions := make([]game.Question, 0, len(raw))
	for _, q := range raw {
		questions = append(questions, game.Question{
			ID:       q.ID,
			Text:     q.TR.Text,
			Choices:  q.TR.Choices,
			Correct:  q.Correct,
			Duration: 15 * time.Second,
		})
	}
	return questions, nil
}

func serveTestPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(testPageHTML))
}
