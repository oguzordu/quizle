package game

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		name    string
		correct bool
		streak  int
		want    int
	}{
		{name: "correct answer gets flat base points", correct: true, streak: 0, want: 100},
		{name: "wrong answer scores zero", correct: false, streak: 5, want: 0},
		{name: "streak of 3 applies 1.2x multiplier", correct: true, streak: 3, want: 120},
		{name: "streak of 5 applies 1.35x multiplier", correct: true, streak: 5, want: 135},
		{name: "streak of 7 applies 1.5x cap", correct: true, streak: 7, want: 150},
		{name: "streak of 20 stays capped at 1.5x", correct: true, streak: 20, want: 150},
		{name: "streak of 2 stays at 1.0x below threshold", correct: true, streak: 2, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.correct, tt.streak)
			if got != tt.want {
				t.Errorf("Score(%v, %d) = %d, want %d", tt.correct, tt.streak, got, tt.want)
			}
		})
	}
}
