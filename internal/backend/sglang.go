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

// gaugeRe matches a labeled or bare Prometheus sample line.
func gaugeRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`sglang:` + name + `(?:\{[^}]*\})?\s+([\d.eE+-]+)`)
}

func (SGLang) Parse(body string) (metrics.Snapshot, error) {
	s := metrics.Snapshot{Timestamp: time.Now()}

	var (
		specAcceptLen float64 // windowed mean accepted tokens per verify pass
		specDraftCap  float64 // configured draft tokens per verify pass
	)

	rules := []struct {
		key string
		re  *regexp.Regexp
		set func(*metrics.Snapshot, float64)
	}{
		{"sglang:prompt_tokens_total", gaugeRe(`prompt_tokens_total`),
			func(s *metrics.Snapshot, v float64) { s.PromptTokensTotal = v }},
		{"sglang:generation_tokens_total", gaugeRe(`generation_tokens_total`),
			func(s *metrics.Snapshot, v float64) { s.GenTokensTotal = v }},
		// SGLang >= 0.5 renamed the request gauges (was num_running_requests / num_waiting_requests)
		{"sglang:num_running_reqs", gaugeRe(`num_running_reqs`),
			func(s *metrics.Snapshot, v float64) { s.RunningReqs = v }},
		{"sglang:num_queue_reqs", gaugeRe(`num_queue_reqs`),
			func(s *metrics.Snapshot, v float64) { s.WaitingReqs = v }},
		// token_usage is a 0-1 fraction; app.go pushes KVCacheUsagePct*100 into the KV chart
		{"sglang:token_usage", gaugeRe(`token_usage`),
			func(s *metrics.Snapshot, v float64) { s.KVCacheUsagePct = v }},
		{"sglang:cache_hit_rate", gaugeRe(`cache_hit_rate`),
			func(s *metrics.Snapshot, v float64) { s.PrefixCacheHits = v }},
		// speculative decoding: verify calls == number of draft batches (monotonic)
		{"sglang:spec_verify_calls_total", gaugeRe(`spec_verify_calls_total`),
			func(s *metrics.Snapshot, v float64) { s.SpecDraftsTotal = v }},
		{"sglang:spec_accept_length", gaugeRe(`spec_accept_length`),
			func(s *metrics.Snapshot, v float64) { specAcceptLen = v }},
		{"sglang:spec_num_draft_tokens", gaugeRe(`spec_num_draft_tokens`),
			func(s *metrics.Snapshot, v float64) { specDraftCap = v }},
		{"sglang:time_to_first_token_seconds_sum", gaugeRe(`time_to_first_token_seconds_sum`),
			func(s *metrics.Snapshot, v float64) { s.TTFTTotalS = v }},
		{"sglang:time_to_first_token_seconds_count", gaugeRe(`time_to_first_token_seconds_count`),
			func(s *metrics.Snapshot, v float64) { s.TTFTCount = v }},
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

	// SGLang exports no cumulative drafted/accepted token counters — only a
	// monotonic verify-call counter plus windowed mean accept length and the
	// configured draft size. Reconstruct cumulative totals so that deltas
	// (accept rate, rejected tokens) stay meaningful across samples.
	if s.SpecDraftsTotal > 0 && specDraftCap > 0 {
		s.SpecDraftToksTotal = s.SpecDraftsTotal * specDraftCap
		s.SpecAcceptedTotal = s.SpecDraftsTotal * specAcceptLen
	}

	// cache_hit_rate is a windowed ratio, not hits/queries counters; fake the
	// denominator so the UI's Hits/Queries*100 formula shows it directly.
	if s.PrefixCacheHits > 0 {
		s.PrefixCacheQueries = 1
	}

	return s, nil
}
