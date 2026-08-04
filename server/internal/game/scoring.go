package game

import "time"

const (
	basePoints    = 100
	speedBonusMax = 50

	// FastestPlayerBonus is a one-time bonus added at game end (not per
	// question) to whoever had the best average response time among their
	// correct answers, on top of the per-question speed bonus below.
	FastestPlayerBonus = 50

	streakTierHigh   = 7
	streakTierMedium = 5
	streakTierLow    = 3

	streakMultHigh   = 1.5
	streakMultMedium = 1.35
	streakMultLow    = 1.2
	streakMultNone   = 1.0
)

// Score computes points for a single answer. Every correct answer is
// guaranteed basePoints regardless of speed — knowing the answer always pays
// off — plus a speed bonus of up to speedBonusMax that scales continuously
// with how much time was left when the answer was submitted. A streak
// multiplier (capped at streakMultHigh) applies to the combined total.
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

	raw := float64(basePoints) + float64(speedBonusMax)*speedFactor

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
