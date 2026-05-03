package ui

import "testing"

func TestFmtNum(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{1500, "1.5K"},
		{1500000, "1.50M"},
		{500, "500"},
		{0, "0"},
	}
	for _, tt := range tests {
		got := fmtNum(tt.v)
		if got != tt.want {
			t.Errorf("fmtNum(%v): want %q, got %q", tt.v, tt.want, got)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	if got := formatDuration(90 * 1e9); got != "1m 30s" {
		t.Fatalf("90s: want %q, got %q", "1m 30s", got)
	}
	if got := formatDuration(3700 * 1e9); got != "1h 01m" {
		t.Fatalf("3700s: want %q, got %q", "1h 01m", got)
	}
}
