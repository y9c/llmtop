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
	total := 0.0
	for _, m := range gaugeRe(name).FindAllStringSubmatch(body, -1) {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			total += v
		}
	}
	return total
}

// seriesRe builds a regexp matching a metric pattern that already carries
// its label filter (e.g. prefill_effective_tokens_total{...mode="input"...}).
func seriesRe(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`(?:sglang:)?` + pattern + `\s+([\d.eE+-]+)`)
}

// sumSeries totals every sample matched by a fully-formed pattern. Used when
// a metric's series must be filtered by a particular label value (e.g. only
// prefill_effective_tokens_total{mode="input"}).
func sumSeries(pattern string, body string) float64 {
	total := 0.0
	for _, m := range seriesRe(pattern).FindAllStringSubmatch(body, -1) {
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
		// Cumulative requests processed, split across is_streaming series; sum all.
		{"num_requests_total", gaugeRe(`num_requests_total`),
			func(s *metrics.Snapshot, v float64) {
				s.NumRequestsTotal = sumGauge(`num_requests_total`, body)
			}},
		// token_usage is a 0-1 fraction; app.go pushes KVCacheUsagePct*100 into the KV chart
		{"token_usage", gaugeRe(`token_usage`),
			func(s *metrics.Snapshot, v float64) { s.KVCacheUsagePct = v }},
		// cache_hit_rate is a windowed 0-1 ratio (drops to 0 when idle). Kept as
		// a fallback for servers that only export it; the cumulative counter path
		// below supplies the preferred, stable CacheHitRate.
		{"cache_hit_rate", gaugeRe(`cache_hit_rate`),
			func(s *metrics.Snapshot, v float64) { s.PrefixCacheHits = v }},
		// Cumulative input tokens (mode=input) — basis for the stable hit rate.
		{"prefill_effective_tokens_total", seriesRe(`prefill_effective_tokens_total\{[^}]*mode="input"[^}]*\}`),
			func(s *metrics.Snapshot, v float64) {
				s.PrefillEffectiveTotal = sumSeries(`prefill_effective_tokens_total\{[^}]*mode="input"[^}]*\}`, body)
			}},
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
			func(s *metrics.Snapshot, v float64) { s.TTFTTotalS = sumGauge(`time_to_first_token_seconds_sum`, body) }},
		{"time_to_first_token_seconds_count", gaugeRe(`time_to_first_token_seconds_count`),
			func(s *metrics.Snapshot, v float64) {
				s.TTFTCount = sumGauge(`time_to_first_token_seconds_count`, body)
			}},
		{"inter_token_latency_seconds_sum", gaugeRe(`inter_token_latency_seconds_sum`),
			func(s *metrics.Snapshot, v float64) { s.TPOTTotalS = sumGauge(`inter_token_latency_seconds_sum`, body) }},
		{"inter_token_latency_seconds_count", gaugeRe(`inter_token_latency_seconds_count`),
			func(s *metrics.Snapshot, v float64) {
				s.TPOTCount = sumGauge(`inter_token_latency_seconds_count`, body)
			}},
		// Live decode throughput windowed by the scheduler (single series:
		// no per-rank repetition). Preferred over the completion-impulsed
		// counter delta so the displayed decode rate is honest.
		{"gen_throughput", gaugeRe(`gen_throughput`),
			func(s *metrics.Snapshot, v float64) { s.GenTokS = v }},
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

	// Stable prefix-cache hit rate from cumulative counters. device/host/storage
	// hit tokens are the re-used tokens; PrefillEffectiveTotal(mode=input) is
	// the processed (non-hit) input. hit% = hit / (hit + input). Unlike the
	// windowed cache_hit_rate this never drops to 0 just because the scheduler
	// window is idle.
	hitIn := sumSeries(`prefill_effective_tokens_total\{[^}]*mode="device_hit"[^}]*\}`, body)
	hitHost := sumSeries(`prefill_effective_tokens_total\{[^}]*mode="host_hit"[^}]*\}`, body)
	hitStore := sumSeries(`prefill_effective_tokens_total\{[^}]*mode="storage_hit"[^}]*\}`, body)
	hitTotal := hitIn + hitHost + hitStore
	if total := hitTotal + s.PrefillEffectiveTotal; total > 0 {
		s.CacheHitRate = hitTotal / total * 100
	} else if s.PrefixCacheHits > 0 {
		// Old server exports only the windowed cache_hit_rate: fake a single
		// query so the UI's hits/queries formula shows it as a percentage.
		s.PrefixCacheQueries = 1
	}

	return s, nil
}
