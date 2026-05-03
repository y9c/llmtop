package backend

import (
	"strings"
	"time"

	"github.com/changye/llmtop/internal/metrics"
)

type LLamaCPP struct{}

func (LLamaCPP) Name() string { return "llama.cpp" }

func (LLamaCPP) Detect(body string) bool {
	return strings.Contains(body, "llm_prompt_tokens") || strings.Contains(body, "slots_")
}

func (LLamaCPP) Parse(body string) (metrics.Snapshot, error) {
	return metrics.Snapshot{Timestamp: time.Now()}, nil
}
