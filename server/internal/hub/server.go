package hub

import (
	"encoding/json"
	"math/rand"
	"net/http"
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

	questionPool       []game.Question
	sampleSize         int // 0 means "use the whole pool, no sampling"
	defaultRevealDelay time.Duration
}

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
	}
}

// SetDefaultQuestions configures a fixed question set used verbatim (no
// sampling) by CreateRoomHandler for every new room.
func (s *Server) SetDefaultQuestions(questions []game.Question, revealDelay time.Duration) {
	s.questionPool = questions
	s.sampleSize = 0
	s.defaultRevealDelay = revealDelay
}

// SetQuestionPool configures a large pool of questions, from which
// CreateRoomHandler draws a fresh random sample of sampleSize for each new
// room — so replays don't repeat the same quiz in the same order.
func (s *Server) SetQuestionPool(pool []game.Question, sampleSize int, revealDelay time.Duration) {
	s.questionPool = pool
	s.sampleSize = sampleSize
	s.defaultRevealDelay = revealDelay
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

	code, _ := s.hub.CreateRoom(s.pickQuestions(), s.defaultRevealDelay)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createRoomResponse{Code: code})
}

// maxQuestionsPerCategory caps how many questions from the same category can
// land in a single round, so a category that dominates the pool by sheer
// question count (e.g. "movies") doesn't dominate every game too.
const maxQuestionsPerCategory = 4

// pickQuestions returns the configured question set, or a random sample of
// it when a sampleSize smaller than the pool has been configured. The sample
// is drawn so no category contributes more than maxQuestionsPerCategory
// questions, falling back to filling remaining slots from any category if
// the pool doesn't have enough variety to honor that cap.
func (s *Server) pickQuestions() []game.Question {
	n := s.sampleSize
	if n <= 0 || n >= len(s.questionPool) {
		return s.questionPool
	}

	shuffled := make([]game.Question, len(s.questionPool))
	copy(shuffled, s.questionPool)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

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

	return picked
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
	actor.Reset(s.pickQuestions())
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
