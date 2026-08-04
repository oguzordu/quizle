package hub

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

// roomCodeAlphabet excludes visually ambiguous characters (0/O, 1/I) since
// codes are read aloud and typed by hand.
const roomCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const roomCodeLength = 6

// Hub is the process-wide registry of active rooms, keyed by their
// human-shareable join code.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*RoomActor
}

// NewHub creates an empty registry.
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]*RoomActor)}
}

// CreateRoom starts a new RoomActor and registers it under a freshly
// generated, collision-free join code.
func (h *Hub) CreateRoom(questions []game.Question, revealDelay time.Duration) (string, *RoomActor) {
	actor := NewRoomActor(nil, questions, revealDelay)

	h.mu.Lock()
	defer h.mu.Unlock()
	var code string
	for {
		code = generateCode()
		if _, exists := h.rooms[code]; !exists {
			break
		}
	}
	h.rooms[code] = actor
	return code, actor
}

// GetRoom looks up a room by its join code.
func (h *Hub) GetRoom(code string) (*RoomActor, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	actor, ok := h.rooms[code]
	return actor, ok
}

// JoinRoom enrolls a new player in the room identified by code, returning
// their assigned PlayerID and a reconnect token the client must present to
// resume the same identity after a dropped connection.
func (h *Hub) JoinRoom(code string, name, avatar string) (game.PlayerID, string, error) {
	actor, ok := h.GetRoom(code)
	if !ok {
		return "", "", fmt.Errorf("room %q not found", code)
	}

	id := game.PlayerID(generateToken(8))
	if err := actor.AddPlayer(id, name, avatar); err != nil {
		return "", "", err
	}

	token := generateToken(16)
	actor.RegisterToken(token, id)

	return id, token, nil
}

func generateCode() string {
	b := make([]byte, roomCodeLength)
	buf := make([]byte, roomCodeLength)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("hub: crypto/rand unavailable: %v", err))
	}
	for i, v := range buf {
		b[i] = roomCodeAlphabet[int(v)%len(roomCodeAlphabet)]
	}
	return string(b)
}

func generateToken(nBytes int) string {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("hub: crypto/rand unavailable: %v", err))
	}
	return fmt.Sprintf("%x", buf)
}
