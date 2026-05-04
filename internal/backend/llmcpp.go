package backend

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/y9c/llmtop/internal/metrics"
)

type LLamaCPP struct{}

func (LLamaCPP) Name() string { return "llama.cpp" }

func (LLamaCPP) Detect(body string) bool {
	return strings.Contains(body, "llamacpp:") || strings.Contains(body, "llm_prompt_tokens") || strings.Contains(body, "slots_")
}

func (LLamaCPP) Parse(body string) (metrics.Snapshot, error) {
	s := metrics.Snapshot{Timestamp: time.Now()}

	rules := []struct {
		key string
		re  *regexp.Regexp
		set func(*metrics.Snapshot, float64)
	}{
		{"prompt_tokens_total", regexp.MustCompile(`(?:llamacpp:)?prompt_tokens_total\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptTokensTotal = v }},
		{"tokens_predicted_total", regexp.MustCompile(`(?:llamacpp:)?tokens_predicted_total\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.GenTokensTotal = v }},
		{"requests_processing", regexp.MustCompile(`(?:llamacpp:)?requests_processing\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.RunningReqs = v }},
		{"requests_deferred", regexp.MustCompile(`(?:llamacpp:)?requests_deferred\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.WaitingReqs = v }},
		{"n_decode_total", regexp.MustCompile(`(?:llamacpp:)?n_decode_total\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) {}},
	}

	for _, rule := range rules {
		if !strings.Contains(body, rule.key) {
			continue
		}
		matches := rule.re.FindStringSubmatch(body)
		if len(matches) >= 2 {
			if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
				rule.set(&s, v)
			}
		}
	}

	return s, nil
}
