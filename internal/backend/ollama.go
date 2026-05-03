package backend

import (
	"strings"
	"time"

	"github.com/y9c/llmtop/internal/metrics"
)

type Ollama struct{}

func (Ollama) Name() string { return "Ollama" }

func (Ollama) Detect(body string) bool { return strings.Contains(body, "ollama:") }

func (Ollama) Parse(body string) (metrics.Snapshot, error) {
	return metrics.Snapshot{Timestamp: time.Now()}, nil
}
