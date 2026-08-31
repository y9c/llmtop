package metrics

import (
	"testing"
)

func TestHistoryPushLen(t *testing.T) {
	h := NewHistory()
	for i := 0; i < 10; i++ {
		h.Push(float64(i))
	}
	vals := h.ValuesInto()
	if got := len(vals); got != 10 {
		t.Fatalf("len after 10 pushes: want 10, got %d", got)
	}
}

func TestHistoryRingOverwrite(t *testing.T) {
	h := NewHistory()
	for i := 0; i < 66; i++ {
		h.Push(float64(i))
	}
	vals := h.ValuesInto()
	if got := len(vals); got != historyLen {
		t.Fatalf("len after 70 pushes: want %d, got %d", historyLen, got)
	}
}

func TestHistoryRecentAvg(t *testing.T) {
	h := NewHistory()
	for i := 1; i <= 5; i++ {
		h.Push(float64(i)) // 1,2,3,4,5
	}
	if got := h.RecentAvg(3); got != 4 {
		t.Fatalf("RecentAvg(3): want 4, got %v", got)
	}
	if got := h.RecentAvg(10); got != 3 { // more than stored -> averages all 5
		t.Fatalf("RecentAvg(10): want 3, got %v", got)
	}
}

func TestHistoryRecentAvgZeroSamples(t *testing.T) {
	h := NewHistory()
	if got := h.RecentAvg(5); got != 0 {
		t.Fatalf("RecentAvg on empty: want 0, got %v", got)
	}
}

func TestHistoryChronologicalOrder(t *testing.T) {
	// Push 70 values (historyLen=60), so buffer wraps
	h := NewHistory()
	for i := 1; i <= 66; i++ {
		h.Push(float64(i))
	}
	vals := h.ValuesInto()
	if len(vals) != historyLen {
		t.Fatalf("want len %d, got %d", historyLen, len(vals))
	}
	// Most recent 60 values in chronological order: 7..66
	want := []float64{7, 8, 9, 10, 11, 12}
	for i, v := range vals[:6] {
		if v != want[i] {
			t.Fatalf("ValuesInto()[%d]: want %v, got %v", i, want[i], v)
		}
	}
	// Check end: 66
	if vals[59] != 66 {
		t.Fatalf("ValuesInto()[59]: want 66, got %v", vals[59])
	}
}
