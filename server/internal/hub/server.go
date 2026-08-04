package hub

import (
	"encoding/json"
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

	defaultQuestions   []game.Question
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

// SetDefaultQuestions configures the question set and reveal delay used by
// CreateRoomHandler for every new room.
func (s *Server) SetDefaultQuestions(questions []game.Question, revealDelay time.Duration) {
	s.defaultQuestions = questions
	s.defaultRevealDelay = revealDelay
}

type createRoomResponse struct {
	Code string `json:"code"`
}

// CreateRoomHandler handles POST /rooms, creating a fresh room with the
// server's configured default question set and returning its join code.
func (s *Server) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code, _ := s.hub.CreateRoom(s.defaultQuestions, s.defaultRevealDelay)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createRoomResponse{Code: code})
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
