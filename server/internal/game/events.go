package game

import "time"

// PlayerJoined is emitted when a new player is enrolled during the lobby
// phase, so already-connected clients can show a live roster.
type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
	Avatar   string
}

// QuestionStarted is emitted when a new question begins.
type QuestionStarted struct {
	Question Question
	Deadline time.Time
}

// AnswerAccepted is emitted in direct response to a player's SubmitAnswer
// call, confirming correctness immediately. Points aren't included here:
// they're computed at QuestionRevealed.
type AnswerAccepted struct {
	PlayerID PlayerID
	Correct  bool
}

// QuestionRevealed is emitted when a question's answer window closes, either
// because everyone answered or the deadline passed. PointsAwarded holds only
// the points earned this round (for a "+80" style popup); Scores holds each
// player's running total.
type QuestionRevealed struct {
	CorrectChoice int
	PointsAwarded map[PlayerID]int
	Scores        map[PlayerID]int
}

// GameFinished is emitted when the last question has been revealed and
// NextQuestion is called with no questions remaining. If HasFastestPlayer is
// true, FastestPlayerID already received FastestPlayerBonus and that bonus
// is reflected in FinalScores.
type GameFinished struct {
	FinalScores      map[PlayerID]int
	HasFastestPlayer bool
	FastestPlayerID  PlayerID
}
