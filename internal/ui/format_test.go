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

func TestFmtMB(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{1024, "1.0GB"},
		{512, "512MB"},
	}
	for _, tt := range tests {
		got := fmtMB(tt.v)
		if got != tt.want {
			t.Errorf("fmtMB(%v): want %q, got %q", tt.v, tt.want, got)
		}
	}
}

func TestMin(t *testing.T) {
	if min(3, 5) != 3 {
		t.Fatal("min(3,5) should be 3")
	}
	if min(5, 3) != 3 {
		t.Fatal("min(5,3) should be 3")
	}
	if min(4, 4) != 4 {
		t.Fatal("min(4,4) should be 4")
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
