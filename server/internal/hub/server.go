package hub

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"time"

	gorilla "github.com/gorilla/websocket"
	"github.com/oguzordu/quizle/internal/game"
)

// Server upgrades HTTP requests into WebSocket connections and wires each one
// to the RoomActor named by the "code" query parameter. It is intentionally
// thin: all game logic lives in game.Room, all timing/serialization lives in
// RoomActor; this type only moves bytes.
type Server struct {
	hub      *Hub
	upgrader gorilla.Upgrader

	// pools holds one question set per language ("tr", "en"). A question
	// appears only in the languages it has been written or translated into,
	// so an English-only import is never served to a Turkish room.
	pools              map[string][]game.Question
	sampleSize         int // 0 means "use the whole pool, no sampling"
	defaultRevealDelay time.Duration

	roomLimiter *rateLimiter
}

// defaultLang is used when a client doesn't ask for a specific language.
const defaultLang = "tr"

// NewServer creates a Server backed by h.
func NewServer(h *Hub) *Server {
	return &Server{
		hub: h,
		upgrader: gorilla.Upgrader{
			// No cookie-based auth is involved; identity comes from the
			// explicit token query parameter, so accepting any origin here
			// doesn't expose anything a same-origin policy would protect.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		// A real player creates at most a handful of rooms per session
		// (rematch reuses the existing room); 10 per minute leaves headroom
		// for that while still stopping a scripted flood.
		roomLimiter: newRateLimiter(10, time.Minute),
	}
}

// SetDefaultQuestions configures a fixed question set used verbatim (no
// sampling) by CreateRoomHandler for every new room.
func (s *Server) SetDefaultQuestions(questions []game.Question, revealDelay time.Duration) {
	s.pools = map[string][]game.Question{defaultLang: questions}
	s.sampleSize = 0
	s.defaultRevealDelay = revealDelay
}

// SetQuestionPool configures a large single-language pool, from which
// CreateRoomHandler draws a fresh random sample of sampleSize for each new
// room — so replays don't repeat the same quiz in the same order.
func (s *Server) SetQuestionPool(pool []game.Question, sampleSize int, revealDelay time.Duration) {
	s.SetQuestionPools(map[string][]game.Question{defaultLang: pool}, sampleSize, revealDelay)
}

// SetQuestionPools configures one pool per language, keyed by language code.
func (s *Server) SetQuestionPools(pools map[string][]game.Question, sampleSize int, revealDelay time.Duration) {
	s.pools = pools
	s.sampleSize = sampleSize
	s.defaultRevealDelay = revealDelay
}

// poolFor returns the question set for lang, falling back to the default
// language so an unrecognised code degrades to a playable room rather than
// an empty one.
func (s *Server) poolFor(lang string) []game.Question {
	if p, ok := s.pools[lang]; ok {
		return p
	}
	return s.pools[defaultLang]
}

type createRoomResponse struct {
	Code string `json:"code"`
}

// CreateRoomHandler handles POST /rooms, creating a fresh room with a
// (possibly randomly sampled) question set and returning its join code.
func (s *Server) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.roomLimiter.Allow(clientIP(r)) {
		http.Error(w, "too many rooms created, try again later", http.StatusTooManyRequests)
		return
	}

	lang := langOrDefault(r)
	category := r.URL.Query().Get("category")

	qs, err := s.pickQuestions(lang, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, actor := s.hub.CreateRoom(qs, s.defaultRevealDelay)
	actor.SetSource(QuestionSource{Lang: lang, Category: category})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createRoomResponse{Code: code})
}

type categoryCount struct {
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}

type categoriesResponse struct {
	Categories []categoryCount `json:"categories"`
}

// CategoriesHandler handles GET /categories?lang=tr, listing the categories
// that can actually fill a full round in that language. Turkish and English
// therefore advertise different lists, and each grows on its own as content
// is added — the client never offers a category it can't play.
func (s *Server) CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	counts := make(map[string]int)
	for _, q := range s.poolFor(langOrDefault(r)) {
		counts[q.Category]++
	}

	out := make([]categoryCount, 0, len(counts))
	for slug, n := range counts {
		if n >= s.sampleSize {
			out = append(out, categoryCount{Slug: slug, Count: n})
		}
	}
	// Biggest first, then alphabetically, so the list is stable between
	// calls rather than reordering on Go's random map iteration.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Slug < out[j].Slug
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categoriesResponse{Categories: out})
}

func langOrDefault(r *http.Request) string {
	if lang := r.URL.Query().Get("lang"); lang != "" {
		return lang
	}
	return defaultLang
}

// wsMaxMessageBytes bounds how large a single client WebSocket frame may be.
const wsMaxMessageBytes = 4096

// maxQuestionsPerCategory caps how many questions from the same category can
// land in a single round, so a category that dominates the pool by sheer
// question count (e.g. "movies") doesn't dominate every game too.
const maxQuestionsPerCategory = 4

// pickQuestions draws a room's question set from the pool for lang. When
// category is empty the sample spans categories, capped at
// maxQuestionsPerCategory each so one large category can't dominate every
// game; when a category is named the sample is drawn from it alone.
//
// It fails rather than returning a short game when the requested category
// can't fill a full round in that language — the client is expected to only
// offer categories CategoriesHandler advertised.
func (s *Server) pickQuestions(lang, category string) ([]game.Question, error) {
	available := s.poolFor(lang)

	if category != "" {
		filtered := make([]game.Question, 0, len(available))
		for _, q := range available {
			if q.Category == category {
				filtered = append(filtered, q)
			}
		}
		if len(filtered) < s.sampleSize {
			return nil, fmt.Errorf("category %q has %d questions in %q, need %d", category, len(filtered), lang, s.sampleSize)
		}
		available = filtered
	}

	n := s.sampleSize
	if n <= 0 || n >= len(available) {
		return available, nil
	}

	shuffled := make([]game.Question, len(available))
	copy(shuffled, available)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	// Balancing only makes sense for a mixed round. When the player asked
	// for one category, every question is meant to come from it.
	if category != "" {
		return shuffled[:n], nil
	}

	picked := make([]game.Question, 0, n)
	counts := make(map[string]int)
	var leftover []game.Question

	for _, q := range shuffled {
		if len(picked) == n {
			break
		}
		if counts[q.Category] < maxQuestionsPerCategory {
			picked = append(picked, q)
			counts[q.Category]++
		} else {
			leftover = append(leftover, q)
		}
	}

	for i := 0; len(picked) < n && i < len(leftover); i++ {
		picked = append(picked, leftover[i])
	}

	return picked, nil
}

// StartRoomHandler handles POST /rooms/{code}/start, moving the named room
// out of its lobby and into the first question.
func (s *Server) StartRoomHandler(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	actor, ok := s.hub.GetRoom(code)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	actor.Start()
	w.WriteHeader(http.StatusOK)
}

// RematchRoomHandler handles POST /rooms/{code}/rematch, resetting the named
// room back to its lobby with the same roster and a freshly sampled
// question set, for a "play again" flow.
func (s *Server) RematchRoomHandler(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	actor, ok := s.hub.GetRoom(code)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	src := actor.Source()
	qs, err := s.pickQuestions(src.Lang, src.Category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	actor.Reset(qs)
	w.WriteHeader(http.StatusOK)
}

// ServeWS handles GET /ws?code=XXXXXX&name=Alice[&token=...].
// A request with a valid token rejoins the same PlayerID (reconnect); a
// request without one is treated as a fresh join.
func (s *Server) ServeWS(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	name := r.URL.Query().Get("name")
	avatar := r.URL.Query().Get("avatar")
	tok := r.URL.Query().Get("token")

	actor, ok := s.hub.GetRoom(code)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	var playerID game.PlayerID
	token := tok
	if tok != "" {
		if id, ok := actor.ResolvePlayer(tok); ok {
			playerID = id
		}
	}
	if playerID == "" {
		id, newToken, err := s.hub.JoinRoom(code, name, avatar)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		playerID = id
		token = newToken
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	// Client messages are tiny ({"type":"submit_answer","choice":N}); cap
	// frame size well above that so a malicious client can't force the
	// server to buffer an unbounded amount of memory per connection.
	conn.SetReadLimit(wsMaxMessageBytes)

	if err := conn.WriteMessage(gorilla.TextMessage, encodeJoined(playerID, token)); err != nil {
		return
	}

	sub, unsubscribe := actor.Subscribe()
	defer unsubscribe()

	for _, p := range actor.Roster() {
		if p.ID == playerID {
			continue
		}
		if err := conn.WriteMessage(gorilla.TextMessage, EncodeEvent(game.PlayerJoined{PlayerID: p.ID, Name: p.Name, Avatar: p.Avatar})); err != nil {
			return
		}
	}

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		readPump(conn, actor, playerID)
	}()

	writePump(conn, sub, readerDone)
}

// readPump blocks reading client messages until the connection errors or
// closes, dispatching each decoded command to the actor.
func readPump(conn *gorilla.Conn, actor *RoomActor, playerID game.PlayerID) {
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		cmd, err := DecodeClientMessage(raw)
		if err != nil {
			continue
		}
		switch c := cmd.(type) {
		case SubmitAnswerCommand:
			actor.SubmitAnswer(playerID, c.Choice)
		}
	}
}

// writePump relays room events to the connection until the reader goroutine
// signals the connection is gone (via done) or a write fails.
func writePump(conn *gorilla.Conn, sub <-chan game.Event, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case ev := <-sub:
			if err := conn.WriteMessage(gorilla.TextMessage, EncodeEvent(ev)); err != nil {
				return
			}
		}
	}
}
