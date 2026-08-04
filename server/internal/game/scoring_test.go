package game

import (
	"testing"
	"time"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name          string
		correct       bool
		timeTotal     time.Duration
		timeRemaining time.Duration
		streak        int
		want          int
	}{
		{
			name:          "correct answered instantly gets full points",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        0,
			want:          1000,
		},
		{
			name:          "correct answered at the last instant gets half points",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 0,
			streak:        0,
			want:          500,
		},
		{
			name:          "correct answered halfway gets three quarters",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 5 * time.Second,
			streak:        0,
			want:          750,
		},
		{
			name:          "wrong answer scores zero regardless of speed",
			correct:       false,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        5,
			want:          0,
		},
		{
			name:          "streak of 3 applies 1.2x multiplier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        3,
			want:          1200,
		},
		{
			name:          "streak of 5 applies 1.35x multiplier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        5,
			want:          1350,
		},
		{
			name:          "streak of 7 applies 1.5x cap",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        7,
			want:          1500,
		},
		{
			name:          "streak of 20 stays capped at 1.5x",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        20,
			want:          1500,
		},
		{
			name:          "streak of 2 stays at 1.0x below threshold",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        2,
			want:          1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Score(tt.correct, tt.timeTotal, tt.timeRemaining, tt.streak)
			if got != tt.want {
				t.Errorf("Score(%v, %v, %v, %d) = %d, want %d",
					tt.correct, tt.timeTotal, tt.timeRemaining, tt.streak, got, tt.want)
			}
		})
	}
}
