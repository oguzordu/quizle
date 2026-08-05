package hub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"
	"github.com/oguzordu/quizle/internal/game"
)

// TestServer_ServeWS_ClosesConnectionOnOversizedMessage guards against a
// client sending an unbounded WebSocket frame to exhaust server memory: the
// connection must be capped and dropped, not buffered without limit.
func TestServer_ServeWS_ClosesConnectionOnOversizedMessage(t *testing.T) {
	h := NewHub()
	questions := []game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: 5 * time.Second}}
	code, actor := h.CreateRoom(questions, time.Second)
	t.Cleanup(actor.Stop)

	srv := NewServer(h)
	ts := httptest.NewServer(http.HandlerFunc(srv.ServeWS))
	t.Cleanup(ts.Close)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	conn, aliceID := dialPlayer(t, wsURL, code, "Alice")
	_ = aliceID

	oversized := make([]byte, wsMaxMessageBytes+1)
	if err := conn.WriteMessage(gorilla.TextMessage, oversized); err != nil {
		// A write-side error is an acceptable way for this to surface too.
		return
	}

	conn.SetReadDeadline(time.Now().Add(testTimeout))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected connection to close after an oversized message, got no error")
	}
}
