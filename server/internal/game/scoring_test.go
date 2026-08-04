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
			name:          "correct within first 40% of time (fast) gets top tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second, // used 0%
			streak:        0,
			want:          100,
		},
		{
			name:          "correct right at the 40% boundary still gets top tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 6 * time.Second, // used exactly 40%
			streak:        0,
			want:          100,
		},
		{
			name:          "correct just past the 40% boundary drops to second tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 5900 * time.Millisecond, // used just over 40%
			streak:        0,
			want:          80,
		},
		{
			name:          "correct at the 70% boundary still gets second tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 3 * time.Second, // used exactly 70%
			streak:        0,
			want:          80,
		},
		{
			name:          "correct just past the 70% boundary drops to third tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 2900 * time.Millisecond,
			streak:        0,
			want:          60,
		},
		{
			name:          "correct at the 90% boundary still gets third tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 1 * time.Second, // used exactly 90%
			streak:        0,
			want:          60,
		},
		{
			name:          "correct in the last 10% gets the bottom tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 500 * time.Millisecond,
			streak:        0,
			want:          40,
		},
		{
			name:          "correct answered at the very last instant still gets bottom tier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 0,
			streak:        0,
			want:          40,
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
			want:          120,
		},
		{
			name:          "streak of 5 applies 1.35x multiplier",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        5,
			want:          135,
		},
		{
			name:          "streak of 7 applies 1.5x cap",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        7,
			want:          150,
		},
		{
			name:          "streak of 20 stays capped at 1.5x",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        20,
			want:          150,
		},
		{
			name:          "streak of 2 stays at 1.0x below threshold",
			correct:       true,
			timeTotal:     10 * time.Second,
			timeRemaining: 10 * time.Second,
			streak:        2,
			want:          100,
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
