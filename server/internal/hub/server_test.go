package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"
	"github.com/oguzordu/quizle/internal/game"
)

// dialPlayer joins a room over real HTTP + WebSocket (httptest.Server, real
// gorilla/websocket client) and returns the connection plus the player's
// assigned id, read off the server's "joined" message.
func dialPlayer(t *testing.T, wsURL, code, name string) (*gorilla.Conn, game.PlayerID) {
	t.Helper()
	conn, _, err := gorilla.DefaultDialer.Dial(wsURL+"?code="+code+"&name="+name, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	msg := readServerMessage(t, conn)
	if msg.Type != "joined" {
		t.Fatalf("first message type = %q, want joined", msg.Type)
	}
	var payload joinedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal joined payload: %v", err)
	}
	return conn, payload.PlayerID
}

func readServerMessage(t *testing.T, conn *gorilla.Conn) ServerMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(testTimeout))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	var msg ServerMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal server message: %v", err)
	}
	return msg
}

// readServerMessageOfType drains messages until one of the wanted type
// arrives, skipping ones that don't match (e.g. a second player's own
// answer_accepted arriving before the reveal).
func readServerMessageOfType(t *testing.T, conn *gorilla.Conn, wantType string) ServerMessage {
	t.Helper()
	for i := 0; i < 10; i++ {
		msg := readServerMessage(t, conn)
		if msg.Type == wantType {
			return msg
		}
	}
	t.Fatalf("never saw a %q message", wantType)
	return ServerMessage{}
}

func TestServer_TwoPlayersPlayFullRoundOverRealWebSockets(t *testing.T) {
	h := NewHub()
	questions := []game.Question{
		{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 1, Duration: 5 * time.Second},
	}
	code, actor := h.CreateRoom(questions, 50*time.Millisecond)
	t.Cleanup(actor.Stop)

	srv := NewServer(h)
	ts := httptest.NewServer(http.HandlerFunc(srv.ServeWS))
	t.Cleanup(ts.Close)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	aliceConn, aliceID := dialPlayer(t, wsURL, code, "Alice")
	bobConn, bobID := dialPlayer(t, wsURL, code, "Bob")
	_ = aliceID
	_ = bobID

	actor.Start()

	readServerMessageOfType(t, aliceConn, "question_started")
	readServerMessageOfType(t, bobConn, "question_started")

	mustWriteJSON(t, aliceConn, map[string]any{"type": "submit_answer", "choice": 1})
	mustWriteJSON(t, bobConn, map[string]any{"type": "submit_answer", "choice": 0})

	revealed := readServerMessageOfType(t, aliceConn, "question_revealed")
	var payload questionRevealedPayload
	if err := json.Unmarshal(revealed.Payload, &payload); err != nil {
		t.Fatalf("unmarshal reveal payload: %v", err)
	}
	// Per-question scoring is flat — speed only matters for the separate
	// end-of-game fastest-player bonus, not here.
	if payload.Scores[aliceID] != 100 {
		t.Errorf("alice score = %d, want 100", payload.Scores[aliceID])
	}
	if payload.Scores[bobID] != 0 {
		t.Errorf("bob score = %d, want 0", payload.Scores[bobID])
	}
}

func TestServer_UnknownRoomCodeRejectsUpgrade(t *testing.T) {
	h := NewHub()
	srv := NewServer(h)
	ts := httptest.NewServer(http.HandlerFunc(srv.ServeWS))
	t.Cleanup(ts.Close)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	_, resp, err := gorilla.DefaultDialer.Dial(wsURL+"?code=ZZZZZZ&name=Nobody", nil)
	if err == nil {
		t.Fatal("expected dial to fail for unknown room code")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("response status = %v, want 404", resp)
	}
}

func mustWriteJSON(t *testing.T, conn *gorilla.Conn, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := conn.WriteMessage(gorilla.TextMessage, b); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
}
