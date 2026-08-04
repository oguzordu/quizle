package game

import "time"

// QuestionStarted is emitted when a new question begins.
type QuestionStarted struct {
	Question Question
	Deadline time.Time
}

// AnswerAccepted is emitted in direct response to a player's SubmitAnswer call.
type AnswerAccepted struct {
	PlayerID      PlayerID
	Correct       bool
	PointsAwarded int
}

// QuestionRevealed is emitted when a question's answer window closes, either
// because everyone answered or the deadline passed.
type QuestionRevealed struct {
	CorrectChoice int
	Scores        map[PlayerID]int
}

// GameFinished is emitted when the last question has been revealed and
// NextQuestion is called with no questions remaining.
type GameFinished struct {
	FinalScores map[PlayerID]int
}
