package game

const (
	basePoints   = 100
	fastestBonus = 10

	streakTierHigh   = 7
	streakTierMedium = 5
	streakTierLow    = 3

	streakMultHigh   = 1.5
	streakMultMedium = 1.35
	streakMultLow    = 1.2
	streakMultNone   = 1.0
)

// Score computes points for a single answer. Knowing the answer is the
// dominant factor: every correct answer is worth basePoints regardless of
// how many other players also got it right or how quickly. Being the
// fastest correct responder to a question only adds a small fastestBonus —
// enough to matter over many rounds, never enough that speed can beat
// knowledge. A streak multiplier (capped at streakMultHigh) applies on top.
func Score(correct, fastest bool, streak int) int {
	if !correct {
		return 0
	}

	raw := basePoints
	if fastest {
		raw += fastestBonus
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

	return int(float64(raw)*mult + 0.5)
}
