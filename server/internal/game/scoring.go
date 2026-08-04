package game

const (
	rankTopPoints   = 100
	rankFloorPoints = 50

	streakTierHigh   = 7
	streakTierMedium = 5
	streakTierLow    = 3

	streakMultHigh   = 1.5
	streakMultMedium = 1.35
	streakMultLow    = 1.2
	streakMultNone   = 1.0
)

// RankScore computes points for a correct answer based on the order it was
// submitted in relative to other correct answers to the same question. rank
// is 1-indexed (1 = first correct answer); totalCorrect is how many players
// answered correctly this round. The first correct responder earns
// rankTopPoints; the last-place correct responder still earns
// rankFloorPoints — knowing the answer always beats not knowing it, even for
// whoever was slowest among those who got it right. A streak multiplier
// (capped at streakMultHigh) applies on top, same as before.
func RankScore(rank, totalCorrect, streak int) int {
	if totalCorrect <= 0 {
		return 0
	}
	if rank < 1 {
		rank = 1
	}
	if rank > totalCorrect {
		rank = totalCorrect
	}

	raw := float64(rankTopPoints)
	if totalCorrect > 1 {
		frac := float64(rank-1) / float64(totalCorrect-1)
		raw -= float64(rankTopPoints-rankFloorPoints) * frac
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
