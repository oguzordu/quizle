package game

const (
	basePoints = 100

	// FastestPlayerBonus is a one-time bonus added at game end (not per
	// question) to whoever had the best average response time among their
	// correct answers. Speed is rewarded as a separate, bounded prize
	// rather than a per-question multiplier, so it can never make up for
	// knowing fewer answers than an opponent.
	FastestPlayerBonus = 50

	streakTierHigh   = 7
	streakTierMedium = 5
	streakTierLow    = 3

	streakMultHigh   = 1.5
	streakMultMedium = 1.35
	streakMultLow    = 1.2
	streakMultNone   = 1.0
)

// Score computes points for a single answer. Every correct answer is worth
// the same basePoints, regardless of speed — knowing the answer is what
// counts. A streak multiplier (capped at streakMultHigh) applies on top.
func Score(correct bool, streak int) int {
	if !correct {
		return 0
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

	return int(float64(basePoints)*mult + 0.5)
}
