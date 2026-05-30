package sync

import (
	"testing"
	"time"
)

func TestResolveRemoteWins(t *testing.T) {
	base := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name       string
		localMtime time.Time
		remoteTime time.Time
		wantRemote bool
	}{
		{"remote newer wins", base, base.Add(time.Minute), true},
		{"local newer wins", base.Add(time.Minute), base, false},
		{"tie prefers remote", base, base, true},
		{"remote much older loses", base, base.Add(-time.Hour), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveRemoteWins(tc.localMtime, tc.remoteTime); got != tc.wantRemote {
				t.Errorf("resolveRemoteWins = %v, want %v", got, tc.wantRemote)
			}
		})
	}
}

func TestIsHidden(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{".DS_Store", true},
		{"docs/.DS_Store", true},
		{".lumo/state.db", true},
		{".git/config", true},
		{"docs/a.txt", false},
		{"a/b/c.txt", false},
		{".", false},
		{"notlumo/a", false},
	}
	for _, tc := range cases {
		if got := isHidden(tc.rel); got != tc.want {
			t.Errorf("isHidden(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
}
