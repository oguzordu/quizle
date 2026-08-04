package game

import "testing"

func TestRankScore(t *testing.T) {
	tests := []struct {
		name         string
		rank         int
		totalCorrect int
		streak       int
		want         int
	}{
		{name: "sole correct answerer gets top points", rank: 1, totalCorrect: 1, streak: 0, want: 100},
		{name: "first of two correct gets top points", rank: 1, totalCorrect: 2, streak: 0, want: 100},
		{name: "second (last) of two correct gets the floor", rank: 2, totalCorrect: 2, streak: 0, want: 50},
		{name: "first of five correct gets top points", rank: 1, totalCorrect: 5, streak: 0, want: 100},
		{name: "middle of five correct gets an interpolated value", rank: 3, totalCorrect: 5, streak: 0, want: 75},
		{name: "last of five correct still gets the floor, never zero", rank: 5, totalCorrect: 5, streak: 0, want: 50},
		{name: "nobody answered correctly scores zero", rank: 1, totalCorrect: 0, streak: 0, want: 0},
		{name: "streak of 3 applies 1.2x multiplier", rank: 1, totalCorrect: 1, streak: 3, want: 120},
		{name: "streak of 5 applies 1.35x multiplier", rank: 1, totalCorrect: 1, streak: 5, want: 135},
		{name: "streak of 7 applies 1.5x cap", rank: 1, totalCorrect: 1, streak: 7, want: 150},
		{name: "streak of 20 stays capped at 1.5x", rank: 1, totalCorrect: 1, streak: 20, want: 150},
		{name: "streak of 2 stays at 1.0x below threshold", rank: 1, totalCorrect: 1, streak: 2, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RankScore(tt.rank, tt.totalCorrect, tt.streak)
			if got != tt.want {
				t.Errorf("RankScore(%d, %d, %d) = %d, want %d", tt.rank, tt.totalCorrect, tt.streak, got, tt.want)
			}
		})
	}
}
