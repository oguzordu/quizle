package game

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		name    string
		correct bool
		fastest bool
		streak  int
		want    int
	}{
		{name: "correct but not fastest gets flat base points", correct: true, fastest: false, streak: 0, want: 100},
		{name: "correct and fastest gets a small bonus on top", correct: true, fastest: true, streak: 0, want: 110},
		{name: "wrong answer scores zero even if fastest", correct: false, fastest: true, streak: 5, want: 0},
		{name: "wrong answer scores zero", correct: false, fastest: false, streak: 0, want: 0},
		{name: "streak of 3 applies 1.2x to the base", correct: true, fastest: false, streak: 3, want: 120},
		{name: "streak of 5 applies 1.35x to the base", correct: true, fastest: false, streak: 5, want: 135},
		{name: "streak of 7 applies 1.5x cap to the base", correct: true, fastest: false, streak: 7, want: 150},
		{name: "streak of 20 stays capped at 1.5x", correct: true, fastest: false, streak: 20, want: 150},
		{name: "streak of 2 stays at 1.0x below threshold", correct: true, fastest: false, streak: 2, want: 100},
		{name: "streak and fastest bonus combine", correct: true, fastest: true, streak: 7, want: 165},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.correct, tt.fastest, tt.streak)
			if got != tt.want {
				t.Errorf("Score(%v, %v, %d) = %d, want %d", tt.correct, tt.fastest, tt.streak, got, tt.want)
			}
		})
	}
}
