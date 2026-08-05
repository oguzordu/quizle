package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

func TestServer_CreateRoomHandler_rateLimitsPerIP(t *testing.T) {
	h := NewHub()
	srv := NewServer(h)
	srv.SetDefaultQuestions([]game.Question{{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 0, Duration: time.Second}}, time.Second)
	srv.roomLimiter = newRateLimiter(3, time.Minute)

	newReq := func(remoteAddr string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
		req.RemoteAddr = remoteAddr
		w := httptest.NewRecorder()
		srv.CreateRoomHandler(w, req)
		return w
	}

	for i := 0; i < 3; i++ {
		if w := newReq("203.0.113.5:1111"); w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, w.Code)
		}
	}

	if w := newReq("203.0.113.5:2222"); w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request from same IP: status = %d, want 429", w.Code)
	}

	// A different IP has its own budget and isn't blocked by the first one's.
	if w := newReq("198.51.100.9:3333"); w.Code != http.StatusOK {
		t.Errorf("request from a different IP: status = %d, want 200", w.Code)
	}
}

func TestClientIP_PrefersFlyClientIPHeaderOverRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	req.RemoteAddr = "10.0.0.1:1234" // the proxy's own address
	req.Header.Set("Fly-Client-IP", "203.0.113.42")

	if got := clientIP(req); got != "203.0.113.42" {
		t.Errorf("clientIP() = %q, want the Fly-Client-IP header value", got)
	}
}

func TestClientIP_FallsBackToRemoteAddrWithoutProxyHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	req.RemoteAddr = "203.0.113.42:5555"

	if got := clientIP(req); got != "203.0.113.42" {
		t.Errorf("clientIP() = %q, want 203.0.113.42", got)
	}
}
