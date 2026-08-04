package hub

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oguzordu/quizle/internal/game"
)

// ServerMessage is the wire envelope for every message the server sends to a
// client. Payload is deferred decoding so callers can dispatch on Type first.
type ServerMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type playerJoinedPayload struct {
	PlayerID game.PlayerID `json:"player_id"`
	Name     string        `json:"name"`
	Avatar   string        `json:"avatar,omitempty"`
}

type questionStartedPayload struct {
	QuestionID string    `json:"question_id"`
	Text       string    `json:"text"`
	Image      string    `json:"image,omitempty"`
	Choices    []string  `json:"choices"`
	Deadline   time.Time `json:"deadline"`
	Index      int       `json:"index"`
	Total      int       `json:"total"`
}

type answerAcceptedPayload struct {
	PlayerID game.PlayerID `json:"player_id"`
	Correct  bool          `json:"correct"`
}

type questionRevealedPayload struct {
	CorrectChoice int                   `json:"correct_choice"`
	PointsAwarded map[game.PlayerID]int `json:"points_awarded"`
	Scores        map[game.PlayerID]int `json:"scores"`
}

type gameFinishedPayload struct {
	FinalScores      map[game.PlayerID]int `json:"final_scores"`
	HasFastestPlayer bool                  `json:"has_fastest_player"`
	FastestPlayerID  game.PlayerID         `json:"fastest_player_id,omitempty"`
}

// joinedPayload is sent once, immediately after a WebSocket connection joins
// a room, so the client learns its assigned identity and reconnect token.
type joinedPayload struct {
	PlayerID game.PlayerID `json:"player_id"`
	Token    string        `json:"token"`
}

func encodeJoined(id game.PlayerID, token string) []byte {
	payloadBytes, _ := json.Marshal(joinedPayload{PlayerID: id, Token: token})
	out, _ := json.Marshal(ServerMessage{Type: "joined", Payload: payloadBytes})
	return out
}

// EncodeEvent serializes a game.Event into its wire representation. Unknown
// event types produce a message with an empty payload rather than panicking,
// since a malformed broadcast should never crash the server.
func EncodeEvent(ev game.Event) []byte {
	var msgType string
	var payload any

	switch e := ev.(type) {
	case game.PlayerJoined:
		msgType = "player_joined"
		payload = playerJoinedPayload{PlayerID: e.PlayerID, Name: e.Name, Avatar: e.Avatar}
	case game.QuestionStarted:
		msgType = "question_started"
		payload = questionStartedPayload{
			QuestionID: e.Question.ID,
			Text:       e.Question.Text,
			Image:      e.Question.Image,
			Choices:    e.Question.Choices,
			Deadline:   e.Deadline,
			Index:      e.Index,
			Total:      e.Total,
		}
	case game.AnswerAccepted:
		msgType = "answer_accepted"
		payload = answerAcceptedPayload{
			PlayerID: e.PlayerID,
			Correct:  e.Correct,
		}
	case game.QuestionRevealed:
		msgType = "question_revealed"
		payload = questionRevealedPayload{
			CorrectChoice: e.CorrectChoice,
			PointsAwarded: e.PointsAwarded,
			Scores:        e.Scores,
		}
	case game.GameReset:
		msgType = "game_reset"
		payload = struct{}{}
	case game.GameFinished:
		msgType = "game_finished"
		payload = gameFinishedPayload{
			FinalScores:      e.FinalScores,
			HasFastestPlayer: e.HasFastestPlayer,
			FastestPlayerID:  e.FastestPlayerID,
		}
	default:
		msgType = "unknown"
		payload = struct{}{}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		payloadBytes = []byte(`{}`)
	}
	out, err := json.Marshal(ServerMessage{Type: msgType, Payload: payloadBytes})
	if err != nil {
		return []byte(`{"type":"unknown","payload":{}}`)
	}
	return out
}

// ClientCommand is a marker interface for decoded inbound messages.
type ClientCommand interface{}

// SubmitAnswerCommand is sent by a client choosing an answer for the current
// question.
type SubmitAnswerCommand struct {
	Choice int
}

type clientMessageEnvelope struct {
	Type   string `json:"type"`
	Choice int    `json:"choice"`
}

// DecodeClientMessage parses a raw inbound WebSocket frame into a
// ClientCommand. Unknown types are rejected so a typo'd client can't silently
// no-op.
func DecodeClientMessage(raw []byte) (ClientCommand, error) {
	var env clientMessageEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode client message: %w", err)
	}

	switch env.Type {
	case "submit_answer":
		return SubmitAnswerCommand{Choice: env.Choice}, nil
	default:
		return nil, fmt.Errorf("unknown client message type %q", env.Type)
	}
}
