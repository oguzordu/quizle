package game

import "time"

const (
	basePoints = 1000

	streakTierHigh   = 7
	streakTierMedium = 5
	streakTierLow    = 3

	streakMultHigh   = 1.5
	streakMultMedium = 1.35
	streakMultLow    = 1.2
	streakMultNone   = 1.0
)

// Score computes points for a single answer. Wrong or unanswered scores zero.
// Correct answers scale from 50% to 100% of basePoints based on how much time
// remained when the answer was submitted, then get boosted by a streak
// multiplier that caps at streakMultHigh so a single miss can't make the game
// mathematically unwinnable.
func Score(correct bool, timeTotal, timeRemaining time.Duration, streak int) int {
	if !correct || timeTotal <= 0 {
		return 0
	}

	speedFactor := float64(timeRemaining) / float64(timeTotal)
	if speedFactor < 0 {
		speedFactor = 0
	}
	if speedFactor > 1 {
		speedFactor = 1
	}

	raw := float64(basePoints) * (0.5 + 0.5*speedFactor)

	mult := streakMultNone
	switch {
	case streak >= streakTierHigh:
		mult = streakMultHigh
	case streak >= streakTierMedium:
		mult = streakMultMedium
	case streak >= streakTierLow:
		mult = streakMultLow
	}

	return int(raw*mult + 0.5)
}
