package metrics

import (
	"sync"
)

// Ring buffer for 60 seconds of metric history.
const historyLen = 60

type History struct {
	mu    sync.Mutex
	buf   [historyLen]float64
	count int
}

func NewHistory() *History { return &History{} }

func (h *History) Push(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buf[h.count%historyLen] = v
	h.count++
}

// ValuesInto returns the stored samples in chronological order, oldest first.
func (h *History) ValuesInto() []float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := h.count
	if n > historyLen {
		n = historyLen
	}
	if n == 0 {
		return nil
	}
	out := make([]float64, n)
	start := h.count - n
	for i := 0; i < n; i++ {
		out[i] = h.buf[(start+i)%historyLen]
	}
	return out
}

// RecentAvg returns the mean of the most recent n samples (fewer if not enough yet).
func (h *History) RecentAvg(n int) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	if n > historyLen {
		n = historyLen
	}
	if n > h.count {
		n = h.count
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += h.buf[(h.count-1-i)%historyLen]
	}
	return sum / float64(n)
}
