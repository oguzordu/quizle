package hub

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

func TestEncodeEvent_QuestionStarted(t *testing.T) {
	deadline := time.Date(2026, 1, 1, 0, 0, 10, 0, time.UTC)
	ev := game.QuestionStarted{
		Question: game.Question{ID: "q1", Choices: []string{"a", "b", "c", "d"}, Correct: 1},
		Deadline: deadline,
	}

	raw := EncodeEvent(ev)

	var msg ServerMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if msg.Type != "question_started" {
		t.Fatalf("Type = %q, want question_started", msg.Type)
	}

	var payload questionStartedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.QuestionID != "q1" {
		t.Errorf("QuestionID = %q, want q1", payload.QuestionID)
	}
	if len(payload.Choices) != 4 {
		t.Errorf("len(Choices) = %d, want 4", len(payload.Choices))
	}
	if !payload.Deadline.Equal(deadline) {
		t.Errorf("Deadline = %v, want %v", payload.Deadline, deadline)
	}

	// The correct choice must never be leaked to the client before reveal.
	raw2 := string(raw)
	if containsCorrectField(raw2) {
		t.Errorf("question_started payload leaks the correct answer: %s", raw2)
	}
}

func TestEncodeEvent_AnswerAccepted(t *testing.T) {
	ev := game.AnswerAccepted{PlayerID: "alice", Correct: true}

	raw := EncodeEvent(ev)

	var msg ServerMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if msg.Type != "answer_accepted" {
		t.Fatalf("Type = %q, want answer_accepted", msg.Type)
	}

	var payload answerAcceptedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.PlayerID != "alice" || !payload.Correct {
		t.Errorf("payload = %+v, want alice/true", payload)
	}
}

func TestEncodeEvent_QuestionRevealed(t *testing.T) {
	ev := game.QuestionRevealed{
		CorrectChoice: 2,
		PointsAwarded: map[game.PlayerID]int{"alice": 100},
		Scores:        map[game.PlayerID]int{"alice": 950, "bob": 0},
	}

	raw := EncodeEvent(ev)

	var msg ServerMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if msg.Type != "question_revealed" {
		t.Fatalf("Type = %q, want question_revealed", msg.Type)
	}

	var payload questionRevealedPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.CorrectChoice != 2 {
		t.Errorf("CorrectChoice = %d, want 2", payload.CorrectChoice)
	}
	if payload.PointsAwarded["alice"] != 100 {
		t.Errorf("PointsAwarded[alice] = %d, want 100", payload.PointsAwarded["alice"])
	}
	if payload.Scores["alice"] != 950 {
		t.Errorf("Scores[alice] = %d, want 950", payload.Scores["alice"])
	}
}

func TestEncodeEvent_GameFinished(t *testing.T) {
	ev := game.GameFinished{FinalScores: map[game.PlayerID]int{"alice": 1900}}

	raw := EncodeEvent(ev)

	var msg ServerMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if msg.Type != "game_finished" {
		t.Fatalf("Type = %q, want game_finished", msg.Type)
	}
}

func TestDecodeClientMessage_SubmitAnswer(t *testing.T) {
	raw := []byte(`{"type":"submit_answer","choice":2}`)

	msg, err := DecodeClientMessage(raw)
	if err != nil {
		t.Fatalf("DecodeClientMessage: %v", err)
	}
	sa, ok := msg.(SubmitAnswerCommand)
	if !ok {
		t.Fatalf("got %T, want SubmitAnswerCommand", msg)
	}
	if sa.Choice != 2 {
		t.Errorf("Choice = %d, want 2", sa.Choice)
	}
}

func TestDecodeClientMessage_UnknownTypeErrors(t *testing.T) {
	raw := []byte(`{"type":"nonsense"}`)

	_, err := DecodeClientMessage(raw)
	if err == nil {
		t.Fatal("expected error for unknown message type, got nil")
	}
}

func containsCorrectField(s string) bool {
	return jsonContains(s, `"correct"`) || jsonContains(s, `"Correct"`)
}

func jsonContains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
