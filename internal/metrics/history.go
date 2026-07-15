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

func (h *History) ValuesInto([]float64) []float64 {
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
