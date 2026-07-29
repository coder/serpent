package serpent_test

import (
	"net/url"
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

func TestURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		expectedHost string
		expectedStr  string
		wantErr      bool
	}{
		{
			name:         "BareIPv6WithPort",
			input:        "http://2600:2222::22:8080",
			expectedHost: "[2600:2222::22]:8080",
			expectedStr:  "http://[2600:2222::22]:8080",
		},
		{
			name:         "BareIPv6WithPortFromIssue",
			input:        "http://2600:2222:2222:2222:22::22:8080",
			expectedHost: "[2600:2222:2222:2222:22::22]:8080",
			expectedStr:  "http://[2600:2222:2222:2222:22::22]:8080",
		},
		{
			name:         "AlreadyBracketedIPv6",
			input:        "http://[::1]:8080",
			expectedHost: "[::1]:8080",
			expectedStr:  "http://[::1]:8080",
		},
		{
			name:         "IPv4WithPort",
			input:        "http://127.0.0.1:8080",
			expectedHost: "127.0.0.1:8080",
			expectedStr:  "http://127.0.0.1:8080",
		},
		{
			name:         "HostnameWithPort",
			input:        "http://example.com:8080",
			expectedHost: "example.com:8080",
			expectedStr:  "http://example.com:8080",
		},
		{
			name:         "HostnameWithPath",
			input:        "http://example.com/path",
			expectedHost: "example.com",
			expectedStr:  "http://example.com/path",
		},
		{
			name:         "BareIPv6WithPath",
			input:        "http://2600:2222::22:8080/path",
			expectedHost: "[2600:2222::22]:8080",
			expectedStr:  "http://[2600:2222::22]:8080/path",
		},
		{
			// Authority is terminated by '?' with no leading '/'.
			name:         "BareIPv6WithQuery",
			input:        "http://2600:2222::22:8080?foo=bar",
			expectedHost: "[2600:2222::22]:8080",
			expectedStr:  "http://[2600:2222::22]:8080?foo=bar",
		},
		{
			// Authority is terminated by '#' with no leading '/'.
			name:         "BareIPv6WithFragment",
			input:        "http://2600:2222::22:8080#frag",
			expectedHost: "[2600:2222::22]:8080",
			expectedStr:  "http://[2600:2222::22]:8080#frag",
		},
		{
			// userinfo must not be mistaken for part of the host.
			name:         "BareIPv6WithUserinfo",
			input:        "http://user:pass@2600:2222::22:8080/path",
			expectedHost: "[2600:2222::22]:8080",
			expectedStr:  "http://user:pass@[2600:2222::22]:8080/path",
		},
		{
			// IPv6 without a port must be left alone; url.Parse handles
			// bracketed literals, and bare literals without ports are
			// ambiguous and should error naturally.
			name:         "AlreadyBracketedIPv6NoPort",
			input:        "http://[::1]",
			expectedHost: "[::1]",
			expectedStr:  "http://[::1]",
		},
		{
			// A hostname containing colons that isn't a valid IPv6
			// literal should not be auto-bracketed; url.Parse will
			// return its own error.
			name:    "InvalidIPv6ShapedHost",
			input:   "http://not:an:ipv6:literal:8080",
			wantErr: true,
		},
		{
			name:    "Empty",
			input:   "",
			wantErr: false, // url.Parse("") succeeds and yields an empty URL
		},
		{
			name:    "Malformed",
			input:   "http://[::1:8080", // unterminated bracket
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var u serpent.URL
			err := u.Set(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.expectedHost != "" {
				require.Equal(t, tt.expectedHost, u.Value().Host)
			}
			if tt.expectedStr != "" {
				require.Equal(t, tt.expectedStr, u.String())
			}

			// Verify String() returns a value that is still parseable and
			// round-trips to the same result.
			str := u.String()
			var u2 serpent.URL
			err = u2.Set(str)
			require.NoError(t, err)
			require.Equal(t, u.Value().String(), u2.Value().String(), "String() should return a parseable value")
		})
	}
}

func TestURLOf(t *testing.T) {
	t.Parallel()

	base := &url.URL{Scheme: "http", Host: "example.com:8080"}
	u := serpent.URLOf(base)
	require.NotNil(t, u)
	require.Equal(t, base, u.Value())
}
