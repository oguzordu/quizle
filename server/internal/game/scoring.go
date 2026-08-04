package game

import "time"

const (
	basePoints = 100

	// Speed tiers are deliberately generous at the top: a player who answers
	// anywhere in the first 40% of the time window gets full points. Only
	// real hesitation (past 90% of the window) costs the most. This keeps
	// the game decided mostly by who *knew* the answer rather than by
	// who reacted a few hundred milliseconds faster.
	speedTierFastThreshold   = 0.6 // remaining/total >= this -> used <=40% of time
	speedTierMediumThreshold = 0.3 // used <=70% of time
	speedTierSlowThreshold   = 0.1 // used <=90% of time

	speedTierFast   = basePoints
	speedTierMedium = 80
	speedTierSlow   = 60
	speedTierLast   = 40

	streakTierHigh   = 7
	streakTierMedium = 5
	streakTierLow    = 3

	streakMultHigh   = 1.5
	streakMultMedium = 1.35
	streakMultLow    = 1.2
	streakMultNone   = 1.0
)

// Score computes points for a single answer. Wrong or unanswered scores zero.
// Correct answers fall into one of four speed tiers (100/80/60/40) based on
// how much time remained when the answer was submitted, then get boosted by
// a streak multiplier that caps at streakMultHigh so a single miss can't make
// the game mathematically unwinnable.
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

	var raw float64
	switch {
	case speedFactor >= speedTierFastThreshold:
		raw = speedTierFast
	case speedFactor >= speedTierMediumThreshold:
		raw = speedTierMedium
	case speedFactor >= speedTierSlowThreshold:
		raw = speedTierSlow
	default:
		raw = speedTierLast
	}

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
