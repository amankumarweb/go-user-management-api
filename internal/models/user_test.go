package models

import (
	"testing"
	"time"
)

func TestCalculateAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		dob      time.Time
		expected int
	}{
		{
			name:     "exactly 30 years ago",
			dob:      now.AddDate(-30, 0, 0),
			expected: 30,
		},
		{
			name:     "birthday is today (25 years)",
			dob:      time.Date(now.Year()-25, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			expected: 25,
		},
		{
			name:     "birthday is tomorrow (still 24)",
			dob:      now.AddDate(-25, 0, 1),
			expected: 24,
		},
		{
			name:     "birthday was yesterday (already 25)",
			dob:      now.AddDate(-25, 0, -1),
			expected: 25,
		},
		{
			name:     "newborn (born today)",
			dob:      now,
			expected: 0,
		},
		{
			name:     "one year old born yesterday",
			dob:      now.AddDate(-1, 0, -1),
			expected: 1,
		},
		{
			name:     "known fixed date",
			dob:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: calcExpected(now, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateAge(tt.dob)
			if got != tt.expected {
				t.Errorf("CalculateAge(%v) = %d, want %d", tt.dob.Format("2006-01-02"), got, tt.expected)
			}
		})
	}
}

// calcExpected is a helper that computes age independently to verify
// the known-date test case stays correct regardless of when the test runs.
func calcExpected(now, dob time.Time) int {
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() ||
		(now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age
}
