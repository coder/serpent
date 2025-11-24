package serpent_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	serpent "github.com/coder/serpent"
)

func TestDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Standard time.Duration formats (should still work)
		{
			name:     "Nanoseconds",
			input:    "100ns",
			expected: 100 * time.Nanosecond,
		},
		{
			name:     "Microseconds",
			input:    "100us",
			expected: 100 * time.Microsecond,
		},
		{
			name:     "Milliseconds",
			input:    "100ms",
			expected: 100 * time.Millisecond,
		},
		{
			name:     "Seconds",
			input:    "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "Minutes",
			input:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "Hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "Combined",
			input:    "1h30m",
			expected: 90 * time.Minute,
		},
		// New formats with days and weeks support
		{
			name:     "Days",
			input:    "1d",
			expected: 24 * time.Hour,
		},
		{
			name:     "MultipleDays",
			input:    "7d",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "Weeks",
			input:    "1w",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "MultipleWeeks",
			input:    "2w",
			expected: 14 * 24 * time.Hour,
		},
		{
			name:     "CombinedWithDays",
			input:    "1d12h",
			expected: 36 * time.Hour,
		},
		{
			name:     "CombinedWithWeeks",
			input:    "1w2d",
			expected: (7 + 2) * 24 * time.Hour,
		},
		{
			name:     "ComplexCombination",
			input:    "2w3d4h5m6s",
			expected: (14 + 3) * 24 * time.Hour + 4*time.Hour + 5*time.Minute + 6*time.Second,
		},
		// Error cases
		{
			name:    "Invalid",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "Empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var d serpent.Duration
			err := d.Set(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, d.Value())

			// Verify String() returns a parseable value
			str := d.String()
			var d2 serpent.Duration
			err = d2.Set(str)
			require.NoError(t, err)
			require.Equal(t, d.Value(), d2.Value(), "String() should return a parseable value")
		})
	}
}

func TestDurationOf(t *testing.T) {
	t.Parallel()

	td := 5 * time.Minute
	d := serpent.DurationOf(&td)
	require.NotNil(t, d)
	require.Equal(t, td, d.Value())

	// Test modification through pointer
	newVal := 10 * time.Minute
	*d = serpent.Duration(newVal)
	require.Equal(t, newVal, time.Duration(td))
}
