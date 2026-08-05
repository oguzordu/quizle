// Command quizle runs the Quizle game server: a WebSocket endpoint and a
// room creation endpoint. The real client is the Next.js app; this binary
// only serves the API.
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
	pools, err := loadQuestionPools("data/questions")
	if err != nil {
		log.Fatalf("soru paketi yüklenemedi: %v", err)
	}

	h := hub.NewHub()
	srv := hub.NewServer(h)
	srv.SetQuestionPools(pools, 10, 4*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/rooms", srv.CreateRoomHandler)
	mux.HandleFunc("/categories", srv.CategoriesHandler)
	mux.HandleFunc("POST /rooms/{code}/start", srv.StartRoomHandler)
	mux.HandleFunc("POST /rooms/{code}/rematch", srv.RematchRoomHandler)
	mux.HandleFunc("/ws", srv.ServeWS)
	mux.HandleFunc("/", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Quizle sunucusu %s portunda çalışıyor (TR: %d soru, EN: %d soru)", addr, len(pools["tr"]), len(pools["en"]))
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

// loadQuestionPools reads every *.json pack in dir into one pool per
// language. A question joins a language's pool only if it has been written
// in that language, so English-only imports (OpenTDB) never surface in a
// Turkish room and vice versa.
func loadQuestionPools(dir string) (map[string][]game.Question, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pools := map[string][]game.Question{"tr": {}, "en": {}}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if err := loadQuestionPack(filepath.Join(dir, entry.Name()), pools); err != nil {
			return nil, err
		}
	}
	return pools, nil
}

// localizedPack mirrors the on-disk question format. Only the fields the
// game engine needs are decoded.
type localizedPack struct {
	Text    string   `json:"text"`
	Choices []string `json:"choices"`
}

func loadQuestionPack(path string, pools map[string][]game.Question) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw []struct {
		ID       string        `json:"id"`
		Category string        `json:"category"`
		Correct  int           `json:"correct"`
		Image    string        `json:"image"`
		TR       localizedPack `json:"tr"`
		EN       localizedPack `json:"en"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	for _, q := range raw {
		for lang, l := range map[string]localizedPack{"tr": q.TR, "en": q.EN} {
			if l.Text == "" || len(l.Choices) == 0 {
				continue // not written in this language
			}
			pools[lang] = append(pools[lang], game.Question{
				ID:       q.ID,
				Category: q.Category,
				Text:     l.Text,
				Image:    q.Image,
				Choices:  l.Choices,
				Correct:  q.Correct,
				Duration: 15 * time.Second,
			})
		}
	}
	return nil
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Quizle API"))
}
