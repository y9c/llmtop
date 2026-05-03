package backend

import (
	"strings"
	"time"

	"github.com/changye/llmtop/internal/metrics"
)

type SGLang struct{}

func (SGLang) Name() string { return "SGLang" }

func (SGLang) Detect(body string) bool { return strings.Contains(body, "sgl:") }

func (SGLang) Parse(body string) (metrics.Snapshot, error) {
	return metrics.Snapshot{Timestamp: time.Now()}, nil
}
