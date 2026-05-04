package backend

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/y9c/llmtop/internal/metrics"
)

type SGLang struct{}

func (SGLang) Name() string { return "SGLang" }

func (SGLang) Detect(body string) bool { return strings.Contains(body, "sglang:") }

func (SGLang) Parse(body string) (metrics.Snapshot, error) {
	s := metrics.Snapshot{Timestamp: time.Now()}

	rules := []struct {
		key string
		re  *regexp.Regexp
		set func(*metrics.Snapshot, float64)
	}{
		{"sglang:prompt_tokens_total", regexp.MustCompile(`sglang:prompt_tokens_total(?:\{[^}]*\})?\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptTokensTotal = v }},
		{"sglang:generation_tokens_total", regexp.MustCompile(`sglang:generation_tokens_total(?:\{[^}]*\})?\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.GenTokensTotal = v }},
		{"sglang:num_running_requests", regexp.MustCompile(`sglang:num_running_requests(?:\{[^}]*\})?\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.RunningReqs = v }},
		{"sglang:num_waiting_requests", regexp.MustCompile(`sglang:num_waiting_requests(?:\{[^}]*\})?\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.WaitingReqs = v }},
		{"sglang:prefix_cache_hit_rate", regexp.MustCompile(`sglang:prefix_cache_hit_rate(?:\{[^}]*\})?\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.KVCacheUsagePct = v * 100 }},
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
