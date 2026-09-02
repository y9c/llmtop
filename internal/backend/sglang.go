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

func (SGLang) Detect(body string) bool {
	if strings.Contains(body, "sglang:") {
		return true
	}
	// Real SGLang can serve vLLM-compatible metric names, with or without the
	// sglang: prefix and with pre-0.5 gauge names, so disambiguate by
	// SGLang-only gauges that a genuine vLLM never exports.
	for _, m := range sglangOnlyMarkers {
		if strings.Contains(body, m) {
			return true
		}
	}
	return false
}

// Matched as bare substrings so prefixed (sglang:token_usage) and unprefixed
// bodies both detect. None of these names is exported by a genuine vLLM.
var sglangOnlyMarkers = []string{
	"token_usage",
	"cache_hit_rate",
	"num_running_reqs",
	"num_running_requests",
	"num_queue_reqs",
	"num_waiting_requests",
	"spec_accept_length",
	"gen_throughput",
}

// gaugeRe matches a labeled or bare Prometheus sample line.
func gaugeRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?:sglang:)?` + name + `(?:\{[^}]*\})?\s+([\d.eE+-]+)`)
}

// sumGauge totals every sample of a metric. SGLang exports token counters as
// multiple labeled series (e.g. is_streaming="true"/"false"); the streaming
// series is the only one that advances during generation, so both must count.
func sumGauge(name string, body string) float64 {
	re := gaugeRe(name)
	total := 0.0
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			total += v
		}
	}
	return total
}

func (SGLang) Parse(body string) (metrics.Snapshot, error) {
	s := metrics.Snapshot{Timestamp: time.Now()}

	rules := []struct {
		key string
		re  *regexp.Regexp
		set func(*metrics.Snapshot, float64)
	}{
		{"prompt_tokens_total", gaugeRe(`prompt_tokens_total`),
			func(s *metrics.Snapshot, v float64) { s.PromptTokensTotal = sumGauge(`prompt_tokens_total`, body) }},
		{"generation_tokens_total", gaugeRe(`generation_tokens_total`),
			func(s *metrics.Snapshot, v float64) { s.GenTokensTotal = sumGauge(`generation_tokens_total`, body) }},
		// SGLang >= 0.5 exports num_running_reqs/num_queue_reqs; older
		// releases used num_running_requests/num_waiting_requests.
		{"num_", gaugeRe(`num_(?:running_reqs|running_requests)`),
			func(s *metrics.Snapshot, v float64) { s.RunningReqs = v }},
		{"num_", gaugeRe(`num_(?:queue_reqs|waiting_requests)`),
			func(s *metrics.Snapshot, v float64) { s.WaitingReqs = v }},
		// token_usage is a 0-1 fraction; app.go pushes KVCacheUsagePct*100 into the KV chart
		{"token_usage", gaugeRe(`token_usage`),
			func(s *metrics.Snapshot, v float64) { s.KVCacheUsagePct = v }},
		{"cache_hit_rate", gaugeRe(`cache_hit_rate`),
			func(s *metrics.Snapshot, v float64) { s.PrefixCacheHits = v }},
		// speculative decoding: verify calls == number of draft batches (monotonic)
		{"spec_verify_calls_total", gaugeRe(`spec_verify_calls_total`),
			func(s *metrics.Snapshot, v float64) { s.SpecDraftsTotal = v }},
		// acceptance gauges are server-side windowed means; the server exports
		// no cumulative drafted/accepted counters, so never reconstruct them
		{"spec_accept_length", gaugeRe(`spec_accept_length`),
			func(s *metrics.Snapshot, v float64) { s.SpecAcceptLen = v }},
		{"spec_accept_rate", gaugeRe(`spec_accept_rate`),
			func(s *metrics.Snapshot, v float64) { s.SpecAcceptRate = v }},
		{"time_to_first_token_seconds_sum", gaugeRe(`time_to_first_token_seconds_sum`),
			func(s *metrics.Snapshot, v float64) { s.TTFTTotalS = v }},
		{"time_to_first_token_seconds_count", gaugeRe(`time_to_first_token_seconds_count`),
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

	// cache_hit_rate is a windowed ratio, not hits/queries counters; fake the
	// denominator so the UI's Hits/Queries*100 formula shows it directly.
	if s.PrefixCacheHits > 0 {
		s.PrefixCacheQueries = 1
	}

	return s, nil
}
